package export

import (
	"fmt"
	"os"
	"os/user"
	"sort"
	"strings"
	"time"

	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/version"
)

// formatEnvForLog renders env pairs deterministically as
// "KEY=val KEY2=val2", values quoted if they contain whitespace.
// Used by FormatLog so consecutive runs with the same env produce
// identical strings.
func formatEnvForLog(env map[string]string) string {
	if len(env) == 0 {
		return ""
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		v := env[k]
		if strings.ContainsAny(v, " \t") {
			v = `"` + v + `"`
		}
		parts = append(parts, k+"="+v)
	}
	return strings.Join(parts, " ")
}

// ExecutionRecord represents a complete execution result for export
type ExecutionRecord struct {
	// Execution metadata
	Timestamp    time.Time `json:"timestamp"`
	MakefilePath string    `json:"makefile_path"`
	TargetName   string    `json:"target_name"`

	// Timing data
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	DurationMs int64         `json:"duration_ms"` // Human-friendly milliseconds

	// Execution results
	Success      bool   `json:"success"`
	ExitCode     int    `json:"exit_code"`
	Output       string `json:"output"`
	ErrorMessage string `json:"error_message,omitempty"`

	// Run parameters supplied via the interactive form (issue #37).
	// Omitted from JSON when empty so plain runs serialize identically
	// to previous versions of the format.
	Env   map[string]string `json:"env,omitempty"`
	Flags []string          `json:"flags,omitempty"`

	// Environment context
	WorkingDir      string `json:"working_dir"`
	User            string `json:"user,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	LazymakeVersion string `json:"lazymake_version,omitempty"`
}

// NewExecutionRecord creates an ExecutionRecord from execution data.
// Env and flags default to empty; use NewExecutionRecordWithParams
// to record interactive-form parameters (issue #37).
func NewExecutionRecord(makefilePath, targetName string, result executor.Result) *ExecutionRecord {
	return NewExecutionRecordWithParams(makefilePath, targetName, result, nil, nil)
}

// NewExecutionRecordWithParams is the full-fidelity factory. env and
// flags may be nil for plain runs; both are emitted in JSON output
// only when non-empty.
func NewExecutionRecordWithParams(makefilePath, targetName string, result executor.Result, env map[string]string, flags []string) *ExecutionRecord {
	// Get working directory
	workingDir, _ := os.Getwd()

	// Get current user
	currentUser := ""
	if u, err := user.Current(); err == nil {
		currentUser = u.Username
	}

	// Get hostname
	hostname, _ := os.Hostname()

	// Extract error message if present
	errMsg := ""
	if result.Err != nil {
		errMsg = result.Err.Error()
	}

	return &ExecutionRecord{
		Timestamp:       result.EndTime,
		MakefilePath:    makefilePath,
		TargetName:      targetName,
		StartTime:       result.StartTime,
		EndTime:         result.EndTime,
		Duration:        result.Duration,
		DurationMs:      result.Duration.Milliseconds(),
		Success:         result.Err == nil,
		ExitCode:        result.ExitCode,
		Output:          result.Output,
		ErrorMessage:    errMsg,
		Env:             env,
		Flags:           flags,
		WorkingDir:      workingDir,
		User:            currentUser,
		Hostname:        hostname,
		LazymakeVersion: version.Version,
	}
}

// FormatLog formats the execution record as a human-readable log
func (r *ExecutionRecord) FormatLog() string {
	var b strings.Builder

	// Header
	b.WriteString(strings.Repeat("=", 80))
	b.WriteString("\nLazymake Execution Log\n")
	b.WriteString(strings.Repeat("=", 80))
	b.WriteString("\n")

	// Metadata
	fmt.Fprintf(&b, "Target:        %s\n", r.TargetName)
	fmt.Fprintf(&b, "Makefile:      %s\n", r.MakefilePath)
	fmt.Fprintf(&b, "Timestamp:     %s\n", r.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(&b, "Duration:      %.3fs\n", r.Duration.Seconds())
	fmt.Fprintf(&b, "Exit Code:     %d\n", r.ExitCode)

	status := "SUCCESS"
	if !r.Success {
		status = "FAILED"
	}
	fmt.Fprintf(&b, "Status:        %s\n", status)

	fmt.Fprintf(&b, "Working Dir:   %s\n", r.WorkingDir)
	if r.User != "" {
		fmt.Fprintf(&b, "User:          %s\n", r.User)
	}
	if r.Hostname != "" {
		fmt.Fprintf(&b, "Host:          %s\n", r.Hostname)
	}

	// Interactive run parameters, when present (issue #37).
	if len(r.Env) > 0 {
		fmt.Fprintf(&b, "Env:           %s\n", formatEnvForLog(r.Env))
	}
	if len(r.Flags) > 0 {
		fmt.Fprintf(&b, "Flags:         %s\n", strings.Join(r.Flags, " "))
	}

	// Output section
	b.WriteString(strings.Repeat("=", 80))
	b.WriteString("\n\nOUTPUT:\n")
	b.WriteString(r.Output)
	if !strings.HasSuffix(r.Output, "\n") {
		b.WriteString("\n")
	}

	// Footer
	b.WriteString("\n")
	b.WriteString(strings.Repeat("=", 80))
	b.WriteString("\n")
	if r.Success {
		fmt.Fprintf(&b, "Execution completed successfully in %.3fs\n", r.Duration.Seconds())
	} else {
		fmt.Fprintf(&b, "Execution failed after %.3fs\n", r.Duration.Seconds())
		if r.ErrorMessage != "" {
			fmt.Fprintf(&b, "Error: %s\n", r.ErrorMessage)
		}
	}
	b.WriteString(strings.Repeat("=", 80))
	b.WriteString("\n")

	return b.String()
}

// GenerateFilename generates a filename based on the naming strategy
func (r *ExecutionRecord) GenerateFilename(strategy string, extension string) string {
	// Sanitize target name for filesystem
	sanitized := strings.ReplaceAll(r.TargetName, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	switch strategy {
	case "target":
		// Overwrite previous for same target
		return fmt.Sprintf("%s_latest.%s", sanitized, extension)

	case "sequential":
		// Sequential numbering handled by exporter
		return fmt.Sprintf("%s.%s", sanitized, extension)

	case "timestamp":
		fallthrough
	default:
		// Timestamp-based naming
		timestamp := r.Timestamp.Format("20060102_150405")
		return fmt.Sprintf("%s_%s.%s", sanitized, timestamp, extension)
	}
}
