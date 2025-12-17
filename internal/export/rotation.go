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
	if config.MaxFiles == 0 && config.KeepDays == 0 {
		return nil // No rotation configured
	}

	// Expand output directory
	expandedDir := expandPath(outputDir)
	if expandedDir == "" {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			if home, err := os.UserHomeDir(); err == nil {
				cacheDir = filepath.Join(home, ".cache")
			}
		}
		expandedDir = filepath.Join(cacheDir, "lazymake", "exports")
	}

	// Sanitize target name
	sanitized := strings.ReplaceAll(targetName, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

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
		return nil // No files to rotate
	}

	// Get file info for all files
	type fileInfo struct {
		path    string
		modTime time.Time
		size    int64
	}

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

	// Apply KeepDays rotation
	if config.KeepDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -config.KeepDays)
		for _, fi := range fileInfos {
			if fi.modTime.Before(cutoff) {
				_ = os.Remove(fi.path)
			}
		}

		// Re-filter fileInfos to only include non-removed files
		var remainingFiles []fileInfo
		for _, fi := range fileInfos {
			if fi.modTime.After(cutoff) || fi.modTime.Equal(cutoff) {
				remainingFiles = append(remainingFiles, fi)
			}
		}
		fileInfos = remainingFiles
	}

	// Apply MaxFiles rotation
	// Group files by extension to maintain balance
	if config.MaxFiles > 0 {
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
		if len(jsonFiles) > config.MaxFiles {
			toRemove := jsonFiles[:len(jsonFiles)-config.MaxFiles]
			for _, f := range toRemove {
				_ = os.Remove(f)
			}
		}

		if len(logFiles) > config.MaxFiles {
			toRemove := logFiles[:len(logFiles)-config.MaxFiles]
			for _, f := range toRemove {
				_ = os.Remove(f)
			}
		}
	}

	return nil
}
