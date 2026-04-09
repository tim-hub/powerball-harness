//go:build windows

package hookhandler

import (
	"os"
)

// isProcessAlive は指定 PID のプロセスが生存しているかを確認する（Windows 実装）。
// Windows では syscall.Kill が存在しないため、os.FindProcess + Signal(0) で判定する。
// os.FindProcess は Windows では常に成功するため、Signal(os.Interrupt) の送信で確認する。
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// Windows では Signal(os.Interrupt) を送信してプロセス存在を確認する。
	// プロセスが存在しない場合はエラーが返る。
	err = proc.Signal(os.Interrupt)
	// os.Interrupt 送信が成功（またはアクセス拒否エラー）なら生存していると判断する。
	// nil エラーか "access denied" 系はプロセスが存在することを示す。
	return err == nil || isWindowsAccessDenied(err)
}

// isWindowsAccessDenied は Windows の "Access is denied" エラーを判定する。
// このエラーはプロセスが存在するが権限がないことを意味するため、プロセス生存と判断する。
func isWindowsAccessDenied(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return msg == "Access is denied." || msg == "access denied"
}
