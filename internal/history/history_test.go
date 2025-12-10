package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoad_NonExistentFile(t *testing.T) {
	// Create a temporary directory for test cache
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "history.json")

	h := &History{
		Entries: make(map[string][]Entry),
		path:    testPath,
	}

	// Should return empty history if file doesn't exist
	if len(h.Entries) != 0 {
		t.Errorf("Expected empty history, got %d entries", len(h.Entries))
	}
}

func TestLoad_CorruptJSON(t *testing.T) {
	// Create a temporary directory and corrupt file
	tempDir := t.TempDir()
	testPath := filepath.Join(tempDir, "history.json")

	// Write corrupt JSON
	err := os.WriteFile(testPath, []byte("{invalid json}"), 0644)
	if err != nil {
		t.Fatalf("Failed to write corrupt file: %v", err)
	}

	// Load should return empty history and not crash
	h, err := Load()
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if len(h.Entries) != 0 {
		t.Errorf("Expected empty history for corrupt JSON, got %d entries", len(h.Entries))
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
