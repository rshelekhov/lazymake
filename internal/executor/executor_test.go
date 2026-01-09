package executor

import (
	"os"
	"testing"
	"time"
)

func TestExecuteSuccess(t *testing.T) {
	// Create a simple test Makefile
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: test
test:
	@echo "test successful"
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	// Change to temp directory
	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Execute the target
	result := Execute("test", makefile)

	// Verify success
	if result.Err != nil {
		t.Errorf("Expected no error, got: %v", result.Err)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got: %d", result.ExitCode)
	}

	if result.Output == "" {
		t.Error("Expected non-empty output")
	}

	if result.Output[0:len(result.Output)-1] != "test successful" {
		t.Logf("Output: %q", result.Output)
	}
}

func TestExecuteFailure(t *testing.T) {
	// Create a Makefile with a failing target
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: fail
fail:
	@exit 42
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Execute the failing target
	result := Execute("fail", makefile)

	// Verify failure
	if result.Err == nil {
		t.Error("Expected error, got nil")
	}

	// Make returns exit code 2 when a recipe fails (not the recipe's exit code)
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code, got: %d", result.ExitCode)
	}
}

func TestExecuteNonExistentTarget(t *testing.T) {
	// Create an empty Makefile
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := ``
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Execute non-existent target
	result := Execute("nonexistent", makefile)

	// Verify error
	if result.Err == nil {
		t.Error("Expected error for non-existent target, got nil")
	}

	// Exit code should be non-zero (make returns 2 for "no rule to make target")
	if result.ExitCode == 0 {
		t.Errorf("Expected non-zero exit code, got: %d", result.ExitCode)
	}

	// Should have error message in output
	if result.Output == "" {
		t.Error("Expected error message in output")
	}
}

func TestExecuteWithOutput(t *testing.T) {
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: echo
echo:
	@echo "line 1"
	@echo "line 2"
	@echo "line 3"
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	result := Execute("echo", makefile)

	// Verify output contains all lines
	expectedLines := []string{"line 1", "line 2", "line 3"}
	for _, line := range expectedLines {
		if !contains(result.Output, line) {
			t.Errorf("Expected output to contain %q, got: %q", line, result.Output)
		}
	}
}

