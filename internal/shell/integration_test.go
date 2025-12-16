package shell

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectShell(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     string
	}{
		{
			name:     "bash shell",
			shellEnv: "/bin/bash",
			want:     "bash",
		},
		{
			name:     "zsh shell",
			shellEnv: "/usr/bin/zsh",
			want:     "zsh",
		},
		{
			name:     "fish shell",
			shellEnv: "/usr/local/bin/fish",
			want:     "fish",
		},
		{
			name:     "unknown shell",
			shellEnv: "/bin/sh",
			want:     "none",
		},
		{
			name:     "empty shell",
			shellEnv: "",
			want:     "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original SHELL env
			originalShell := os.Getenv("SHELL")
			defer os.Setenv("SHELL", originalShell)

			// Set test SHELL env
			os.Setenv("SHELL", tt.shellEnv)

			got := DetectShell()
			if got != tt.want {
				t.Errorf("DetectShell() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHistoryFile(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	tests := []struct {
		name      string
		shellType string
		want      string
		wantErr   bool
	}{
		{
			name:      "bash history",
			shellType: "bash",
			want:      filepath.Join(home, ".bash_history"),
			wantErr:   false,
		},
		{
			name:      "zsh history",
			shellType: "zsh",
			want:      filepath.Join(home, ".zsh_history"),
			wantErr:   false,
		},
		{
			name:      "fish history",
			shellType: "fish",
			want:      filepath.Join(home, ".local/share/fish/fish_history"),
			wantErr:   false,
		},
		{
			name:      "unknown shell",
			shellType: "unknown",
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetHistoryFile(tt.shellType)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetHistoryFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetHistoryFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetHistoryFileWithHISTFILE(t *testing.T) {
	// Save original HISTFILE env
	originalHistfile := os.Getenv("HISTFILE")
	defer func() {
		if originalHistfile != "" {
			os.Setenv("HISTFILE", originalHistfile)
		} else {
			os.Unsetenv("HISTFILE")
		}
	}()

	// Set custom HISTFILE
	customPath := "/tmp/custom_history"
	os.Setenv("HISTFILE", customPath)

	got, err := GetHistoryFile("zsh")
	if err != nil {
		t.Fatalf("GetHistoryFile() error = %v", err)
	}

	if got != customPath {
		t.Errorf("GetHistoryFile() with HISTFILE = %v, want %v", got, customPath)
	}
}

func TestBashWriter(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "bash_history")

	writer := NewBashWriter(tempFile)

	// Test writing entries
	entries := []string{
		"make build",
		"make test",
		"make deploy",
	}

	for _, entry := range entries {
		if err := writer.Append(entry); err != nil {
			t.Fatalf("Append(%q) failed: %v", entry, err)
		}
	}

	// Verify file contents
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != len(entries) {
		t.Errorf("Expected %d lines, got %d", len(entries), len(lines))
	}

	for i, want := range entries {
		if lines[i] != want {
			t.Errorf("Line %d = %q, want %q", i, lines[i], want)
		}
	}
}

func TestBashWriterConcurrent(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "bash_history")
	writer := NewBashWriter(tempFile)

	// Write concurrently to test file locking
	done := make(chan bool)
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		go func(n int) {
			entry := "make test"
			if err := writer.Append(entry); err != nil {
				t.Errorf("Concurrent Append failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify all entries were written
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != numGoroutines {
		t.Errorf("Expected %d lines after concurrent writes, got %d", numGoroutines, len(lines))
	}
}

func TestZshWriterStandardFormat(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "zsh_history")

	// Create history file with standard format
	standardHistory := "make build\nmake test\n"
	if err := os.WriteFile(tempFile, []byte(standardHistory), 0600); err != nil {
		t.Fatalf("Failed to create test history file: %v", err)
	}

	writer := NewZshWriter(tempFile)

	// Verify standard format detected
	if writer.extendedHistory {
		t.Error("Expected standard format, but extended history was detected")
	}

	// Append new entry
	if err := writer.Append("make deploy"); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify format is still standard
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	lastLine := getLastLine(string(content))
	if lastLine != "make deploy" {
		t.Errorf("Expected standard format 'make deploy', got %q", lastLine)
	}
}

func TestZshWriterExtendedFormat(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "zsh_history")

	// Create history file with extended format
	extendedHistory := ": 1234567890:0;make build\n: 1234567891:0;make test\n"
	if err := os.WriteFile(tempFile, []byte(extendedHistory), 0600); err != nil {
		t.Fatalf("Failed to create test history file: %v", err)
	}

	writer := NewZshWriter(tempFile)

	// Verify extended format detected
	if !writer.extendedHistory {
		t.Error("Expected extended format, but standard format was detected")
	}

	// Append new entry
	if err := writer.Append("make deploy"); err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	// Verify format is extended
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	lastLine := getLastLine(string(content))
	if !strings.HasPrefix(lastLine, ": ") || !strings.Contains(lastLine, ":0;make deploy") {
		t.Errorf("Expected extended format, got %q", lastLine)
	}
}

func TestZshWriterEmptyFile(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "zsh_history")

	// Create empty file
	if err := os.WriteFile(tempFile, []byte(""), 0600); err != nil {
		t.Fatalf("Failed to create test history file: %v", err)
	}

	writer := NewZshWriter(tempFile)

	// Should default to standard format for empty file
	if writer.extendedHistory {
		t.Error("Expected standard format for empty file, got extended")
	}
}

func TestNewIntegration(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name       string
		config     *Config
		wantNil    bool
		wantWriter bool
	}{
		{
			name: "enabled with bash",
			config: &Config{
				Enabled:     true,
				Shell:       "bash",
				HistoryFile: filepath.Join(tempDir, "bash_history"),
			},
			wantNil:    false,
			wantWriter: true,
		},
		{
			name: "enabled with zsh",
			config: &Config{
				Enabled:     true,
				Shell:       "zsh",
				HistoryFile: filepath.Join(tempDir, "zsh_history"),
			},
			wantNil:    false,
			wantWriter: true,
		},
		{
			name: "disabled",
			config: &Config{
				Enabled: false,
			},
			wantNil:    true,
			wantWriter: false,
		},
		{
			name:       "nil config",
			config:     nil,
			wantNil:    true,
			wantWriter: false,
		},
		{
			name: "unsupported shell",
			config: &Config{
				Enabled: true,
				Shell:   "powershell",
			},
			wantNil:    true,
			wantWriter: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			integ, err := NewIntegration(tt.config)

			if tt.wantNil && integ != nil {
				t.Errorf("Expected nil integration, got %v", integ)
			}

			if !tt.wantNil && integ == nil {
				t.Errorf("Expected non-nil integration, got nil")
			}

			if err != nil {
				t.Logf("NewIntegration error (may be expected): %v", err)
			}

			if integ != nil && tt.wantWriter && integ.writer == nil {
				t.Error("Expected writer to be initialized, got nil")
			}
		})
	}
}

func TestRecordExecution(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "bash_history")

	config := &Config{
		Enabled:        true,
		Shell:          "bash",
		HistoryFile:    tempFile,
		FormatTemplate: "make {target}",
	}

	integ, err := NewIntegration(config)
	if err != nil {
		t.Fatalf("Failed to create integration: %v", err)
	}

	// Record execution
	if err := integ.RecordExecution("build"); err != nil {
		t.Fatalf("RecordExecution failed: %v", err)
	}

	// Verify entry was written
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	if !strings.Contains(string(content), "make build") {
		t.Errorf("Expected 'make build' in history, got: %s", string(content))
	}
}

