//go:build windows

package shell

import (
	"os"
	"syscall"
	"unsafe"
)

var (
	kernel32         = syscall.NewLazyDLL("kernel32.dll")
	procLockFileEx   = kernel32.NewProc("LockFileEx")
	procUnlockFileEx = kernel32.NewProc("UnlockFileEx")
)

const (
	lockfileExclusiveLock = 0x00000002
)

// lockFile acquires an exclusive lock on the file using Windows API
func lockFile(f *os.File) error {
	var overlapped syscall.Overlapped
	
	// LockFileEx parameters:
	// - hFile: file handle
	// - dwFlags: LOCKFILE_EXCLUSIVE_LOCK
	// - dwReserved: must be zero
	// - nNumberOfBytesToLockLow: lock entire file (0xFFFFFFFF)
	// - nNumberOfBytesToLockHigh: lock entire file (0xFFFFFFFF)
	// - lpOverlapped: pointer to OVERLAPPED structure
	ret, _, err := procLockFileEx.Call(
		uintptr(f.Fd()),
		uintptr(lockfileExclusiveLock),
		uintptr(0),
		uintptr(0xFFFFFFFF),
		uintptr(0xFFFFFFFF),
		uintptr(unsafe.Pointer(&overlapped)),
	)
	
	if ret == 0 {
		return err
	}
	return nil
}

// unlockFile releases the lock on the file using Windows API
func unlockFile(f *os.File) error {
	var overlapped syscall.Overlapped
	
	// UnlockFileEx parameters:
	// - hFile: file handle
	// - dwReserved: must be zero
	// - nNumberOfBytesToUnlockLow: unlock entire file (0xFFFFFFFF)
	// - nNumberOfBytesToUnlockHigh: unlock entire file (0xFFFFFFFF)
	// - lpOverlapped: pointer to OVERLAPPED structure
	ret, _, err := procUnlockFileEx.Call(
		uintptr(f.Fd()),
		uintptr(0),
		uintptr(0xFFFFFFFF),
		uintptr(0xFFFFFFFF),
		uintptr(unsafe.Pointer(&overlapped)),
	)
	
	if ret == 0 {
		return err
	}
	return nil
}
