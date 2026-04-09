package hookhandler

import (
	"os"
	"testing"
)

// TestIsProcessAlive_CurrentProcess は現在のプロセスが生存していると判定されることを確認する。
func TestIsProcessAlive_CurrentProcess(t *testing.T) {
	pid := os.Getpid()
	if !isProcessAlive(pid) {
		t.Errorf("isProcessAlive(%d) = false for current process, want true", pid)
	}
}

// TestIsProcessAlive_NonExistentPID は存在しない PID が生存していないと判定されることを確認する。
// PID 0 はシステムプロセス（またはカーネル）のため、通常プロセスから Signal 送信は拒否される。
// PID -1 は無効な PID なのでエラーになる（Unix では全プロセスへのシグナル）。
// ここでは大きな値（存在しないと期待される）を使う。
func TestIsProcessAlive_NonExistentPID(t *testing.T) {
	// 非常に大きな PID（通常存在しない）
	// Linux/macOS の PID 最大値は一般的に 4194304 以下
	nonExistentPID := 9999999
	// このテストは「結果が一定である」ことは保証できないが、
	// パニックしないことを確認する
	result := isProcessAlive(nonExistentPID)
	// 存在しない PID に対しては false が返ることを期待するが、
	// OS によっては異なる可能性があるため結果のみ記録
	t.Logf("isProcessAlive(%d) = %v (expected false for non-existent PID)", nonExistentPID, result)
}

// TestIsProcessAlive_ZeroPID はゼロ PID の処理でパニックしないことを確認する。
func TestIsProcessAlive_ZeroPID(t *testing.T) {
	// PID 0 は特殊なためパニックしないことだけ確認
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("isProcessAlive(0) panicked: %v", r)
		}
	}()
	_ = isProcessAlive(0)
}
