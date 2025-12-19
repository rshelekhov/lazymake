package shell

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// ZshWriter handles zsh history writing
type ZshWriter struct {
	historyFile     string
	extendedHistory bool
}

// NewZshWriter creates a new ZshWriter and detects extended history format
func NewZshWriter(historyFile string) *ZshWriter {
	// Detect if extended history is enabled
	extended := detectZshExtendedHistory(historyFile)

	return &ZshWriter{
		historyFile:     historyFile,
		extendedHistory: extended,
	}
}

// detectZshExtendedHistory checks if the history file uses extended history format
func detectZshExtendedHistory(path string) bool {
	// Read first 5 lines to check format
	f, err := os.Open(path)
	if err != nil {
		return false // Default to standard format if file doesn't exist
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for i := 0; i < 5 && scanner.Scan(); i++ {
		line := scanner.Text()
		// Extended history starts with ": <timestamp>:"
		if strings.HasPrefix(line, ": ") && strings.Contains(line, ":0;") {
			return true
		}
	}

	return false
}

// Append appends an entry to zsh history with file locking
func (w *ZshWriter) Append(entry string) error {
	var formattedEntry string

	if w.extendedHistory {
		// Extended history format: : <timestamp>:0;command
		timestamp := time.Now().Unix()
		formattedEntry = fmt.Sprintf(": %d:0;%s", timestamp, entry)
	} else {
		// Standard format
		formattedEntry = entry
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
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		return err
	}
	defer func() {
		_ = syscall.Flock(int(f.Fd()), syscall.LOCK_UN) // Explicitly ignore unlock error
	}()

	// Write entry with newline
	_, err = f.WriteString(formattedEntry + "\n")
	return err
}
