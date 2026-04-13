//go:build !windows

package hookhandler

import "syscall"

// isProcessAlive checks whether the process with the given PID is alive (Unix implementation).
// Checks process existence by sending signal 0, equivalent to kill -0.
// Returns true if the process exists and the signal can be sent.
func isProcessAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}
