package export

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rshelekhov/lazymake/internal/executor"
)

func TestNewExporter(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Enabled:        true,
				Format:         "json",
				NamingStrategy: "timestamp",
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			config: &Config{
				Enabled:        true,
				Format:         "invalid",
				NamingStrategy: "timestamp",
			},
			wantErr: true,
		},
		{
			name: "invalid naming strategy",
			config: &Config{
				Enabled:        true,
				Format:         "json",
				NamingStrategy: "invalid",
			},
			wantErr: true,
		},
		{
			name:    "nil config uses defaults",
			config:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewExporter(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExporter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExportJSON(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		Enabled:        true,
		OutputDir:      tempDir,
		Format:         "json",
		NamingStrategy: "timestamp",
	}

	exporter, err := NewExporter(config)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	// Create test execution record
	result := executor.Result{
		Output:    "test output",
		Err:       nil,
		Duration:  time.Second,
		ExitCode:  0,
		StartTime: time.Now().Add(-time.Second),
		EndTime:   time.Now(),
	}

	record := NewExecutionRecord("/tmp/Makefile", "test", result)

	// Export the record
	err = exporter.Export(record)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file was created
	files, err := filepath.Glob(filepath.Join(tempDir, "test_*.json"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	// Verify JSON content
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	var exported ExecutionRecord
	if err := json.Unmarshal(data, &exported); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if exported.TargetName != "test" {
		t.Errorf("Expected target 'test', got '%s'", exported.TargetName)
	}
	if exported.Success != true {
		t.Errorf("Expected success=true, got %v", exported.Success)
	}
	if exported.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exported.ExitCode)
	}
}

func TestExportLog(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		Enabled:        true,
		OutputDir:      tempDir,
		Format:         "log",
		NamingStrategy: "timestamp",
	}

	exporter, err := NewExporter(config)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	result := executor.Result{
		Output:    "test output\n",
		Err:       nil,
		Duration:  time.Second * 2,
		ExitCode:  0,
		StartTime: time.Now().Add(-time.Second * 2),
		EndTime:   time.Now(),
	}

	record := NewExecutionRecord("/tmp/Makefile", "build", result)

	err = exporter.Export(record)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify log file was created
	files, err := filepath.Glob(filepath.Join(tempDir, "build_*.log"))
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(files) != 1 {
		t.Fatalf("Expected 1 file, got %d", len(files))
	}

	// Verify log content
	content, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	logStr := string(content)
	if !containsAll(logStr, []string{"build", "SUCCESS", "test output", "Exit Code:     0"}) {
		t.Errorf("Log content missing expected strings")
	}
}

func TestExportBoth(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		Enabled:        true,
		OutputDir:      tempDir,
		Format:         "both",
		NamingStrategy: "timestamp",
	}

	exporter, err := NewExporter(config)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	result := executor.Result{
		Output:    "output",
		Err:       nil,
		Duration:  time.Millisecond * 100,
		ExitCode:  0,
		StartTime: time.Now().Add(-time.Millisecond * 100),
		EndTime:   time.Now(),
	}

	record := NewExecutionRecord("/tmp/Makefile", "test", result)

	err = exporter.Export(record)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify both JSON and log files were created
	jsonFiles, _ := filepath.Glob(filepath.Join(tempDir, "test_*.json"))
	logFiles, _ := filepath.Glob(filepath.Join(tempDir, "test_*.log"))

	if len(jsonFiles) != 1 {
		t.Errorf("Expected 1 JSON file, got %d", len(jsonFiles))
	}
	if len(logFiles) != 1 {
		t.Errorf("Expected 1 log file, got %d", len(logFiles))
	}
}

func TestExportWithExclusions(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		Enabled:        true,
		OutputDir:      tempDir,
		Format:         "json",
		NamingStrategy: "timestamp",
		ExcludeTargets: []string{"watch", "dev"},
	}

	exporter, err := NewExporter(config)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	result := executor.Result{
		Output:    "output",
		Err:       nil,
		Duration:  time.Second,
		ExitCode:  0,
		StartTime: time.Now().Add(-time.Second),
		EndTime:   time.Now(),
	}

	// Try to export excluded target
	record := NewExecutionRecord("/tmp/Makefile", "watch", result)
	err = exporter.Export(record)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify no files were created
	files, _ := filepath.Glob(filepath.Join(tempDir, "watch_*.json"))
	if len(files) != 0 {
		t.Errorf("Expected no files for excluded target, got %d", len(files))
	}

	// Export non-excluded target
	record2 := NewExecutionRecord("/tmp/Makefile", "build", result)
	err = exporter.Export(record2)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file was created
	files2, _ := filepath.Glob(filepath.Join(tempDir, "build_*.json"))
	if len(files2) != 1 {
		t.Errorf("Expected 1 file for non-excluded target, got %d", len(files2))
	}
}

