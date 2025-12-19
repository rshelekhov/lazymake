package workspace

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMaxDepth    = 3
	defaultScanTimeout = 5 * time.Second
)

// DiscoveryOptions configures Makefile discovery behavior
type DiscoveryOptions struct {
	MaxDepth        int           // Maximum directory depth to search (default: 3)
	ExcludePatterns []string      // Directory names to exclude
	Timeout         time.Duration // Max time to scan (default: 5s)
}

// DefaultDiscoveryOptions returns default discovery settings
func DefaultDiscoveryOptions() DiscoveryOptions {
	return DiscoveryOptions{
		MaxDepth: defaultMaxDepth,
		ExcludePatterns: []string{
			".git",
			"node_modules",
			"vendor",
			".venv",
			"venv",
			"build",
			"dist",
			".cache",
			".idea",
			".vscode",
			"target", // Rust/Java build dir
			"__pycache__",
		},
		Timeout: defaultScanTimeout,
	}
}

// DiscoveryResult represents a discovered Makefile
type DiscoveryResult struct {
	Path    string    // Absolute path to Makefile
	RelPath string    // Relative path from search root
	ModTime time.Time // Last modification time
}

// pathDepth tracks path and its depth for BFS
type pathDepth struct {
	path  string
	depth int
}

// DiscoverMakefiles finds all Makefiles in a directory tree using BFS
func DiscoverMakefiles(rootDir string, opts DiscoveryOptions) ([]DiscoveryResult, error) {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), opts.Timeout)
	defer cancel()

	// Build exclusion map for fast lookup
	excludeDirs := make(map[string]bool)
	for _, pattern := range opts.ExcludePatterns {
		excludeDirs[pattern] = true
	}

	var results []DiscoveryResult

	// BFS queue
	queue := []pathDepth{{path: rootDir, depth: 0}}

	for len(queue) > 0 {
		// Check timeout
		select {
		case <-ctx.Done():
			// Timeout reached, return what we have
			return results, nil
		default:
		}

		// Dequeue
		current := queue[0]
		queue = queue[1:]

		// Check depth limit
		if current.depth > opts.MaxDepth {
			continue
		}

		// Read directory entries
		entries, err := os.ReadDir(current.path)
		if err != nil {
			// Permission denied or other error - skip this directory
			continue
		}

		for _, entry := range entries {
			fullPath := filepath.Join(current.path, entry.Name())

			if entry.IsDir() {
				// Skip excluded directories
				if excludeDirs[entry.Name()] {
					continue
				}

				// Add to queue for BFS
				queue = append(queue, pathDepth{
					path:  fullPath,
					depth: current.depth + 1,
				})
			} else if IsMakefile(entry.Name()) {
				// Check if it's a Makefile
				// Get file info for mod time
				info, err := entry.Info()
				if err != nil {
					continue
				}

				// Compute relative path from root
				relPath, err := filepath.Rel(rootDir, fullPath)
				if err != nil {
					relPath = fullPath
				}

				results = append(results, DiscoveryResult{
					Path:    fullPath,
					RelPath: relPath,
					ModTime: info.ModTime(),
				})
			}
		}
	}

	return results, nil
}

// FindMakefilesInParents searches upward from a directory to find Makefiles
// Useful for finding project root when in a subdirectory
func FindMakefilesInParents(startDir string, maxLevels int) ([]string, error) {
	var results []string
	currentDir := startDir

	for level := 0; level < maxLevels; level++ {
		// Check for Makefile in current directory
		makefilePath := filepath.Join(currentDir, "Makefile")
		if _, err := os.Stat(makefilePath); err == nil {
			absPath, _ := filepath.Abs(makefilePath)
			results = append(results, absPath)
		}

		// Check for makefile (lowercase)
		makefilePath = filepath.Join(currentDir, "makefile")
		if _, err := os.Stat(makefilePath); err == nil {
			absPath, _ := filepath.Abs(makefilePath)
			results = append(results, absPath)
		}

		// Move up one directory
		parent := filepath.Dir(currentDir)
		if parent == currentDir {
			// Reached root
			break
		}
		currentDir = parent
	}

	return results, nil
}

// IsMakefile returns true if the given path is a Makefile
func IsMakefile(path string) bool {
	name := filepath.Base(path)

	// Exact matches
	if name == "Makefile" || name == "makefile" || name == "GNUmakefile" {
		return true
	}

	// Files starting with "Makefile" (e.g., Makefile.local, Makefile.inc)
	if strings.HasPrefix(name, "Makefile") || strings.HasPrefix(name, "makefile") {
		return true
	}

	// Common Makefile extensions
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".mk" || ext == ".mak"
}

// FormatSize formats a file size in human-readable format
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return strconv.FormatInt(bytes, 10) + " B"
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	div := int64(unit)
	exp := 0

	for n := bytes / unit; n >= unit && exp < len(units)-1; n /= unit {
		div *= unit
		exp++
	}

	// Clamp exp to valid range
	if exp < 0 {
		exp = 0
	}
	if exp >= len(units) {
		exp = len(units) - 1
	}

	value := float64(bytes) / float64(div)
	return fmt.Sprintf("%.1f %s", value, units[exp])
}
