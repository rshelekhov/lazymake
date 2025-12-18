package export

import (
	"fmt"
	"os"
	"os/user"
	"strings"
	"time"

	"github.com/rshelekhov/lazymake/internal/executor"
	"github.com/rshelekhov/lazymake/version"
)

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

	// Environment context
	WorkingDir      string `json:"working_dir"`
	User            string `json:"user,omitempty"`
	Hostname        string `json:"hostname,omitempty"`
	LazymakeVersion string `json:"lazymake_version,omitempty"`
}

// NewExecutionRecord creates an ExecutionRecord from execution data
func NewExecutionRecord(makefilePath, targetName string, result executor.Result) *ExecutionRecord {
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
	b.WriteString(fmt.Sprintf("Target:        %s\n", r.TargetName))
	b.WriteString(fmt.Sprintf("Makefile:      %s\n", r.MakefilePath))
	b.WriteString(fmt.Sprintf("Timestamp:     %s\n", r.Timestamp.Format("2006-01-02 15:04:05")))
	b.WriteString(fmt.Sprintf("Duration:      %.3fs\n", r.Duration.Seconds()))
	b.WriteString(fmt.Sprintf("Exit Code:     %d\n", r.ExitCode))

	status := "SUCCESS"
	if !r.Success {
		status = "FAILED"
	}
	b.WriteString(fmt.Sprintf("Status:        %s\n", status))

	b.WriteString(fmt.Sprintf("Working Dir:   %s\n", r.WorkingDir))
	if r.User != "" {
		b.WriteString(fmt.Sprintf("User:          %s\n", r.User))
	}
	if r.Hostname != "" {
		b.WriteString(fmt.Sprintf("Host:          %s\n", r.Hostname))
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
		b.WriteString(fmt.Sprintf("Execution completed successfully in %.3fs\n", r.Duration.Seconds()))
	} else {
		b.WriteString(fmt.Sprintf("Execution failed after %.3fs\n", r.Duration.Seconds()))
		if r.ErrorMessage != "" {
			b.WriteString(fmt.Sprintf("Error: %s\n", r.ErrorMessage))
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