func TestExportSuccessOnly(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		Enabled:        true,
		OutputDir:      tempDir,
		Format:         "json",
		NamingStrategy: "timestamp",
		SuccessOnly:    true,
	}

	exporter, err := NewExporter(config)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	// Failed execution
	failedResult := executor.Result{
		Output:    "error output",
		Err:       os.ErrNotExist,
		Duration:  time.Second,
		ExitCode:  1,
		StartTime: time.Now().Add(-time.Second),
		EndTime:   time.Now(),
	}

	record := NewExecutionRecord("/tmp/Makefile", "test", failedResult)
	err = exporter.Export(record)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify no files for failed execution
	files, _ := filepath.Glob(filepath.Join(tempDir, "test_*.json"))
	if len(files) != 0 {
		t.Errorf("Expected no files for failed execution with success_only, got %d", len(files))
	}

	// Successful execution
	successResult := executor.Result{
		Output:    "success output",
		Err:       nil,
		Duration:  time.Second,
		ExitCode:  0,
		StartTime: time.Now().Add(-time.Second),
		EndTime:   time.Now(),
	}

	record2 := NewExecutionRecord("/tmp/Makefile", "test", successResult)
	err = exporter.Export(record2)
	if err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	// Verify file for successful execution
	files2, _ := filepath.Glob(filepath.Join(tempDir, "test_*.json"))
	if len(files2) != 1 {
		t.Errorf("Expected 1 file for successful execution, got %d", len(files2))
	}
}

func TestNamingStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy string
		target   string
		wantGlob string
	}{
		{
			name:     "timestamp strategy",
			strategy: "timestamp",
			target:   "build",
			wantGlob: "build_*.json",
		},
		{
			name:     "target strategy",
			strategy: "target",
			target:   "test",
			wantGlob: "test_latest.json",
		},
		{
			name:     "sequential strategy",
			strategy: "sequential",
			target:   "deploy",
			wantGlob: "deploy.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()

			config := &Config{
				Enabled:        true,
				OutputDir:      tempDir,
				Format:         "json",
				NamingStrategy: tt.strategy,
			}

			exporter, err := NewExporter(config)
			if err != nil {
				t.Fatalf("Failed to create exporter: %v", err)
			}

			result := executor.Result{
				Output:    "output",
				Err:       nil,
				Duration:  time.Second,
				ExitCode:  0,
				StartTime: time.Now().Add(-time.Second),
				EndTime:   time.Now(),
			}

			record := NewExecutionRecord("/tmp/Makefile", tt.target, result)
			err = exporter.Export(record)
			if err != nil {
				t.Fatalf("Export failed: %v", err)
			}

			// Verify file naming
			files, _ := filepath.Glob(filepath.Join(tempDir, tt.wantGlob))
			if len(files) < 1 {
				t.Errorf("Expected at least 1 file matching %s, got %d", tt.wantGlob, len(files))
			}
		})
	}
}

func TestRotation(t *testing.T) {
	tempDir := t.TempDir()

	config := &Config{
		Enabled:        true,
		OutputDir:      tempDir,
		Format:         "json",
		NamingStrategy: "timestamp",
		MaxFiles:       3,
	}

	exporter, err := NewExporter(config)
	if err != nil {
		t.Fatalf("Failed to create exporter: %v", err)
	}

	result := executor.Result{
		Output:    "output",
		Err:       nil,
		Duration:  time.Second,
		ExitCode:  0,
		StartTime: time.Now().Add(-time.Second),
		EndTime:   time.Now(),
	}

	// Create 5 files
	for i := 0; i < 5; i++ {
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
		record := NewExecutionRecord("/tmp/Makefile", "build", result)
		record.Timestamp = time.Now()
		err = exporter.Export(record)
		if err != nil {
			t.Fatalf("Export %d failed: %v", i, err)
		}
	}

	// Verify only MaxFiles remain
	files, _ := filepath.Glob(filepath.Join(tempDir, "build_*.json"))
	if len(files) > config.MaxFiles {
		t.Errorf("Expected at most %d files after rotation, got %d", config.MaxFiles, len(files))
	}
}

func TestExpandPath(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantType string // "contains" or "equals"
	}{
		{
			name:     "tilde expansion",
			path:     "~/test",
			wantType: "contains",
		},
		{
			name:     "empty path",
			path:     "",
			wantType: "equals",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPath(tt.path)
			if tt.wantType == "equals" && got != tt.path {
				t.Errorf("expandPath(%q) = %q, want %q", tt.path, got, tt.path)
			}
			if tt.wantType == "contains" && tt.path == "~/test" && got == tt.path {
				t.Errorf("expandPath(%q) should expand ~, got %q", tt.path, got)
			}
		})
	}
}

// Helper function
func containsAll(str string, substrs []string) bool {
	for _, substr := range substrs {
		found := false
		for i := 0; i <= len(str)-len(substr); i++ {
			if str[i:i+len(substr)] == substr {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
