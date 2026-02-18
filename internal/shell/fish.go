package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// FishWriter handles fish shell history writing
type FishWriter struct {
	historyFile      string
	includeTimestamp bool
}

// NewFishWriter creates a new FishWriter.
// When includeTimestamp is true, entries include the "when:" timestamp field.
// When false, only the "- cmd:" line is written.
func NewFishWriter(historyFile string, includeTimestamp bool) *FishWriter {
	return &FishWriter{
		historyFile:      historyFile,
		includeTimestamp: includeTimestamp,
	}
}

// Append appends an entry to fish history with file locking.
// Fish history format:
//
//	- cmd: <command>
//	  when: <unix_timestamp>
func (w *FishWriter) Append(entry string) error {
	// Format entry in fish history format
	var formattedEntry string
	if w.includeTimestamp {
		formattedEntry = fmt.Sprintf("- cmd: %s\n  when: %d\n", entry, time.Now().Unix())
	} else {
		formattedEntry = fmt.Sprintf("- cmd: %s\n", entry)
	}

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
	if err := lockFile(f); err != nil {
		return err
	}
	defer func() {
		_ = unlockFile(f) // Explicitly ignore unlock error
	}()

	// Write formatted entry (already includes trailing newline)
	_, err = f.WriteString(formattedEntry)
	return err
}
