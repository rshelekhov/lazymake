package export

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RotateFiles performs file rotation based on config settings
func RotateFiles(outputDir, targetName string, config *Config) error {
	if config.MaxFiles == 0 && config.KeepDays == 0 && config.MaxFileSize == 0 {
		return nil // No rotation configured
	}

	// Expand and prepare output directory
	expandedDir := getExpandedOutputDir(outputDir)

	// Sanitize target name for file matching
	sanitized := strings.ReplaceAll(targetName, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	// Collect file information
	fileInfos := collectFileInfos(expandedDir, sanitized)
	if len(fileInfos) == 0 {
		return nil // No files to rotate
	}

	// Apply size-based rotation
	if config.MaxFileSize > 0 {
		fileInfos = applySizeLimit(fileInfos, config.MaxFileSize)
	}

	// Apply age-based rotation
	if config.KeepDays > 0 {
		fileInfos = applyAgeCutoff(fileInfos, config.KeepDays)
	}

	// Apply count-based rotation
	if config.MaxFiles > 0 {
		applyMaxFilesLimit(fileInfos, config.MaxFiles)
	}

	return nil
}

// getExpandedOutputDir expands the output directory path
func getExpandedOutputDir(outputDir string) string {
	expandedDir := expandPath(outputDir)
	if expandedDir != "" {
		return expandedDir
	}

	// Use default cache directory
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		if home, err := os.UserHomeDir(); err == nil {
			cacheDir = filepath.Join(home, ".cache")
		}
	}
	return filepath.Join(cacheDir, "lazymake", "exports")
}

// fileInfo holds information about a file for rotation
type fileInfo struct {
	path    string
	modTime time.Time
	size    int64
}

// collectFileInfos collects file information for rotation
func collectFileInfos(expandedDir, sanitized string) []fileInfo {
	// Find all files for this target (both .json and .log)
	patterns := []string{
		filepath.Join(expandedDir, sanitized+"*.json"),
		filepath.Join(expandedDir, sanitized+"*.log"),
	}

	var allFiles []string
	for _, pattern := range patterns {
		files, err := filepath.Glob(pattern)
		if err != nil {
			continue
		}
		allFiles = append(allFiles, files...)
	}

	if len(allFiles) == 0 {
		return nil
	}

	// Get file info for all files
	var fileInfos []fileInfo
	for _, f := range allFiles {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		fileInfos = append(fileInfos, fileInfo{
			path:    f,
			modTime: info.ModTime(),
			size:    info.Size(),
		})
	}

	// Sort by modification time (oldest first)
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].modTime.Before(fileInfos[j].modTime)
	})

	return fileInfos
}

// applySizeLimit removes files that exceed the configured size limit
func applySizeLimit(fileInfos []fileInfo, maxSizeMB int64) []fileInfo {
	maxBytes := maxSizeMB * 1024 * 1024
	var remaining []fileInfo
	for _, fi := range fileInfos {
		if fi.size > maxBytes {
			_ = os.Remove(fi.path)
		} else {
			remaining = append(remaining, fi)
		}
	}
	return remaining
}

// applyAgeCutoff removes files older than the specified number of days
func applyAgeCutoff(fileInfos []fileInfo, keepDays int) []fileInfo {
	cutoff := time.Now().AddDate(0, 0, -keepDays)

	// Remove old files
	for _, fi := range fileInfos {
		if fi.modTime.Before(cutoff) {
			_ = os.Remove(fi.path)
		}
	}

	// Filter to keep only remaining files
	var remainingFiles []fileInfo
	for _, fi := range fileInfos {
		if fi.modTime.After(cutoff) || fi.modTime.Equal(cutoff) {
			remainingFiles = append(remainingFiles, fi)
		}
	}

	return remainingFiles
}

// applyMaxFilesLimit removes excess files beyond the maximum count
func applyMaxFilesLimit(fileInfos []fileInfo, maxFiles int) {
	// Group files by extension to maintain balance
	jsonFiles := []string{}
	logFiles := []string{}

	for _, fi := range fileInfos {
		if strings.HasSuffix(fi.path, ".json") {
			jsonFiles = append(jsonFiles, fi.path)
		} else if strings.HasSuffix(fi.path, ".log") {
			logFiles = append(logFiles, fi.path)
		}
	}

	// Remove oldest files if exceeding MaxFiles
	if len(jsonFiles) > maxFiles {
		toRemove := jsonFiles[:len(jsonFiles)-maxFiles]
		for _, f := range toRemove {
			_ = os.Remove(f)
		}
	}

	if len(logFiles) > maxFiles {
		toRemove := logFiles[:len(logFiles)-maxFiles]
		for _, f := range toRemove {
			_ = os.Remove(f)
		}
	}
}
