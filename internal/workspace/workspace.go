package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	maxRecentWorkspaces = 10
	workspaceFileName   = "workspaces.json"
)

// Workspace represents a Makefile workspace
type Workspace struct {
	Path         string    `json:"path"`           // Absolute path to Makefile
	LastAccessed time.Time `json:"last_accessed"`  // When last accessed
	AccessCount  int       `json:"access_count"`   // Total access count
	IsFavorite   bool      `json:"is_favorite"`    // User-marked favorite
	DisplayName  string    `json:"display_name,omitempty"` // Optional custom name
}

// Manager handles workspace persistence and operations
type Manager struct {
	Workspaces []Workspace `json:"workspaces"`
	path       string      // Cache file path
}

// Load reads workspace data from the cache directory
// Returns an empty manager on error (graceful degradation)
func Load() (*Manager, error) {
	path, err := getCachePath()
	if err != nil {
		return newEmpty(), fmt.Errorf("failed to get cache path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist yet - return empty manager
		if os.IsNotExist(err) {
			m := newEmpty()
			m.path = path
			return m, nil
		}
		return newEmpty(), fmt.Errorf("failed to read workspaces file: %w", err)
	}

	var m Manager
	if err := json.Unmarshal(data, &m); err != nil {
		// Corrupt JSON - return empty manager and log warning
		_, _ = fmt.Fprintf(os.Stderr, "Warning: corrupt workspaces file, resetting: %v\n", err)
		m = *newEmpty()
	}

	if m.Workspaces == nil {
		m.Workspaces = []Workspace{}
	}

	// Clean up invalid workspace paths (files that no longer exist)
	m.cleanupInvalidWorkspaces()

	m.path = path
	return &m, nil
}

// Save writes workspace data to disk
func (m *Manager) Save() error {
	if m.path == "" {
		return fmt.Errorf("workspace path not set")
	}

	// Ensure cache directory exists
	dir := filepath.Dir(m.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal workspaces: %w", err)
	}

	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return fmt.Errorf("failed to write workspaces file: %w", err)
	}

	return nil
}

// RecordAccess updates or creates a workspace entry when accessed
func (m *Manager) RecordAccess(makefilePath string) {
	// Convert to absolute path
	absPath, err := filepath.Abs(makefilePath)
	if err != nil {
		absPath = makefilePath
	}

	// Find existing workspace
	for i := range m.Workspaces {
		if m.Workspaces[i].Path == absPath {
			m.Workspaces[i].LastAccessed = time.Now()
			m.Workspaces[i].AccessCount++
			return
		}
	}

	// Create new workspace entry
	m.Workspaces = append(m.Workspaces, Workspace{
		Path:         absPath,
		LastAccessed: time.Now(),
		AccessCount:  1,
		IsFavorite:   false,
	})
}

// GetRecent returns the most recently accessed workspaces, sorted by last access time
// Favorites always appear first, then sorted by last accessed
func (m *Manager) GetRecent(limit int) []Workspace {
	if limit <= 0 {
		limit = maxRecentWorkspaces
	}

	// Sort workspaces: favorites first, then by last accessed
	sorted := make([]Workspace, len(m.Workspaces))
	copy(sorted, m.Workspaces)

	sort.Slice(sorted, func(i, j int) bool {
		// Favorites first
		if sorted[i].IsFavorite != sorted[j].IsFavorite {
			return sorted[i].IsFavorite
		}
		// Then by last accessed
		return sorted[i].LastAccessed.After(sorted[j].LastAccessed)
	})

	// Return up to limit
	if len(sorted) > limit {
		return sorted[:limit]
	}
	return sorted
}

// ToggleFavorite toggles the favorite status of a workspace
func (m *Manager) ToggleFavorite(makefilePath string) {
	// Convert to absolute path
	absPath, err := filepath.Abs(makefilePath)
	if err != nil {
		absPath = makefilePath
	}

	// Find and toggle
	for i := range m.Workspaces {
		if m.Workspaces[i].Path == absPath {
			m.Workspaces[i].IsFavorite = !m.Workspaces[i].IsFavorite
			return
		}
	}
}

// GetRelativePath returns a display-friendly relative path from cwd to makefile
// Returns basename on error or if relative path is complex
func (m *Manager) GetRelativePath(makefilePath, cwd string) string {
	// Try to compute relative path
	relPath, err := filepath.Rel(cwd, makefilePath)
	if err != nil {
		// Fallback to basename
		return filepath.Base(makefilePath)
	}

	// Clean up relative path for display
	// "Makefile" -> "./Makefile"
	if relPath == "Makefile" {
		return "./Makefile"
	}

	// Ensure paths that don't start with "." or ".." get "./" prefix
	if !strings.HasPrefix(relPath, ".") {
		return "./" + relPath
	}

	return relPath
}

// cleanupInvalidWorkspaces removes workspace entries for Makefiles that no longer exist
func (m *Manager) cleanupInvalidWorkspaces() {
	valid := make([]Workspace, 0, len(m.Workspaces))
	for _, ws := range m.Workspaces {
		if _, err := os.Stat(ws.Path); err == nil {
			valid = append(valid, ws)
		}
	}
	m.Workspaces = valid
}

// newEmpty creates a new empty workspace manager
func newEmpty() *Manager {
	return &Manager{
		Workspaces: []Workspace{},
	}
}

// NewEmpty creates a new empty workspace manager (exported for tests and error cases)
func NewEmpty() *Manager {
	return newEmpty()
}

// getCachePath returns the path to the workspaces cache file
func getCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(cacheDir, "lazymake", workspaceFileName), nil
}
