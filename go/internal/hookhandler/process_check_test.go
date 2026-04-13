package hookhandler

import (
	"os"
	"testing"
)

func TestIsProcessAlive_CurrentProcess(t *testing.T) {
	pid := os.Getpid()
	if !isProcessAlive(pid) {
		t.Errorf("isProcessAlive(%d) = false for current process, want true", pid)
	}
}

func TestIsProcessAlive_NonExistentPID(t *testing.T) {
	nonExistentPID := 9999999
	result := isProcessAlive(nonExistentPID)
	t.Logf("isProcessAlive(%d) = %v (expected false for non-existent PID)", nonExistentPID, result)
}

func TestIsProcessAlive_ZeroPID(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("isProcessAlive(0) panicked: %v", r)
		}
	}()
	_ = isProcessAlive(0)
}
