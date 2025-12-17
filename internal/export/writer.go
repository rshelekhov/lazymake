package export

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteJSON writes the execution record as JSON to the specified path
func WriteJSON(record *ExecutionRecord, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Marshal to JSON with pretty printing
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to file atomically
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write JSON file: %w", err)
	}

	return nil
}

// WriteLog writes the execution record as a plain text log to the specified path
func WriteLog(record *ExecutionRecord, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}

	// Format as log
	logContent := record.FormatLog()

	// Write to file
	if err := os.WriteFile(path, []byte(logContent), 0644); err != nil {
		return fmt.Errorf("failed to write log file: %w", err)
	}

	return nil
}

// GenerateExportPath generates the full path for an export file
func GenerateExportPath(outputDir, filename string) string {
	// Expand ~ and environment variables
	expandedDir := expandPath(outputDir)

	// If still empty, use default cache directory
	if expandedDir == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			// Fallback to ~/.cache for Unix-like systems
			if home, err := os.UserHomeDir(); err == nil {
				cacheDir = filepath.Join(home, ".cache")
			}
		}
		expandedDir = filepath.Join(cacheDir, "lazymake", "exports")
	}

	return filepath.Join(expandedDir, filename)
}

// expandPath expands ~ and environment variables in a path
func expandPath(path string) string {
	if path == "" {
		return path
	}

	// Expand ~ to home directory
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	// Expand environment variables
	path = os.ExpandEnv(path)

	return path
}
