//go:build windows

package hookhandler

// isProcessAlive checks whether the process with the given PID is alive (Windows implementation).
//
// Windows does not support signal 0 (the equivalent of syscall.Kill(pid, 0) on Unix).
// Sending os.Interrupt would actually terminate the process, so it is prohibited here.
//
// Fail-safe policy: on Windows, the processing flag is never removed (i.e., always treated as alive).
// This matches the behavior of `kill -0` not working in Git Bash,
// and leaving the flag behind is harmless because it is overwritten at the next session start.
func isProcessAlive(_ int) bool {
	return true
}
