package executor

import "os/exec"

// Result holds command execution results
type Result struct {
	Output string
	Err    error
}

func Execute(target string) Result {
	cmd := exec.Command("make", target)
	output, err := cmd.CombinedOutput()

	return Result{
		Output: string(output),
		Err:    err,
	}
}
