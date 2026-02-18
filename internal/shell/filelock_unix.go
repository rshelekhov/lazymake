//go:build unix

package shell

import (
	"os"
	"syscall"
)

// lockFile acquires an exclusive lock on the file
func lockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_EX) //nolint:gosec // fd fits in int on supported platforms
}

// unlockFile releases the lock on the file
func unlockFile(f *os.File) error {
	return syscall.Flock(int(f.Fd()), syscall.LOCK_UN) //nolint:gosec // fd fits in int on supported platforms
}
