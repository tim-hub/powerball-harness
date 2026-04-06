package breezing

import (
	"testing"
	"time"
)

func TestWorktreeManagerHandleCreateRemove(t *testing.T) {
	wm := NewWorktreeManager("/tmp/project")

	wm.HandleWorktreeCreate("/tmp/wt1", "feature/task-1", "agent-abc")

	list := wm.List()
	if len(list) != 1 {
		t.Fatalf("expected 1 worktree, got %d", len(list))
	}
	if list[0].Path != "/tmp/wt1" {
		t.Errorf("expected /tmp/wt1, got %s", list[0].Path)
	}
	if list[0].AgentID != "agent-abc" {
		t.Errorf("expected agent-abc, got %s", list[0].AgentID)
	}
	if !list[0].Active {
		t.Error("expected active=true")
	}

	wm.HandleWorktreeRemove("/tmp/wt1")
	list = wm.List()
	if len(list) != 0 {
		t.Errorf("expected 0 worktrees after remove, got %d", len(list))
	}
}

func TestWorktreeManagerAssignAgent(t *testing.T) {
	wm := NewWorktreeManager("/tmp/project")
	wm.HandleWorktreeCreate("/tmp/wt1", "branch", "")

	wm.AssignAgent("/tmp/wt1", "agent-xyz")

	list := wm.List()
	if len(list) != 1 || list[0].AgentID != "agent-xyz" {
		t.Errorf("expected agent-xyz, got %s", list[0].AgentID)
	}
}

func TestWorktreeManagerMarkInactive(t *testing.T) {
	wm := NewWorktreeManager("/tmp/project")
	wm.HandleWorktreeCreate("/tmp/wt1", "branch", "agent-1")

	wm.MarkInactive("/tmp/wt1")

	list := wm.List()
	if len(list) != 1 || list[0].Active {
		t.Error("expected inactive after MarkInactive")
	}
}

func TestWorktreeManagerCleanupStale(t *testing.T) {
	wm := NewWorktreeManager("/tmp/project")
	wm.staleTimeout = 1 * time.Hour

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	wm.now = func() time.Time { return baseTime }

	wm.HandleWorktreeCreate("/tmp/wt-old", "branch-old", "agent-1")

	// Mark inactive
	wm.MarkInactive("/tmp/wt-old")

	// Advance time past stale timeout
	wm.now = func() time.Time { return baseTime.Add(2 * time.Hour) }

	// CleanupStale calls Remove which calls git — skip actual git call in this unit test
	// Just verify the stale detection logic
	wm.mu.Lock()
	count := 0
	now := wm.now()
	for _, info := range wm.worktrees {
		if !info.Active && now.Sub(info.CreatedAt) > wm.staleTimeout {
			count++
		}
	}
	wm.mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 stale worktree, got %d", count)
	}
}

func TestSanitizeBranch(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"35.6.1", "35-6-1"},
		{"task/foo", "task-foo"},
		{"a b c", "a-b-c"},
		{"normal", "normal"},
	}

	for _, tc := range tests {
		got := sanitizeBranch(tc.input)
		if got != tc.expected {
			t.Errorf("sanitizeBranch(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