func TestRecordExecutionWithExclusions(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "bash_history")

	config := &Config{
		Enabled:        true,
		Shell:          "bash",
		HistoryFile:    tempFile,
		ExcludeTargets: []string{"help", "list"},
	}

	integ, err := NewIntegration(config)
	if err != nil {
		t.Fatalf("Failed to create integration: %v", err)
	}

	// Record excluded target
	if err := integ.RecordExecution("help"); err != nil {
		t.Fatalf("RecordExecution failed: %v", err)
	}

	// Verify no entry was written
	content, err := os.ReadFile(tempFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("Failed to read history file: %v", err)
	}

	if strings.Contains(string(content), "make help") {
		t.Errorf("Excluded target 'help' should not be in history")
	}

	// Record non-excluded target
	if err := integ.RecordExecution("build"); err != nil {
		t.Fatalf("RecordExecution failed: %v", err)
	}

	// Verify entry was written
	content2, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read history file: %v", err)
	}

	if !strings.Contains(string(content2), "make build") {
		t.Errorf("Expected 'make build' in history")
	}
}

func TestFormatEntry(t *testing.T) {
	tests := []struct {
		name     string
		template string
		target   string
		want     string
	}{
		{
			name:     "default template",
			template: "",
			target:   "build",
			want:     "make build",
		},
		{
			name:     "simple template",
			template: "make {target}",
			target:   "test",
			want:     "make test",
		},
		{
			name:     "custom template",
			template: "run make {target}",
			target:   "deploy",
			want:     "run make deploy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatEntry(tt.template, tt.target)
			if got != tt.want {
				t.Errorf("formatEntry(%q, %q) = %q, want %q", tt.template, tt.target, got, tt.want)
			}
		})
	}
}

