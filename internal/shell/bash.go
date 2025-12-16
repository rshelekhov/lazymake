package shell

import (
	"os"
	"path/filepath"
	"syscall"
)

// BashWriter handles bash history writing
type BashWriter struct {
	historyFile string
}

// NewBashWriter creates a new BashWriter
func NewBashWriter(historyFile string) *BashWriter {
	return &BashWriter{
		historyFile: historyFile,
	}
}

// Append appends an entry to bash history with file locking
func (w *BashWriter) Append(entry string) error {
	// Create parent directory if it doesn't exist
	dir := filepath.Dir(w.historyFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open file with append mode, create if doesn't exist
	f, err := os.OpenFile(w.historyFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	// Acquire exclusive lock to prevent corruption
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	// Write entry with newline
	_, err = f.WriteString(entry + "\n")
	return err
}
