package breezing

import (
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// DependencyGraph tests
// ============================================================

func TestDepGraphReady(t *testing.T) {
	dg := NewDependencyGraph()
	dg.Add(&Task{ID: "A"})
	dg.Add(&Task{ID: "B", DependsOn: []string{"A"}})
	dg.Add(&Task{ID: "C"})

	ready := dg.Ready()
	if len(ready) != 2 {
		t.Fatalf("expected 2 ready tasks (A, C), got %d", len(ready))
	}
	// sorted by ID
	if ready[0].ID != "A" || ready[1].ID != "C" {
		t.Errorf("expected [A, C], got [%s, %s]", ready[0].ID, ready[1].ID)
	}
}

func TestDepGraphMarkCompleted(t *testing.T) {
	dg := NewDependencyGraph()
	dg.Add(&Task{ID: "A"})
	dg.Add(&Task{ID: "B", DependsOn: []string{"A"}})
	dg.Add(&Task{ID: "C", DependsOn: []string{"B"}})

	// Initially: only A is ready
	ready := dg.Ready()
	if len(ready) != 1 || ready[0].ID != "A" {
		t.Fatalf("expected [A] ready, got %v", taskIDs(ready))
	}

	// Complete A → B becomes ready
	unblocked := dg.MarkCompleted("A")
	if len(unblocked) != 1 || unblocked[0].ID != "B" {
		t.Fatalf("expected [B] unblocked, got %v", taskIDs(unblocked))
	}

	// Complete B → C becomes ready
	unblocked = dg.MarkCompleted("B")
	if len(unblocked) != 1 || unblocked[0].ID != "C" {
		t.Fatalf("expected [C] unblocked, got %v", taskIDs(unblocked))
	}
}

func TestDepGraphMarkFailed(t *testing.T) {
	dg := NewDependencyGraph()
	dg.Add(&Task{ID: "A"})
	dg.Add(&Task{ID: "B", DependsOn: []string{"A"}})
	dg.Add(&Task{ID: "C", DependsOn: []string{"B"}})
	dg.Add(&Task{ID: "D"}) // independent

	blocked := dg.MarkFailed("A")
	if len(blocked) != 2 {
		t.Fatalf("expected 2 blocked (B, C), got %d: %v", len(blocked), blocked)
	}

	// D should still be ready
	ready := dg.Ready()
	if len(ready) != 1 || ready[0].ID != "D" {
		t.Errorf("expected [D] ready, got %v", taskIDs(ready))
	}
}

func TestDepGraphDetectCycle(t *testing.T) {
	dg := NewDependencyGraph()
	dg.Add(&Task{ID: "A", DependsOn: []string{"C"}})
	dg.Add(&Task{ID: "B", DependsOn: []string{"A"}})
	dg.Add(&Task{ID: "C", DependsOn: []string{"B"}})

	cycle := dg.DetectCycle()
	if len(cycle) == 0 {
		t.Fatal("expected cycle to be detected")
	}
}

func TestDepGraphNoCycle(t *testing.T) {
	dg := NewDependencyGraph()
	dg.Add(&Task{ID: "A"})
	dg.Add(&Task{ID: "B", DependsOn: []string{"A"}})
	dg.Add(&Task{ID: "C", DependsOn: []string{"A"}})

	cycle := dg.DetectCycle()
	if len(cycle) != 0 {
		t.Errorf("unexpected cycle detected: %v", cycle)
	}
}

func TestDepGraphTopologicalOrder(t *testing.T) {
	dg := NewDependencyGraph()
	dg.Add(&Task{ID: "C", DependsOn: []string{"A", "B"}})
	dg.Add(&Task{ID: "A"})
	dg.Add(&Task{ID: "B"})

	order, err := dg.TopologicalOrder()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 items, got %d", len(order))
	}
	// C must come after A and B
	idxA, idxB, idxC := indexOf(order, "A"), indexOf(order, "B"), indexOf(order, "C")
	if idxA >= idxC || idxB >= idxC {
		t.Errorf("topological order violated: A=%d, B=%d, C=%d", idxA, idxB, idxC)
	}
}

func TestDepGraphTopologicalCycleError(t *testing.T) {
	dg := NewDependencyGraph()
	dg.Add(&Task{ID: "A", DependsOn: []string{"B"}})
	dg.Add(&Task{ID: "B", DependsOn: []string{"A"}})

	_, err := dg.TopologicalOrder()
	if err == nil {
		t.Fatal("expected error for circular dependency")
	}
}

// ============================================================
// FileLock tests
// ============================================================

func TestFileLockClaimRelease(t *testing.T) {
	dir := t.TempDir()
	fl := NewFileLock(dir)

	if err := fl.Claim("src/main.go", "worker-1"); err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	if fl.Owner("src/main.go") != "worker-1" {
		t.Error("expected owner worker-1")
	}

	// Lock file should exist on disk
	lockPath := filepath.Join(dir, "src__main.go.lock")
	if _, err := os.Stat(lockPath); os.IsNotExist(err) {
		t.Error("lock file should exist on disk")
	}

	if err := fl.Release("src/main.go", "worker-1"); err != nil {
		t.Fatalf("release failed: %v", err)
	}
	if fl.Owner("src/main.go") != "" {
		t.Error("expected no owner after release")
	}
}

func TestFileLockConflict(t *testing.T) {
	dir := t.TempDir()
	fl := NewFileLock(dir)

	if err := fl.Claim("file.go", "worker-1"); err != nil {
		t.Fatalf("first claim failed: %v", err)
	}

	err := fl.Claim("file.go", "worker-2")
	if err == nil {
		t.Fatal("expected conflict error")
	}

	// Same owner re-claiming is OK (idempotent)
	if err := fl.Claim("file.go", "worker-1"); err != nil {
		t.Fatalf("same owner re-claim should succeed: %v", err)
	}
}

func TestFileLockReleaseAll(t *testing.T) {
	dir := t.TempDir()
	fl := NewFileLock(dir)

	fl.Claim("a.go", "w1") //nolint:errcheck
	fl.Claim("b.go", "w1") //nolint:errcheck
	fl.Claim("c.go", "w2") //nolint:errcheck

	fl.ReleaseAll("w1")

	if fl.Owner("a.go") != "" {
		t.Error("a.go should be released")
	}
	if fl.Owner("b.go") != "" {
		t.Error("b.go should be released")
	}
	if fl.Owner("c.go") != "w2" {
		t.Error("c.go should still be owned by w2")
	}
}

func TestFileLockWrongOwnerRelease(t *testing.T) {
	dir := t.TempDir()
	fl := NewFileLock(dir)

	fl.Claim("file.go", "w1") //nolint:errcheck
	err := fl.Release("file.go", "w2")
	if err == nil {
		t.Fatal("expected error when wrong owner releases")
	}
}

// ============================================================
// Helpers
// ============================================================

func taskIDs(tasks []*Task) []string {
	ids := make([]string, len(tasks))
	for i, t := range tasks {
		ids[i] = t.ID
	}
	return ids
}

func indexOf(s []string, target string) int {
	for i, v := range s {
		if v == target {
			return i
		}
	}
	return -1
}