func TestAutoDetection(t *testing.T) {
	tempDir := t.TempDir()

	// Save original SHELL env
	originalShell := os.Getenv("SHELL")
	defer os.Setenv("SHELL", originalShell)

	// Test auto-detection with bash
	os.Setenv("SHELL", "/bin/bash")

	config := &Config{
		Enabled:     true,
		Shell:       "auto",
		HistoryFile: filepath.Join(tempDir, "auto_history"),
	}

	integ, err := NewIntegration(config)
	if err != nil {
		t.Fatalf("Failed to create integration: %v", err)
	}

	if integ == nil {
		t.Fatal("Expected non-nil integration with auto detection")
	}

	// Verify writer is bash writer
	if _, ok := integ.writer.(*BashWriter); !ok {
		t.Errorf("Expected BashWriter with auto-detection of bash, got %T", integ.writer)
	}
}

// Helper function to get last non-empty line from string
func getLastLine(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	if len(lines) == 0 {
		return ""
	}
	return lines[len(lines)-1]
}

// Test edge case: file permissions
func TestFilePermissions(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "bash_history")

	writer := NewBashWriter(tempFile)

	// Write initial entry
	if err := writer.Append("make build"); err != nil {
		t.Fatalf("Initial append failed: %v", err)
	}

	// Verify file has correct permissions (0600)
	info, err := os.Stat(tempFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := info.Mode().Perm()
	expectedMode := os.FileMode(0600)
	if mode != expectedMode {
		t.Errorf("Expected file permissions %v, got %v", expectedMode, mode)
	}
}

// Test detectZshExtendedHistory edge cases
func TestDetectZshExtendedHistory(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "extended history",
			content: ": 1234567890:0;make build\n: 1234567891:2;make test\n",
			want:    true,
		},
		{
			name:    "standard history",
			content: "make build\nmake test\n",
			want:    false,
		},
		{
			name:    "mixed format prefers extended",
			content: ": 1234567890:0;make build\nmake test\n",
			want:    true,
		},
		{
			name:    "empty file",
			content: "",
			want:    false,
		},
		{
			name:    "single standard line",
			content: "make build\n",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempFile := filepath.Join(t.TempDir(), "zsh_history")
			if err := os.WriteFile(tempFile, []byte(tt.content), 0600); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			got := detectZshExtendedHistory(tempFile)
			if got != tt.want {
				t.Errorf("detectZshExtendedHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test non-existent file handling
func TestNonExistentFile(t *testing.T) {
	tempFile := filepath.Join(t.TempDir(), "nonexistent", "bash_history")

	writer := NewBashWriter(tempFile)

	// Should create parent directory and file
	if err := writer.Append("make build"); err != nil {
		t.Fatalf("Append to non-existent file failed: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(tempFile); err != nil {
		t.Errorf("File should have been created: %v", err)
	}
}

func TestReadmeExample(t *testing.T) {
	// This test verifies the example from README.md works
	tempFile := filepath.Join(t.TempDir(), "bash_history")

	config := &Config{
		Enabled:        true,
		Shell:          "bash",
		HistoryFile:    tempFile,
		FormatTemplate: "make {target}",
	}

	integ, err := NewIntegration(config)
	if err != nil {
		t.Fatalf("Failed to create integration: %v", err)
	}

	// Simulate running 'build' and 'test' targets
	targets := []string{"build", "test", "build", "lint"}
	for _, target := range targets {
		if err := integ.RecordExecution(target); err != nil {
			t.Fatalf("RecordExecution(%q) failed: %v", target, err)
		}
	}

	// Read history file
	f, err := os.Open(tempFile)
	if err != nil {
		t.Fatalf("Failed to open history file: %v", err)
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Verify all entries are present
	expected := []string{"make build", "make test", "make build", "make lint"}
	if len(lines) != len(expected) {
		t.Errorf("Expected %d history entries, got %d", len(expected), len(lines))
	}

	for i, want := range expected {
		if i < len(lines) && lines[i] != want {
			t.Errorf("Line %d: got %q, want %q", i, lines[i], want)
		}
	}
}