func TestExecuteTiming(t *testing.T) {
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	// Create a target that sleeps for a known duration
	makefileContent := `
.PHONY: slow
slow:
	@sleep 0.1
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	result := Execute("slow", makefile)

	// Verify timing fields are set
	if result.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}

	if result.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}

	if result.Duration == 0 {
		t.Error("Expected Duration to be non-zero")
	}

	// Verify EndTime is after StartTime
	if !result.EndTime.After(result.StartTime) {
		t.Error("Expected EndTime to be after StartTime")
	}

	// Verify Duration matches EndTime - StartTime
	expectedDuration := result.EndTime.Sub(result.StartTime)
	if result.Duration != expectedDuration {
		t.Errorf("Duration mismatch: got %v, calculated %v", result.Duration, expectedDuration)
	}

	// Verify execution took at least 100ms (sleep time)
	if result.Duration < 100*time.Millisecond {
		t.Errorf("Expected duration >= 100ms, got: %v", result.Duration)
	}
}

func TestExecuteExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		makefileBody string
		wantExitCode int
		wantError    bool
	}{
		{
			name:   "exit code 0",
			target: "success",
			makefileBody: `
success:
	@exit 0
`,
			wantExitCode: 0,
			wantError:    false,
		},
		{
			name:   "exit code 1",
			target: "fail1",
			makefileBody: `
fail1:
	@exit 1
`,
			wantExitCode: 2, // Make returns 2 for failed recipes
			wantError:    true,
		},
		{
			name:   "exit code 127",
			target: "fail127",
			makefileBody: `
fail127:
	@exit 127
`,
			wantExitCode: 2, // Make returns 2 for failed recipes
			wantError:    true,
		},
		{
			name:   "exit code 255",
			target: "fail255",
			makefileBody: `
fail255:
	@exit 255
`,
			wantExitCode: 2, // Make returns 2 for failed recipes
			wantError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			makefile := tempDir + "/Makefile"

			if err := os.WriteFile(makefile, []byte(tt.makefileBody), 0644); err != nil {
				t.Fatalf("Failed to create test Makefile: %v", err)
			}

			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tempDir)

			result := Execute(tt.target, makefile)

			if (result.Err != nil) != tt.wantError {
				t.Errorf("Execute() error = %v, wantError %v", result.Err, tt.wantError)
			}

			if result.ExitCode != tt.wantExitCode {
				t.Errorf("Execute() exit code = %d, want %d", result.ExitCode, tt.wantExitCode)
			}
		})
	}
}

func TestExecuteStderrAndStdout(t *testing.T) {
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: mixed
mixed:
	@echo "stdout message"
	@echo "stderr message" >&2
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	result := Execute("mixed", makefile)

	// CombinedOutput should contain both stdout and stderr
	if !contains(result.Output, "stdout message") {
		t.Error("Expected output to contain stdout message")
	}

	if !contains(result.Output, "stderr message") {
		t.Error("Expected output to contain stderr message")
	}
}

func TestExecuteResultStructure(t *testing.T) {
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: test
test:
	@echo "hello"
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	result := Execute("test", makefile)

	// Verify all fields in Result struct are populated correctly
	if result.Output == "" {
		t.Error("Expected Output to be non-empty")
	}

	// For successful execution, Err should be nil
	if result.Err != nil {
		t.Errorf("Expected Err to be nil for successful execution, got: %v", result.Err)
	}

	if result.Duration <= 0 {
		t.Errorf("Expected Duration > 0, got: %v", result.Duration)
	}

	if result.ExitCode != 0 {
		t.Errorf("Expected ExitCode 0, got: %d", result.ExitCode)
	}

	if result.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}

	if result.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}

	// Verify chronological order
	if result.StartTime.After(result.EndTime) {
		t.Error("StartTime should be before EndTime")
	}
}

func TestExecuteCommandNotFound(t *testing.T) {
	// This simulates a case where make command itself might not be found
	// or some other non-exit error occurs
	// We can't easily test this without modifying PATH, so we'll skip
	// or test indirectly through the exit code -1 handling

	t.Skip("Command not found scenario is difficult to test reliably")
}

func TestExecuteMultipleTargets(t *testing.T) {
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: one two three
one:
	@echo "one"
two:
	@echo "two"
three:
	@echo "three"
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Test executing different targets sequentially
	targets := []string{"one", "two", "three"}
	for _, target := range targets {
		result := Execute(target, makefile)

		if result.Err != nil {
			t.Errorf("Execute(%q) error: %v", target, result.Err)
		}

		if !contains(result.Output, target) {
			t.Errorf("Execute(%q) output should contain %q, got: %q", target, target, result.Output)
		}
	}
}

func TestExecuteConcurrent(t *testing.T) {
	tempDir := t.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: concurrent
concurrent:
	@echo "concurrent execution"
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		t.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	// Execute same target concurrently
	done := make(chan Result, 5)

	for i := 0; i < 5; i++ {
		go func() {
			result := Execute("concurrent", makefile)
			done <- result
		}()
	}

	// Collect results
	for i := 0; i < 5; i++ {
		result := <-done
		if result.Err != nil {
			t.Errorf("Concurrent execution %d failed: %v", i, result.Err)
		}
		if result.ExitCode != 0 {
			t.Errorf("Concurrent execution %d exit code: %d", i, result.ExitCode)
		}
	}
}

func TestExecute_CommandNotFound_SetsAllFields(t *testing.T) {
	// This test simulates an unreachable make binary, by temporarily altering PATH
	oldPath := os.Getenv("PATH")
	defer os.Setenv("PATH", oldPath)
	os.Setenv("PATH", "/nonexistent") // Should fail to find 'make'

	start := time.Now()
	result := Execute("anytarget", "makefile")
	end := time.Now()

	if result.Err == nil {
		t.Error("Expected error when make binary is not found")
	}
	if result.ExitCode != -1 {
		t.Errorf("Expected exit code -1 for not found, got %d", result.ExitCode)
	}
	if result.StartTime.IsZero() {
		t.Error("Expected StartTime to be set")
	}
	if result.EndTime.IsZero() {
		t.Error("Expected EndTime to be set")
	}
	if !result.EndTime.After(result.StartTime) && result.StartTime != result.EndTime {
		t.Errorf("EndTime (got %v) should be after StartTime (got %v)", result.EndTime, result.StartTime)
	}
	if result.Duration <= 0 {
		t.Errorf("Expected Duration to be positive, got %v", result.Duration)
	}
	if result.StartTime.Before(start) || result.EndTime.After(end) {
		t.Errorf("Result times out of bounds: start=%v, end=%v, got StartTime=%v, EndTime=%v", start, end, result.StartTime, result.EndTime)
	}
	// Output may be empty for command not found on some systems, so don't check it
}

func TestExecute_ResultFields_AlwaysSetEvenOnFailure(t *testing.T) {
	// Test various error scenarios: fail exit, missing target, command not found

	cases := []struct {
		name      string
		target    string
		mkContent string
		exitCheck func(int) bool
		shouldErr bool
	}{
		{
			name:      "fail_exit_7",
			target:    "fail7",
			mkContent: ".PHONY: fail7\nfail7:\n\texit 7\n",
			exitCheck: func(code int) bool { return code == 2 }, // Make returns 2 for failed recipes
			shouldErr: true,
		},
		{
			name:      "missing_target",
			target:    "nope",
			mkContent: ".PHONY:\n",
			exitCheck: func(code int) bool { return code != 0 },
			shouldErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tempDir := t.TempDir()
			f := tempDir + "/Makefile"
			if err := os.WriteFile(f, []byte(tc.mkContent), 0644); err != nil {
				t.Fatalf("makefile write: %v", err)
			}
			oldDir, _ := os.Getwd()
			defer os.Chdir(oldDir)
			os.Chdir(tempDir)

			res := Execute(tc.target, f)

			if tc.shouldErr && res.Err == nil {
				t.Error("expected error, got nil")
			}
			if !tc.exitCheck(res.ExitCode) {
				t.Errorf("exit code failed check: got %d", res.ExitCode)
			}
			if res.StartTime.IsZero() {
				t.Error("StartTime should be set")
			}
			if res.EndTime.IsZero() {
				t.Error("EndTime should be set")
			}
			if !res.EndTime.After(res.StartTime) && res.StartTime != res.EndTime {
				t.Errorf("EndTime not after StartTime: %v vs %v", res.StartTime, res.EndTime)
			}
			if res.Duration <= 0 {
				t.Errorf("Duration should be positive, got %v", res.Duration)
			}
			if res.Output == "" {
				t.Error("Expected Output to contain error message")
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Benchmark test
func BenchmarkExecute(b *testing.B) {
	tempDir := b.TempDir()
	makefile := tempDir + "/Makefile"

	makefileContent := `
.PHONY: bench
bench:
	@echo "benchmark"
`
	if err := os.WriteFile(makefile, []byte(makefileContent), 0644); err != nil {
		b.Fatalf("Failed to create test Makefile: %v", err)
	}

	oldDir, _ := os.Getwd()
	defer os.Chdir(oldDir)
	os.Chdir(tempDir)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Execute("bench", makefile)
	}
}
