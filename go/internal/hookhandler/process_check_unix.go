//go:build !windows

package hookhandler

import "syscall"

// isProcessAlive は指定 PID のプロセスが生存しているかを確認する（Unix 実装）。
// kill -0 相当の操作（シグナル 0 送信）でプロセス存在を確認する。
// プロセスが存在しシグナル送信可能な場合に true を返す。
func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
