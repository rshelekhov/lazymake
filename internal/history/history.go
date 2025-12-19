package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const (
	maxRecentTargets     = 5
	maxRecentExecutions  = 10
	regressionMultiplier = 1.25
	historyFileName      = "history.json"
)

// ExecutionRecord represents a single target execution with timing
type ExecutionRecord struct {
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
	Success   bool          `json:"success"`
}

// Entry represents a single target execution record
type Entry struct {
	Name             string            `json:"name"`
	LastUsed         time.Time         `json:"last_used"`
	UseCount         int               `json:"use_count"`
	RecentExecutions []ExecutionRecord `json:"recent_executions,omitempty"`
}

// PerformanceStats contains calculated performance statistics for a target
type PerformanceStats struct {
	AvgDuration    time.Duration
	LastDuration   time.Duration
	MinDuration    time.Duration
	MaxDuration    time.Duration
	ExecutionCount int
	IsRegressed    bool // >25% slower than avg
}

// History manages execution history across multiple Makefiles
type History struct {
	// Map of absolute Makefile path -> list of recent target entries
	Entries map[string][]Entry `json:"entries"`
	path    string             // Cache file path
}

// Load reads history from the cache directory
// Returns an empty history on error (graceful degradation)
func Load() (*History, error) {
	path, err := getCachePath()
	if err != nil {
		return newEmptyHistory(), fmt.Errorf("failed to get cache path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// File doesn't exist yet - return empty history
		if os.IsNotExist(err) {
			h := newEmptyHistory()
			h.path = path
			return h, nil
		}
		return newEmptyHistory(), fmt.Errorf("failed to read history file: %w", err)
	}

	var h History
	if err := json.Unmarshal(data, &h); err != nil {
		// Corrupt JSON - return empty history and log warning
		_, _ = fmt.Fprintf(os.Stderr, "Warning: corrupt history file, resetting: %v\n", err)
		h = *newEmptyHistory()
	}

	if h.Entries == nil {
		h.Entries = make(map[string][]Entry)
	}

	h.path = path
	return &h, nil
}

// Save writes history to disk
func (h *History) Save() error {
	if h.path == "" {
		return fmt.Errorf("history path not set")
	}

	// Ensure cache directory exists
	dir := filepath.Dir(h.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	data, err := json.MarshalIndent(h, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	if err := os.WriteFile(h.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	return nil
}

// RecordExecution adds or updates a target execution record (legacy method without timing)
// Implements LRU eviction: keeps only the maxRecentTargets most recent targets
func (h *History) RecordExecution(makefilePath, targetName string) {
	h.RecordExecutionWithTiming(makefilePath, targetName, 0, true)
}

// RecordExecutionWithTiming adds or updates a target execution record with performance data
// Implements LRU eviction: keeps only the maxRecentTargets most recent targets
// Implements execution history LRU: keeps only the maxRecentExecutions most recent executions
func (h *History) RecordExecutionWithTiming(makefilePath, targetName string, duration time.Duration, success bool) {
	now := time.Now()

	entries := h.Entries[makefilePath]

	// Find if target already exists
	found := false
	for i := range entries {
		if entries[i].Name == targetName {
			// Update existing entry
			entries[i].LastUsed = now
			entries[i].UseCount++

			// Add execution record if duration is provided
			if duration > 0 {
				exec := ExecutionRecord{
					Duration:  duration,
					Timestamp: now,
					Success:   success,
				}
				entries[i].RecentExecutions = append(entries[i].RecentExecutions, exec)

				// Keep only the most recent maxRecentExecutions
				if len(entries[i].RecentExecutions) > maxRecentExecutions {
					entries[i].RecentExecutions = entries[i].RecentExecutions[len(entries[i].RecentExecutions)-maxRecentExecutions:]
				}
			}

			found = true
			break
		}
	}

	if !found {
		// Add new entry
		entry := Entry{
			Name:     targetName,
			LastUsed: now,
			UseCount: 1,
		}

		// Add execution record if duration is provided
		if duration > 0 {
			entry.RecentExecutions = []ExecutionRecord{{
				Duration:  duration,
				Timestamp: now,
				Success:   success,
			}}
		}

		entries = append(entries, entry)
	}

	// Sort by LastUsed descending (most recent first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].LastUsed.After(entries[j].LastUsed)
	})

	// Keep only the most recent maxRecentTargets
	if len(entries) > maxRecentTargets {
		entries = entries[:maxRecentTargets]
	}

	h.Entries[makefilePath] = entries
}

// GetPerformanceStats calculates and returns performance statistics for a target
// Returns nil if target has no performance data
func (h *History) GetPerformanceStats(makefilePath, targetName string) *PerformanceStats {
	entries := h.Entries[makefilePath]

	// Find target entry
	var entry *Entry
	for i := range entries {
		if entries[i].Name == targetName {
			entry = &entries[i]
			break
		}
	}

	if entry == nil || len(entry.RecentExecutions) == 0 {
		return nil
	}

	// Calculate statistics from successful executions only
	var totalDuration time.Duration
	var minDuration, maxDuration time.Duration
	successCount := 0

	for _, exec := range entry.RecentExecutions {
		if !exec.Success {
			continue
		}

		successCount++
		totalDuration += exec.Duration

		if minDuration == 0 || exec.Duration < minDuration {
			minDuration = exec.Duration
		}
		if exec.Duration > maxDuration {
			maxDuration = exec.Duration
		}
	}

	if successCount == 0 {
		return nil
	}

	avgDuration := totalDuration / time.Duration(successCount)

	// Get last execution (may be failed)
	lastExec := entry.RecentExecutions[len(entry.RecentExecutions)-1]

	// Detect regression: last run is >25% slower than average
	isRegressed := false
	if lastExec.Success && avgDuration > 0 {
		threshold := time.Duration(float64(avgDuration) * regressionMultiplier)
		isRegressed = lastExec.Duration > threshold
	}

	return &PerformanceStats{
		AvgDuration:    avgDuration,
		LastDuration:   lastExec.Duration,
		MinDuration:    minDuration,
		MaxDuration:    maxDuration,
		ExecutionCount: successCount,
		IsRegressed:    isRegressed,
	}
}

// GetRecent returns up to maxRecentTargets recent targets for a Makefile
// Returns targets sorted by LastUsed descending (most recent first)
func (h *History) GetRecent(makefilePath string) []Entry {
	entries := h.Entries[makefilePath]
	if len(entries) == 0 {
		return nil
	}

	// Return a copy to avoid external modifications
	result := make([]Entry, len(entries))
	copy(result, entries)
	return result
}

// FilterValid removes targets that no longer exist in the Makefile
// This prevents showing stale targets that have been removed or renamed
func (h *History) FilterValid(makefilePath string, validTargets []string) {
	entries := h.Entries[makefilePath]
	if len(entries) == 0 {
		return
	}

	// Build set of valid target names for O(1) lookup
	validSet := make(map[string]bool)
	for _, name := range validTargets {
		validSet[name] = true
	}

	// Filter entries
	filtered := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		if validSet[entry.Name] {
			filtered = append(filtered, entry)
		}
	}

	h.Entries[makefilePath] = filtered
}

// getCachePath returns the platform-appropriate cache file path
// Prefers XDG cache directory, falls back to ~/.cache
func getCachePath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		// Fallback to ~/.cache for Unix-like systems
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		cacheDir = filepath.Join(home, ".cache")
	}

	return filepath.Join(cacheDir, "lazymake", historyFileName), nil
}

// newEmptyHistory creates a new empty history instance
func newEmptyHistory() *History {
	return &History{
		Entries: make(map[string][]Entry),
	}
}
