package breezing

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// DependencyGraph はタスク間の依存関係を管理し、実行可能なタスクを判定する。
type DependencyGraph struct {
	mu    sync.RWMutex
	tasks map[string]*depNode
}

type depNode struct {
	task     *Task
	deps     []string // 依存先タスク ID
	status   TaskStatus
	resolved bool // 全依存が completed になった
}

// NewDependencyGraph は新しい DependencyGraph を生成する。
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		tasks: make(map[string]*depNode),
	}
}

// Add はタスクを依存グラフに追加する。
func (dg *DependencyGraph) Add(task *Task) {
	dg.mu.Lock()
	defer dg.mu.Unlock()
	dg.tasks[task.ID] = &depNode{
		task:   task,
		deps:   task.DependsOn,
		status: TaskPending,
	}
}

// Ready は依存が全て解決され、実行可能なタスクのリストを返す。
// 返されるタスクは ID の昇順でソートされる。
func (dg *DependencyGraph) Ready() []*Task {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	var ready []*Task
	for _, node := range dg.tasks {
		if node.status != TaskPending {
			continue
		}
		if dg.allDepsCompleted(node) {
			ready = append(ready, node.task)
		}
	}

	sort.Slice(ready, func(i, j int) bool {
		return ready[i].ID < ready[j].ID
	})
	return ready
}

// MarkCompleted はタスクを完了としてマークし、依存チェーンを再評価する。
// 返される []*Task は新たに unblock されたタスク。
func (dg *DependencyGraph) MarkCompleted(taskID string) []*Task {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if node, exists := dg.tasks[taskID]; exists {
		node.status = TaskCompleted
	}

	// 新たに ready になったタスクを検出
	var unblocked []*Task
	for _, node := range dg.tasks {
		if node.status != TaskPending {
			continue
		}
		if !node.resolved && dg.allDepsCompleted(node) {
			node.resolved = true
			unblocked = append(unblocked, node.task)
		}
	}
	return unblocked
}

// MarkFailed はタスクを失敗としてマークする。
// このタスクに依存する全てのタスクは blocked になる。
func (dg *DependencyGraph) MarkFailed(taskID string) []string {
	dg.mu.Lock()
	defer dg.mu.Unlock()

	if node, exists := dg.tasks[taskID]; exists {
		node.status = TaskFailed
	}

	// 連鎖的にブロックされるタスクを特定
	var blocked []string
	changed := true
	for changed {
		changed = false
		for id, node := range dg.tasks {
			if node.status != TaskPending {
				continue
			}
			for _, depID := range node.deps {
				depNode := dg.tasks[depID]
				if depNode != nil && (depNode.status == TaskFailed || depNode.status == TaskBlocked) {
					node.status = TaskBlocked
					blocked = append(blocked, id)
					changed = true
					break
				}
			}
		}
	}
	return blocked
}

// DetectCycle は依存グラフに循環参照がないかを検出する。
// 循環がある場合は循環に含まれるタスク ID のスライスを返す。
func (dg *DependencyGraph) DetectCycle() []string {
	dg.mu.RLock()
	defer dg.mu.RUnlock()
	return dg.detectCycleLocked()
}

// detectCycleLocked はロック取得済み前提の cycle 検出。
func (dg *DependencyGraph) detectCycleLocked() []string {
	visited := make(map[string]bool)
	inStack := make(map[string]bool)
	var cyclePath []string

	var dfs func(id string) bool
	dfs = func(id string) bool {
		visited[id] = true
		inStack[id] = true

		node, exists := dg.tasks[id]
		if !exists {
			inStack[id] = false
			return false
		}

		for _, depID := range node.deps {
			if !visited[depID] {
				if dfs(depID) {
					cyclePath = append([]string{depID}, cyclePath...)
					return true
				}
			} else if inStack[depID] {
				cyclePath = append(cyclePath, depID)
				return true
			}
		}

		inStack[id] = false
		return false
	}

	for id := range dg.tasks {
		if !visited[id] {
			if dfs(id) {
				return cyclePath
			}
		}
	}
	return nil
}

// TopologicalOrder は依存関係に基づくトポロジカルソート順を返す。
// 循環がある場合はエラーを返す。
// 単一ロック内で cycle 検出 + ソートを行い TOCTOU を防ぐ。
func (dg *DependencyGraph) TopologicalOrder() ([]string, error) {
	dg.mu.RLock()
	defer dg.mu.RUnlock()

	// inline cycle detection
	if cycle := dg.detectCycleLocked(); len(cycle) > 0 {
		return nil, fmt.Errorf("circular dependency detected: %s", strings.Join(cycle, " → "))
	}

	inDegree := make(map[string]int)
	for id := range dg.tasks {
		inDegree[id] = 0
	}
	for _, node := range dg.tasks {
		for _, depID := range node.deps {
			if _, exists := dg.tasks[depID]; exists {
				inDegree[node.task.ID]++
			}
		}
	}

	// Kahn's algorithm
	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue)

	var order []string
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		order = append(order, id)

		// この task に依存しているタスクの in-degree を減らす
		for othID, node := range dg.tasks {
			for _, depID := range node.deps {
				if depID == id {
					inDegree[othID]--
					if inDegree[othID] == 0 {
						queue = append(queue, othID)
						sort.Strings(queue) // 安定ソート
					}
				}
			}
		}
	}

	return order, nil
}

