package breezing

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// WorktreeManager は git worktree の作成・クリーンアップを管理する。
// CC の WorktreeCreate/Remove フックと連携して worktree ライフサイクルを追跡する。
type WorktreeManager struct {
	mu sync.Mutex
	// projectRoot はメインリポジトリのルートパス。
	projectRoot string
	// worktrees は現在管理中の worktree 一覧（path → info）。
	worktrees map[string]*WorktreeInfo
	// staleTimeout は worktree が stale とみなされるまでの期間。
	staleTimeout time.Duration
	// now はテスト差し替え用の時刻関数。
	now func() time.Time
}

// WorktreeInfo は個々の worktree の追跡情報。
type WorktreeInfo struct {
	// Path は worktree のファイルシステムパス。
	Path string
	// Branch は worktree のブランチ名。
	Branch string
	// TaskID は worktree に割り当てられたタスク ID。
	TaskID string
	// AgentID は worktree を使用する CC エージェントの ID。
	AgentID string
	// CreatedAt は作成時刻。
	CreatedAt time.Time
	// Active は worktree が使用中かどうか。
	Active bool
}

// NewWorktreeManager は新しい WorktreeManager を返す。
func NewWorktreeManager(projectRoot string) *WorktreeManager {
	return &WorktreeManager{
		projectRoot:  projectRoot,
		worktrees:    make(map[string]*WorktreeInfo),
		staleTimeout: 24 * time.Hour,
		now:          time.Now,
	}
}

// Create は新しい worktree を作成する。
// branchName が空の場合は taskID から自動生成する。
// 作成した worktree のパスを返す。
func (wm *WorktreeManager) Create(taskID, branchName string) (string, error) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if branchName == "" {
		branchName = fmt.Sprintf("harness/worker/%s", sanitizeBranch(taskID))
	}

	// worktree パスを決定
	worktreeDir := filepath.Join(wm.projectRoot, ".harness-worktrees", sanitizeBranch(taskID))

	// 既に存在する場合は再利用
	if info, exists := wm.worktrees[worktreeDir]; exists && info.Active {
		return worktreeDir, nil
	}

	// git worktree add
	if err := wm.gitWorktreeAdd(worktreeDir, branchName); err != nil {
		return "", fmt.Errorf("worktree create: %w", err)
	}

	wm.worktrees[worktreeDir] = &WorktreeInfo{
		Path:      worktreeDir,
		Branch:    branchName,
		TaskID:    taskID,
		CreatedAt: wm.now(),
		Active:    true,
	}

	return worktreeDir, nil
}

// Remove は worktree を削除する。
// force が true の場合、未コミットの変更があっても削除する。
func (wm *WorktreeManager) Remove(worktreePath string, force bool) error {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	args := []string{"worktree", "remove", worktreePath}
	if force {
		args = append(args, "--force")
	}

	cmd := exec.Command("git", args...)
	cmd.Dir = wm.projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("worktree remove: %s: %w", strings.TrimSpace(string(out)), err)
	}

	// ブランチ削除
	if info, exists := wm.worktrees[worktreePath]; exists {
		wm.deleteBranch(info.Branch)
		delete(wm.worktrees, worktreePath)
	}

	return nil
}

// AssignAgent は worktree に CC エージェント ID を関連付ける。
func (wm *WorktreeManager) AssignAgent(worktreePath, agentID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if info, exists := wm.worktrees[worktreePath]; exists {
		info.AgentID = agentID
	}
}

// CleanupStale は staleTimeout を超過した非アクティブな worktree を削除する。
// 削除された worktree のパス一覧を返す。
func (wm *WorktreeManager) CleanupStale() []string {
	wm.mu.Lock()
	stale := make([]string, 0)
	now := wm.now()
	for path, info := range wm.worktrees {
		if !info.Active && now.Sub(info.CreatedAt) > wm.staleTimeout {
			stale = append(stale, path)
		}
	}
	wm.mu.Unlock()

	var cleaned []string
	for _, path := range stale {
		if err := wm.Remove(path, true); err != nil {
			fmt.Fprintf(os.Stderr, "worktree cleanup: %s: %v\n", path, err)
		} else {
			cleaned = append(cleaned, path)
		}
	}
	return cleaned
}

// MarkInactive は worktree を非アクティブにマークする（エージェント停止時）。
func (wm *WorktreeManager) MarkInactive(worktreePath string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	if info, exists := wm.worktrees[worktreePath]; exists {
		info.Active = false
	}
}

// List は全管理中 worktree の情報を返す。
func (wm *WorktreeManager) List() []*WorktreeInfo {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	result := make([]*WorktreeInfo, 0, len(wm.worktrees))
	for _, info := range wm.worktrees {
		cp := *info
		result = append(result, &cp)
	}
	return result
}

// HandleWorktreeCreate は CC の WorktreeCreate フックイベントを処理する。
// stdin の tool_input からパス情報を取り出して追跡を開始する。
func (wm *WorktreeManager) HandleWorktreeCreate(worktreePath, branch, agentID string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	wm.worktrees[worktreePath] = &WorktreeInfo{
		Path:      worktreePath,
		Branch:    branch,
		AgentID:   agentID,
		CreatedAt: wm.now(),
		Active:    true,
	}
}

// HandleWorktreeRemove は CC の WorktreeRemove フックイベントを処理する。
func (wm *WorktreeManager) HandleWorktreeRemove(worktreePath string) {
	wm.mu.Lock()
	defer wm.mu.Unlock()
	delete(wm.worktrees, worktreePath)
}

// gitWorktreeAdd は git worktree add を実行する。
func (wm *WorktreeManager) gitWorktreeAdd(worktreeDir, branchName string) error {
	// ディレクトリが既に存在する場合は削除
	if _, err := os.Stat(worktreeDir); err == nil {
		rmCmd := exec.Command("git", "worktree", "remove", worktreeDir, "--force")
		rmCmd.Dir = wm.projectRoot
		_ = rmCmd.Run()
	}

	cmd := exec.Command("git", "worktree", "add", "-b", branchName, worktreeDir, "HEAD")
	cmd.Dir = wm.projectRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// deleteBranch は worktree 用に作成したブランチを削除する。
func (wm *WorktreeManager) deleteBranch(branch string) {
	if branch == "" {
		return
	}
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = wm.projectRoot
	_ = cmd.Run()
}

// sanitizeBranch はタスク ID をブランチ名に使えるようサニタイズする。
func sanitizeBranch(s string) string {
	r := strings.NewReplacer(
		" ", "-",
		"/", "-",
		".", "-",
		":", "-",
	)
	return r.Replace(s)
}
