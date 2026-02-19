package export

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRotateFiles_NoRotationConfigured(t *testing.T) {
	config := &Config{
		MaxFiles:    0,
		KeepDays:    0,
		MaxFileSize: 0,
	}

	// Should return nil without error when no rotation is configured
	err := RotateFiles(t.TempDir(), "build", config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRotateFiles_SizeLimit(t *testing.T) {
	dir := t.TempDir()

	// Create files of various sizes
	smallFile := filepath.Join(dir, "build_20250101_000001.json")
	largeFile := filepath.Join(dir, "build_20250101_000002.json")

	// Small file: 100 bytes
	if err := os.WriteFile(smallFile, make([]byte, 100), 0o644); err != nil {
		t.Fatal(err)
	}

	// Large file: 2 MB (exceeds 1 MB limit)
	if err := os.WriteFile(largeFile, make([]byte, 2*1024*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	config := &Config{
		MaxFileSize: 1, // 1 MB
	}

	err := RotateFiles(dir, "build", config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Small file should remain
	if _, err := os.Stat(smallFile); os.IsNotExist(err) {
		t.Error("small file should not have been removed")
	}

	// Large file should be removed
	if _, err := os.Stat(largeFile); !os.IsNotExist(err) {
		t.Error("large file should have been removed")
	}
}

func TestRotateFiles_SizeLimitWithLogFiles(t *testing.T) {
	dir := t.TempDir()

	smallLog := filepath.Join(dir, "build_20250101_000001.log")
	largeLog := filepath.Join(dir, "build_20250101_000002.log")

	if err := os.WriteFile(smallLog, make([]byte, 500*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(largeLog, make([]byte, 3*1024*1024), 0o644); err != nil {
		t.Fatal(err)
	}

	config := &Config{
		MaxFileSize: 1, // 1 MB
	}

	err := RotateFiles(dir, "build", config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(smallLog); os.IsNotExist(err) {
		t.Error("small log should not have been removed")
	}

	if _, err := os.Stat(largeLog); !os.IsNotExist(err) {
		t.Error("large log should have been removed")
	}
}

func TestApplySizeLimit(t *testing.T) {
	tests := []struct {
		name        string
		maxSizeMB   int64
		fileSizes   []int64 // in bytes
		wantRemain  int
		wantRemoved []int // indices of files that should be removed
	}{
		{
			name:       "all under limit",
			maxSizeMB:  10,
			fileSizes:  []int64{100, 200, 500},
			wantRemain: 3,
		},
		{
			name:        "one over limit",
			maxSizeMB:   1,
			fileSizes:   []int64{100, 2 * 1024 * 1024, 500},
			wantRemain:  2,
			wantRemoved: []int{1},
		},
		{
			name:        "all over limit",
			maxSizeMB:   1,
			fileSizes:   []int64{2 * 1024 * 1024, 3 * 1024 * 1024},
			wantRemain:  0,
			wantRemoved: []int{0, 1},
		},
		{
			name:       "exactly at limit",
			maxSizeMB:  1,
			fileSizes:  []int64{1024 * 1024}, // exactly 1 MB
			wantRemain: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			var infos []fileInfo
			for i, size := range tt.fileSizes {
				path := filepath.Join(dir, "file"+string(rune('a'+i))+".json")
				if err := os.WriteFile(path, make([]byte, size), 0o644); err != nil {
					t.Fatal(err)
				}
				infos = append(infos, fileInfo{
					path:    path,
					modTime: time.Now(),
					size:    size,
				})
			}

			remaining := applySizeLimit(infos, tt.maxSizeMB)
			if len(remaining) != tt.wantRemain {
				t.Errorf("expected %d remaining files, got %d", tt.wantRemain, len(remaining))
			}

			for _, idx := range tt.wantRemoved {
				if _, err := os.Stat(infos[idx].path); !os.IsNotExist(err) {
					t.Errorf("file at index %d should have been removed", idx)
				}
			}
		})
	}
}
