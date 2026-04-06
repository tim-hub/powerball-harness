// Package breezing は Breezing モードの goroutine オーケストレーションを提供する。
//
// Lead セッションが Worker/Reviewer を並列 spawn する際の裏側インフラ:
//   - 最大並列数制御（semaphore パターン）
//   - context.Context による graceful shutdown
//   - タスク依存関係の自動解決
//   - worktree ライフサイクル管理
package breezing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Chachamaru127/claude-code-harness/go/internal/lifecycle"
	"github.com/Chachamaru127/claude-code-harness/go/internal/state"
)

// TaskStatus はオーケストレーター内でのタスク進捗を表す。
type TaskStatus string

const (
	TaskPending   TaskStatus = "pending"
	TaskRunning   TaskStatus = "running"
	TaskCompleted TaskStatus = "completed"
	TaskFailed    TaskStatus = "failed"
	TaskBlocked   TaskStatus = "blocked"
)

// Task はオーケストレーターが管理する単一タスクを表す。
type Task struct {
	// ID はタスク識別子（例: "35.6.1"）。
	ID string
	// Description はタスクの概要。
	Description string
	// DependsOn は依存するタスク ID のリスト。
	DependsOn []string
	// AgentType は spawn するエージェント種別（"worker" / "reviewer"）。
	AgentType string
	// WorktreePath は worktree のパス（worktree 使用時）。
	WorktreePath string
}

// TaskResult はタスク完了時の結果を表す。
type TaskResult struct {
	// TaskID は完了したタスクの ID。
	TaskID string
	// AgentID は実行した CC エージェントの ID。
	AgentID string
	// CommitHash はコミットのハッシュ（成功時）。
	CommitHash string
	// Err はエラー（失敗時）。
	Err error
	// Duration は実行時間。
	Duration time.Duration
}

// WorkerFunc はオーケストレーターが各タスクに対して呼び出すコールバック。
// context がキャンセルされた場合は速やかに終了すること。
type WorkerFunc func(ctx context.Context, task *Task) TaskResult

// ProgressFunc はタスク完了時に呼び出される進捗コールバック。
type ProgressFunc func(completed, total int, result TaskResult)

// Orchestrator は Worker/Reviewer の goroutine 並列実行を管理する。
// semaphore パターンで最大並列数を制御し、context.Context で graceful shutdown を実現する。
type Orchestrator struct {
	mu sync.Mutex

	// tasks は管理対象の全タスク。
	tasks []*Task
	// status は各タスクの現在の状態。
	status map[string]TaskStatus
	// results は完了したタスクの結果。
	results map[string]TaskResult

	// maxParallel は最大並列実行数。
	maxParallel int
	// tracker はエージェントライフサイクル追跡。
	tracker *lifecycle.AgentTracker
	// store は SQLite 永続化ストア（nil 許容）。
	store *state.HarnessStore

	// workerFn は各タスクの実行関数。
	workerFn WorkerFunc
	// progressFn は進捗コールバック（nil 許容）。
	progressFn ProgressFunc
}

// OrchestratorOption は Orchestrator の設定オプション。
type OrchestratorOption func(*Orchestrator)

// WithMaxParallel は最大並列数を設定する。
// 0 以下の場合はデフォルト (3) が使用される。
func WithMaxParallel(n int) OrchestratorOption {
	return func(o *Orchestrator) {
		if n > 0 {
			o.maxParallel = n
		}
	}
}

// WithTracker は AgentTracker を設定する。
func WithTracker(t *lifecycle.AgentTracker) OrchestratorOption {
	return func(o *Orchestrator) {
		o.tracker = t
	}
}

// WithStore は HarnessStore を設定する。
func WithStore(s *state.HarnessStore) OrchestratorOption {
	return func(o *Orchestrator) {
		o.store = s
	}
}

// WithProgressFunc は進捗コールバックを設定する。
func WithProgressFunc(fn ProgressFunc) OrchestratorOption {
	return func(o *Orchestrator) {
		o.progressFn = fn
	}
}

