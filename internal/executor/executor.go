package executor

import (
	"os/exec"
	"time"
)

// Result holds command execution results
type Result struct {
	Output   string
	Err      error
	Duration time.Duration
}

func Execute(target string) Result {
	start := time.Now()
	cmd := exec.Command("make", target)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	return Result{
		Output:   string(output),
		Err:      err,
		Duration: duration,
	}
}
