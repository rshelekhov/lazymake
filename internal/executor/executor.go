package executor

import (
	"os/exec"
	"time"
)

// Result holds command execution results
type Result struct {
	Output    string
	Err       error
	Duration  time.Duration
	ExitCode  int       // Exit code from command (0 = success, non-zero = failure, -1 = error)
	StartTime time.Time // When execution started
	EndTime   time.Time // When execution ended
}

func Execute(target, makefilePath string) Result {
	start := time.Now()
	cmd := exec.Command("make", "-f", makefilePath, target)
	output, err := cmd.CombinedOutput()
	end := time.Now()
	duration := end.Sub(start)

	// Extract exit code from error
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			// Non-exit error (e.g., command not found)
			exitCode = -1
		}
	}

	return Result{
		Output:    string(output),
		Err:       err,
		Duration:  duration,
		ExitCode:  exitCode,
		StartTime: start,
		EndTime:   end,
	}
}