// NewOrchestrator は新しい Orchestrator を生成する。
// workerFn は各タスクの実行関数（必須）。
func NewOrchestrator(workerFn WorkerFunc, opts ...OrchestratorOption) *Orchestrator {
	o := &Orchestrator{
		status:      make(map[string]TaskStatus),
		results:     make(map[string]TaskResult),
		maxParallel: 3,
		workerFn:    workerFn,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// AddTask はオーケストレーターにタスクを追加する。
// Run 呼び出し前に全タスクを追加すること。
func (o *Orchestrator) AddTask(task *Task) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.tasks = append(o.tasks, task)
	o.status[task.ID] = TaskPending
}

// Run はタスクを依存関係に従って並列実行する。
// 全タスク完了（成功 or 失敗）まで待つ。context キャンセルで graceful shutdown。
// 返される []TaskResult は完了順。
func (o *Orchestrator) Run(ctx context.Context) ([]TaskResult, error) {
	o.mu.Lock()
	total := len(o.tasks)
	if total == 0 {
		o.mu.Unlock()
		return nil, nil
	}
	o.mu.Unlock()

	// semaphore: 最大並列数を制御するバッファ付きチャネル
	sem := make(chan struct{}, o.maxParallel)
	// resultCh: 完了通知を集約
	resultCh := make(chan TaskResult, total)
	// wg: 全 goroutine の完了待ち
	var wg sync.WaitGroup

	// 依存が解決されたタスクを定期的にチェックして dispatch する
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// dispatch ループ
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				o.dispatchReady(ctx, sem, resultCh, &wg)

				// 全タスクが完了 or 失敗したら終了
				o.mu.Lock()
				allDone := o.allTerminated()
				o.mu.Unlock()
				if allDone {
					return
				}
			}
		}
	}()

	// 結果集約ループ
	var results []TaskResult
	completed := 0
	for completed < total {
		select {
		case result := <-resultCh:
			results = append(results, result)
			completed++
			if o.progressFn != nil {
				o.progressFn(completed, total, result)
			}
		case <-ctx.Done():
			// キャンセル: 残り全タスクの結果を待つ
			wg.Wait()
			// drain remaining results
			for len(resultCh) > 0 {
				result := <-resultCh
				results = append(results, result)
			}
			return results, ctx.Err()
		}
	}

	wg.Wait()
	return results, nil
}

// Status は指定タスクの現在のステータスを返す。
func (o *Orchestrator) Status(taskID string) TaskStatus {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.status[taskID]
}

// Results は全完了タスクの結果を返す。
func (o *Orchestrator) Results() map[string]TaskResult {
	o.mu.Lock()
	defer o.mu.Unlock()
	res := make(map[string]TaskResult, len(o.results))
	for k, v := range o.results {
		res[k] = v
	}
	return res
}

// dispatchReady は依存が解決された pending タスクを goroutine で起動する。
func (o *Orchestrator) dispatchReady(ctx context.Context, sem chan struct{}, resultCh chan<- TaskResult, wg *sync.WaitGroup) {
	o.mu.Lock()
	defer o.mu.Unlock()

	for _, task := range o.tasks {
		if o.status[task.ID] != TaskPending {
			continue
		}

		// 依存先が失敗していれば blocked に（depsResolved より先にチェック）
		if o.depsFailed(task) {
			o.status[task.ID] = TaskBlocked
			resultCh <- TaskResult{
				TaskID: task.ID,
				Err:    fmt.Errorf("blocked: dependency failed"),
			}
			continue
		}

		if !o.depsResolved(task) {
			continue
		}

		// semaphore 取得を試みる（ノンブロッキング）
		select {
		case sem <- struct{}{}:
		default:
			continue // 並列上限に達している
		}

		o.status[task.ID] = TaskRunning
		wg.Add(1)

		go func(t *Task) {
			defer wg.Done()
			defer func() { <-sem }() // semaphore 解放

			start := time.Now()
			result := o.workerFn(ctx, t)
			result.Duration = time.Since(start)
			result.TaskID = t.ID

			o.mu.Lock()
			if result.Err != nil {
				o.status[t.ID] = TaskFailed
			} else {
				o.status[t.ID] = TaskCompleted
			}
			o.results[t.ID] = result
			o.mu.Unlock()

			resultCh <- result
		}(task)
	}
}

// depsResolved は全依存タスクが completed かどうかを返す。
// ロック取得済み前提。
func (o *Orchestrator) depsResolved(task *Task) bool {
	for _, depID := range task.DependsOn {
		if o.status[depID] != TaskCompleted {
			return false
		}
	}
	return true
}

// depsFailed は依存タスクのいずれかが failed/blocked かどうかを返す。
// ロック取得済み前提。
func (o *Orchestrator) depsFailed(task *Task) bool {
	for _, depID := range task.DependsOn {
		s := o.status[depID]
		if s == TaskFailed || s == TaskBlocked {
			return true
		}
	}
	return false
}

// allTerminated は全タスクが終端状態 (completed/failed/blocked) にあるかを返す。
// ロック取得済み前提。
func (o *Orchestrator) allTerminated() bool {
	for _, task := range o.tasks {
		s := o.status[task.ID]
		if s != TaskCompleted && s != TaskFailed && s != TaskBlocked {
			return false
		}
	}
	return true
}
