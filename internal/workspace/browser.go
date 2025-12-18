package workspace

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// DirEntry represents a file or directory entry in the browser
type DirEntry struct {
	Name       string    // Entry name (not full path)
	Path       string    // Absolute path
	IsDir      bool      // True if directory
	IsMakefile bool      // True if this is a Makefile
	Size       int64     // File size in bytes
}

// BrowserState manages the file browser navigation state
type BrowserState struct {
	CurrentDir  string     // Absolute path to current directory
	Entries     []DirEntry // Entries in current directory
	SelectedIdx int        // Currently selected entry index
	Error       error      // Last error encountered
}

// NewBrowser creates a new file browser starting at the given directory
func NewBrowser(startDir string) (*BrowserState, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	// Check if directory exists
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		// If it's a file, use its parent directory
		absPath = filepath.Dir(absPath)
	}

	browser := &BrowserState{
		CurrentDir:  absPath,
		SelectedIdx: 0,
	}

	// Load initial entries
	if err := browser.RefreshEntries(); err != nil {
		return nil, err
	}

	return browser, nil
}

// RefreshEntries reloads the directory entries
func (b *BrowserState) RefreshEntries() error {
	entries, err := os.ReadDir(b.CurrentDir)
	if err != nil {
		b.Error = err
		return err
	}

	b.Entries = make([]DirEntry, 0, len(entries)+1)

	// Add parent directory entry if not at root
	if b.CurrentDir != "/" && b.CurrentDir != filepath.VolumeName(b.CurrentDir)+"/" {
		b.Entries = append(b.Entries, DirEntry{
			Name:  "..",
			Path:  filepath.Dir(b.CurrentDir),
			IsDir: true,
		})
	}

	// Convert to DirEntry and sort
	for _, entry := range entries {
		// Skip hidden files/directories (start with .)
		if strings.HasPrefix(entry.Name(), ".") && entry.Name() != ".." {
			continue
		}

		fullPath := filepath.Join(b.CurrentDir, entry.Name())
		isDir := entry.IsDir()
		isMakefile := false
		var size int64

		if !isDir {
			// Check if it's a Makefile
			isMakefile = IsMakefile(entry.Name())

			// Get file size
			if info, err := entry.Info(); err == nil {
				size = info.Size()
			}
		}

		b.Entries = append(b.Entries, DirEntry{
			Name:       entry.Name(),
			Path:       fullPath,
			IsDir:      isDir,
			IsMakefile: isMakefile,
			Size:       size,
		})
	}

	// Sort: directories first, then files, alphabetically
	sort.Slice(b.Entries, func(i, j int) bool {
		// ".." always first
		if b.Entries[i].Name == ".." {
			return true
		}
		if b.Entries[j].Name == ".." {
			return false
		}

		// Directories before files
		if b.Entries[i].IsDir != b.Entries[j].IsDir {
			return b.Entries[i].IsDir
		}

		// Alphabetical
		return strings.ToLower(b.Entries[i].Name) < strings.ToLower(b.Entries[j].Name)
	})

	// Keep selection in bounds
	if b.SelectedIdx >= len(b.Entries) {
		b.SelectedIdx = len(b.Entries) - 1
	}
	if b.SelectedIdx < 0 {
		b.SelectedIdx = 0
	}

	b.Error = nil
	return nil
}

// NavigateUp moves to the parent directory
func (b *BrowserState) NavigateUp() error {
	parent := filepath.Dir(b.CurrentDir)
	if parent == b.CurrentDir {
		// Already at root
		return nil
	}

	b.CurrentDir = parent
	b.SelectedIdx = 0
	return b.RefreshEntries()
}

// NavigateInto enters the selected directory or selects a Makefile
func (b *BrowserState) NavigateInto() error {
	if len(b.Entries) == 0 {
		return nil
	}

	selected := b.Entries[b.SelectedIdx]

	if selected.IsDir {
		// Enter directory
		b.CurrentDir = selected.Path
		b.SelectedIdx = 0
		return b.RefreshEntries()
	}

	// For files, do nothing (caller will handle Makefile selection)
	return nil
}

// GetCurrentSelection returns the currently selected entry
func (b *BrowserState) GetCurrentSelection() *DirEntry {
	if len(b.Entries) == 0 || b.SelectedIdx < 0 || b.SelectedIdx >= len(b.Entries) {
		return nil
	}
	return &b.Entries[b.SelectedIdx]
}

// MoveUp moves selection up one entry
func (b *BrowserState) MoveUp() {
	if b.SelectedIdx > 0 {
		b.SelectedIdx--
	}
}

// MoveDown moves selection down one entry
func (b *BrowserState) MoveDown() {
	if b.SelectedIdx < len(b.Entries)-1 {
		b.SelectedIdx++
	}
}

// GetBreadcrumb returns a formatted breadcrumb path
func (b *BrowserState) GetBreadcrumb() string {
	// Shorten home directory to ~
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(b.CurrentDir, home) {
		return "~" + strings.TrimPrefix(b.CurrentDir, home)
	}
	return b.CurrentDir
}

// CountMakefiles returns the number of Makefiles in the current directory
func (b *BrowserState) CountMakefiles() int {
	count := 0
	for _, entry := range b.Entries {
		if entry.IsMakefile {
			count++
		}
	}
	return count
}
