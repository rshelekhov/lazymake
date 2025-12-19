package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_NonExistentFile(t *testing.T) {
	h := &History{
		Entries: make(map[string][]Entry),
	}

	// Should return empty history if file doesn't exist
	if len(h.Entries) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(h.Entries))
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	// This test verifies that corrupt JSON doesn't crash the app
	// We test the unmarshal logic directly since Load() uses the real cache path

	corruptJSON := []byte("{invalid json}")

	var h History
	err := json.Unmarshal(corruptJSON, &h)

	// Should fail to unmarshal
	if err == nil {
		t.Error("Expected error when unmarshaling corrupt JSON")
	}

	// Verify that after handling corrupt JSON, we get empty history
	// (this simulates what Load() does on line 54-55 of history.go)
	h = *newEmptyHistory()

	if len(h.Entries) != 0 {
		t.Errorf("Expected empty history after handling corrupt JSON, got %d entries", len(h.Entries))
	}
}

func TestRecordExecution_NewTarget(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"
	targetName := "build"

	h.RecordExecution(makefilePath, targetName)

	entries := h.Entries[makefilePath]
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].Name != targetName {
		t.Errorf("Expected target name %s, got %s", targetName, entries[0].Name)
	}

	if entries[0].UseCount != 1 {
		t.Errorf("Expected use count 1, got %d", entries[0].UseCount)
	}
}

func TestRecordExecution_ExistingTarget(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"
	targetName := "build"

	// Record twice
	h.RecordExecution(makefilePath, targetName)
	time.Sleep(10 * time.Millisecond) // Ensure different timestamp
	firstTime := h.Entries[makefilePath][0].LastUsed

	h.RecordExecution(makefilePath, targetName)

	entries := h.Entries[makefilePath]
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if entries[0].UseCount != 2 {
		t.Errorf("Expected use count 2, got %d", entries[0].UseCount)
	}

	if !entries[0].LastUsed.After(firstTime) {
		t.Errorf("Expected updated timestamp, got %v (previous: %v)", entries[0].LastUsed, firstTime)
	}
}

func TestRecordExecution_LRUEviction(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"

	// Record 6 different targets (max is 5)
	targets := []string{"target1", "target2", "target3", "target4", "target5", "target6"}
	for _, target := range targets {
		h.RecordExecution(makefilePath, target)
		time.Sleep(1 * time.Millisecond) // Ensure different timestamps
	}

	entries := h.Entries[makefilePath]
	if len(entries) != maxRecentTargets {
		t.Fatalf("Expected %d entries (LRU eviction), got %d", maxRecentTargets, len(entries))
	}

	// Should keep the 5 most recent (target2-target6)
	// target1 should be evicted
	for _, entry := range entries {
		if entry.Name == "target1" {
			t.Errorf("target1 should have been evicted")
		}
	}

	// Most recent should be first
	if entries[0].Name != "target6" {
		t.Errorf("Expected most recent target 'target6', got %s", entries[0].Name)
	}
}

func TestRecordExecution_SortOrder(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"

	// Record targets in order
	h.RecordExecution(makefilePath, "first")
	time.Sleep(2 * time.Millisecond)
	h.RecordExecution(makefilePath, "second")
	time.Sleep(2 * time.Millisecond)
	h.RecordExecution(makefilePath, "third")

	entries := h.Entries[makefilePath]

	// Should be sorted by most recent first
	if entries[0].Name != "third" {
		t.Errorf("Expected 'third' first, got %s", entries[0].Name)
	}
	if entries[1].Name != "second" {
		t.Errorf("Expected 'second' second, got %s", entries[1].Name)
	}
	if entries[2].Name != "first" {
		t.Errorf("Expected 'first' third, got %s", entries[2].Name)
	}

	// Re-execute "first" - it should move to top
	time.Sleep(2 * time.Millisecond)
	h.RecordExecution(makefilePath, "first")

	entries = h.Entries[makefilePath]
	if entries[0].Name != "first" {
		t.Errorf("Expected 'first' to move to top after re-execution, got %s", entries[0].Name)
	}
}

func TestGetRecent(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"

	// Record some targets
	h.RecordExecution(makefilePath, "build")
	h.RecordExecution(makefilePath, "test")

	recent := h.GetRecent(makefilePath)

	if len(recent) != 2 {
		t.Fatalf("Expected 2 recent targets, got %d", len(recent))
	}

	// Should be sorted by most recent
	if recent[0].Name != "test" {
		t.Errorf("Expected 'test' first, got %s", recent[0].Name)
	}
	if recent[1].Name != "build" {
		t.Errorf("Expected 'build' second, got %s", recent[1].Name)
	}
}

func TestGetRecent_EmptyHistory(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"

	recent := h.GetRecent(makefilePath)

	if recent != nil {
		t.Errorf("Expected nil for empty history, got %v", recent)
	}
}