// allDepsCompleted は指定ノードの全依存が completed かを返す。
// ロック取得済み前提。
func (dg *DependencyGraph) allDepsCompleted(node *depNode) bool {
	for _, depID := range node.deps {
		dep, exists := dg.tasks[depID]
		if !exists {
			// 外部依存（グラフに存在しない）は完了済みとみなす
			continue
		}
		if dep.status != TaskCompleted {
			return false
		}
	}
	return true
}

// ============================================================
// File-Lock Claiming
// ============================================================

// FileLock はファイルベースのロックを管理する。
// Worker が特定のファイルに対する排他的な変更権を主張するために使用する。
type FileLock struct {
	mu       sync.Mutex
	lockDir  string
	locks    map[string]string // filePath → ownerID
	now      func() time.Time
}

// NewFileLock は新しい FileLock を生成する。
// lockDir は .harness-locks/ ディレクトリのパス。
func NewFileLock(lockDir string) *FileLock {
	return &FileLock{
		lockDir: lockDir,
		locks:   make(map[string]string),
		now:     time.Now,
	}
}

// Claim はファイルに対するロックを取得する。
// 既に他のオーナーがロックしている場合はエラーを返す。
func (fl *FileLock) Claim(filePath, ownerID string) error {
	filePath = filepath.Clean(filePath)
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// メモリ上で他オーナーが保持していれば即エラー
	if existing, exists := fl.locks[filePath]; exists && existing != ownerID {
		return fmt.Errorf("file %q is locked by %q", filePath, existing)
	}

	// ファイルシステムで原子的排他（O_CREATE|O_EXCL）
	lockFile := fl.lockFilePath(filePath)
	if err := os.MkdirAll(filepath.Dir(lockFile), 0o755); err != nil {
		return fmt.Errorf("create lock dir: %w", err)
	}
	content := fmt.Sprintf("%s\n%s\n", ownerID, fl.now().UTC().Format(time.RFC3339))
	f, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		if os.IsExist(err) {
			// ロックファイルが既存 — 同一オーナーの再 claim か確認
			existing, readErr := os.ReadFile(lockFile)
			if readErr == nil {
				lines := strings.SplitN(string(existing), "\n", 2)
				if len(lines) > 0 && strings.TrimSpace(lines[0]) == ownerID {
					// 同一オーナー — ロックファイルを更新して再 claim を許可
					_ = os.WriteFile(lockFile, []byte(content), 0o644)
					fl.locks[filePath] = ownerID
					return nil
				}
			}
			return fmt.Errorf("file %q is locked by another process: %s", filePath, strings.TrimSpace(string(existing)))
		}
		return fmt.Errorf("create lock file: %w", err)
	}
	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(lockFile) // ロールバック: 不完全なロックファイルを削除
		return fmt.Errorf("write lock file: %w", err)
	}
	if err := f.Close(); err != nil {
		os.Remove(lockFile)
		return fmt.Errorf("close lock file: %w", err)
	}

	// ファイルロック成功後にメモリ状態を更新（失敗時はここに到達しない）
	fl.locks[filePath] = ownerID
	return nil
}

// Release はファイルのロックを解放する。
// メモリとディスク両方でオーナーを検証し、一致しない場合はエラーを返す。
func (fl *FileLock) Release(filePath, ownerID string) error {
	filePath = filepath.Clean(filePath)
	fl.mu.Lock()
	defer fl.mu.Unlock()

	// メモリ上のオーナーチェック
	if existing, exists := fl.locks[filePath]; exists && existing != ownerID {
		return fmt.Errorf("file %q is locked by %q, not %q", filePath, existing, ownerID)
	}

	// ディスク上のロックファイルのオーナーも検証（プロセス間安全性）
	lockFile := fl.lockFilePath(filePath)
	if data, err := os.ReadFile(lockFile); err == nil {
		lines := strings.SplitN(string(data), "\n", 2)
		if len(lines) > 0 {
			diskOwner := strings.TrimSpace(lines[0])
			if diskOwner != ownerID {
				return fmt.Errorf("file %q is locked on disk by %q, not %q", filePath, diskOwner, ownerID)
			}
		}
	}

	delete(fl.locks, filePath)
	_ = os.Remove(lockFile)

	return nil
}

// ReleaseAll は指定オーナーの全ロックを解放する。
// メモリ上の所有分のみ解放し、ディスク上のオーナーも検証する。
func (fl *FileLock) ReleaseAll(ownerID string) {
	fl.mu.Lock()
	defer fl.mu.Unlock()

	for filePath, owner := range fl.locks {
		if owner == ownerID {
			lockFile := fl.lockFilePath(filePath)
			// ディスク上のオーナーも検証してから削除
			if data, err := os.ReadFile(lockFile); err == nil {
				lines := strings.SplitN(string(data), "\n", 2)
				if len(lines) > 0 && strings.TrimSpace(lines[0]) != ownerID {
					continue // ディスク上のオーナーが異なる — 削除しない
				}
			}
			delete(fl.locks, filePath)
			_ = os.Remove(lockFile)
		}
	}
}

// Owner はファイルのロックオーナーを返す。ロックされていない場合は空文字列。
func (fl *FileLock) Owner(filePath string) string {
	filePath = filepath.Clean(filePath)
	fl.mu.Lock()
	defer fl.mu.Unlock()
	return fl.locks[filePath]
}

// lockFilePath はロックファイルのパスを返す。
func (fl *FileLock) lockFilePath(filePath string) string {
	// ファイルパスをフラットなファイル名に変換
	safe := strings.NewReplacer("/", "__", "\\", "__", ":", "_").Replace(filePath)
	return filepath.Join(fl.lockDir, safe+".lock")
}
