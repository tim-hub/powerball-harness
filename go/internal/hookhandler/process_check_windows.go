//go:build windows

package hookhandler

// isProcessAlive は指定 PID のプロセスが生存しているかを確認する（Windows 実装）。
//
// Windows では Unix の syscall.Kill(pid, 0) に相当するシグナル 0 送信が
// サポートされていない。os.Interrupt を実際に送信するとプロセスを終了させてしまうため禁止。
//
// 安全側に倒す方針: Windows では processing フラグを削除しない（= alive とみなす）。
// これは bash 版の `kill -0` が Git Bash でも動作しないことと同等の動作であり、
// フラグが残っても次回セッション開始時に上書きされるため実害はない。
func isProcessAlive(_ int) bool {
	return true
}