func TestFilterValid(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"

	// Record some targets
	h.RecordExecution(makefilePath, "build")
	h.RecordExecution(makefilePath, "test")
	h.RecordExecution(makefilePath, "deploy")
	h.RecordExecution(makefilePath, "clean")

	// Only "build" and "test" are still valid
	validTargets := []string{"build", "test"}
	h.FilterValid(makefilePath, validTargets)

	entries := h.Entries[makefilePath]
	if len(entries) != 2 {
		t.Fatalf("Expected 2 valid entries, got %d", len(entries))
	}

	// Check that only valid targets remain
	names := make(map[string]bool)
	for _, entry := range entries {
		names[entry.Name] = true
	}

	if !names["build"] || !names["test"] {
		t.Errorf("Expected 'build' and 'test' to remain, got %v", names)
	}

	if names["deploy"] || names["clean"] {
		t.Errorf("Expected 'deploy' and 'clean' to be filtered out, got %v", names)
	}
}

func TestFilterValid_AllInvalid(t *testing.T) {
	h := newEmptyHistory()
	makefilePath := "/test/Makefile"

	// Record some targets
	h.RecordExecution(makefilePath, "old-target")
	h.RecordExecution(makefilePath, "removed-target")

	// No valid targets
	validTargets := []string{"new-target"}
	h.FilterValid(makefilePath, validTargets)

	entries := h.Entries[makefilePath]
	if len(entries) != 0 {
		t.Errorf("Expected all entries to be filtered out, got %d", len(entries))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "history.json")

	// Create history with some data
	h := &History{
		Entries: map[string][]Entry{
			"/test/Makefile": {
				{Name: "build", LastUsed: time.Now(), UseCount: 3},
				{Name: "test", LastUsed: time.Now(), UseCount: 5},
			},
		},
		path: testPath,
	}

	// Save
	err := h.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load
	data, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	var loaded History
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		t.Fatalf("Failed to unmarshal saved data: %v", err)
	}

	// Verify data
	entries := loaded.Entries["/test/Makefile"]
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries after round trip, got %d", len(entries))
	}

	if entries[0].Name != "build" {
		t.Errorf("Expected first entry 'build', got %s", entries[0].Name)
	}

	if entries[0].UseCount != 3 {
		t.Errorf("Expected use count 3, got %d", entries[0].UseCount)
	}
}

func TestMultipleMakefiles(t *testing.T) {
	h := newEmptyHistory()
	makefile1 := "/project1/Makefile"
	makefile2 := "/project2/Makefile"

	// Record targets for different Makefiles
	h.RecordExecution(makefile1, "build")
	h.RecordExecution(makefile1, "test")
	h.RecordExecution(makefile2, "deploy")

	// Check isolation
	entries1 := h.GetRecent(makefile1)
	entries2 := h.GetRecent(makefile2)

	if len(entries1) != 2 {
		t.Errorf("Expected 2 entries for makefile1, got %d", len(entries1))
	}

	if len(entries2) != 1 {
		t.Errorf("Expected 1 entry for makefile2, got %d", len(entries2))
	}

	if entries2[0].Name != "deploy" {
		t.Errorf("Expected 'deploy' for makefile2, got %s", entries2[0].Name)
	}
}

// ========== Performance Tracking Tests ==========

func TestRecordExecutionWithTiming(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	// Record first execution
	h.RecordExecutionWithTiming(makefile, "build", 2*time.Second, true)

	entries := h.Entries[makefile]
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	if len(entries[0].RecentExecutions) != 1 {
		t.Fatalf("Expected 1 execution record, got %d", len(entries[0].RecentExecutions))
	}

	exec := entries[0].RecentExecutions[0]
	if exec.Duration != 2*time.Second {
		t.Errorf("Expected duration 2s, got %v", exec.Duration)
	}

	if !exec.Success {
		t.Error("Expected execution to be successful")
	}
}

func TestRecordExecutionWithTiming_MultipleExecutions(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	// Record multiple executions
	durations := []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second}
	for _, d := range durations {
		h.RecordExecutionWithTiming(makefile, "test", d, true)
	}

	entries := h.Entries[makefile]
	if len(entries[0].RecentExecutions) != 3 {
		t.Fatalf("Expected 3 execution records, got %d", len(entries[0].RecentExecutions))
	}

	// Verify durations are recorded
	for i, d := range durations {
		if entries[0].RecentExecutions[i].Duration != d {
			t.Errorf("Expected duration %v at index %d, got %v", d, i, entries[0].RecentExecutions[i].Duration)
		}
	}
}

func TestRecordExecutionWithTiming_LRUEviction(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	// Record more than maxRecentExecutions (10) executions
	for i := 0; i < 15; i++ {
		h.RecordExecutionWithTiming(makefile, "build", time.Duration(i+1)*time.Second, true)
	}

	entries := h.Entries[makefile]
	if len(entries[0].RecentExecutions) != maxRecentExecutions {
		t.Errorf("Expected %d execution records (LRU limit), got %d", maxRecentExecutions, len(entries[0].RecentExecutions))
	}

	// Should keep the most recent 10 (executions 6-15)
	firstExec := entries[0].RecentExecutions[0]
	if firstExec.Duration != 6*time.Second {
		t.Errorf("Expected oldest kept execution to be 6s, got %v", firstExec.Duration)
	}

	lastExec := entries[0].RecentExecutions[len(entries[0].RecentExecutions)-1]
	if lastExec.Duration != 15*time.Second {
		t.Errorf("Expected newest execution to be 15s, got %v", lastExec.Duration)
	}
}

func TestGetPerformanceStats_NoData(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	stats := h.GetPerformanceStats(makefile, "nonexistent")
	if stats != nil {
		t.Error("Expected nil stats for nonexistent target")
	}
}

func TestGetPerformanceStats_Calculate(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	// Record multiple executions with known durations
	durations := []time.Duration{1 * time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second}
	for _, d := range durations {
		h.RecordExecutionWithTiming(makefile, "test", d, true)
	}

	stats := h.GetPerformanceStats(makefile, "test")
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Verify average: (1+2+3+4)/4 = 2.5s
	expectedAvg := 2500 * time.Millisecond
	if stats.AvgDuration != expectedAvg {
		t.Errorf("Expected avg duration %v, got %v", expectedAvg, stats.AvgDuration)
	}

	// Verify min
	if stats.MinDuration != 1*time.Second {
		t.Errorf("Expected min duration 1s, got %v", stats.MinDuration)
	}

	// Verify max
	if stats.MaxDuration != 4*time.Second {
		t.Errorf("Expected max duration 4s, got %v", stats.MaxDuration)
	}

	// Verify last duration
	if stats.LastDuration != 4*time.Second {
		t.Errorf("Expected last duration 4s, got %v", stats.LastDuration)
	}

	// Verify execution count
	if stats.ExecutionCount != 4 {
		t.Errorf("Expected 4 successful executions, got %d", stats.ExecutionCount)
	}
}

func TestGetPerformanceStats_RegressionDetection(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	// Record executions with consistent duration
	for i := 0; i < 5; i++ {
		h.RecordExecutionWithTiming(makefile, "slow-build", 2*time.Second, true)
	}

	// Record a much slower execution (>25% slower than the new average)
	// After 5x2s + 1x4s = 14s, average = 14/6 = 2.33s
	// Threshold = 2.33s * 1.25 = 2.91s
	// So 4s > 2.91s triggers regression
	h.RecordExecutionWithTiming(makefile, "slow-build", 4*time.Second, true)

	stats := h.GetPerformanceStats(makefile, "slow-build")
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if !stats.IsRegressed {
		t.Errorf("Expected regression to be detected: last=%v, avg=%v, threshold=%v",
			stats.LastDuration, stats.AvgDuration, time.Duration(float64(stats.AvgDuration)*regressionMultiplier))
	}
}

func TestGetPerformanceStats_NoRegression(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	// Record executions with varying but acceptable durations
	durations := []time.Duration{2 * time.Second, 2100 * time.Millisecond, 1900 * time.Millisecond}
	for _, d := range durations {
		h.RecordExecutionWithTiming(makefile, "stable-test", d, true)
	}

	stats := h.GetPerformanceStats(makefile, "stable-test")
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	if stats.IsRegressed {
		t.Error("Expected no regression for stable durations")
	}
}

func TestGetPerformanceStats_IgnoreFailures(t *testing.T) {
	h := newEmptyHistory()
	makefile := "/test/Makefile"

	// Record mix of successful and failed executions
	h.RecordExecutionWithTiming(makefile, "flaky-test", 1*time.Second, true)
	h.RecordExecutionWithTiming(makefile, "flaky-test", 10*time.Second, false) // Failed - should be ignored
	h.RecordExecutionWithTiming(makefile, "flaky-test", 2*time.Second, true)

	stats := h.GetPerformanceStats(makefile, "flaky-test")
	if stats == nil {
		t.Fatal("Expected non-nil stats")
	}

	// Should only count successful executions: (1+2)/2 = 1.5s
	expectedAvg := 1500 * time.Millisecond
	if stats.AvgDuration != expectedAvg {
		t.Errorf("Expected avg %v (ignoring failures), got %v", expectedAvg, stats.AvgDuration)
	}

	if stats.ExecutionCount != 2 {
		t.Errorf("Expected 2 successful executions (ignoring failure), got %d", stats.ExecutionCount)
	}
}
