package executor

import (
	"bufio"
	"context"
	"io"
	"os/exec"
	"sync"
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

// OutputChunk represents a piece of streamed output
type OutputChunk struct {
	Data string
	Done bool
	Err  error
}

// ExecuteStreaming runs a make target and streams output via channel
// Returns: channel for output chunks, cancel function
func ExecuteStreaming(target, makefilePath string) (<-chan OutputChunk, func()) {
	chunks := make(chan OutputChunk, 100)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(chunks)

		cmd := exec.CommandContext(ctx, "make", "-f", makefilePath, target)

		// Create pipes for stdout and stderr
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			chunks <- OutputChunk{Done: true, Err: err}
			return
		}
		stderr, err := cmd.StderrPipe()
		if err != nil {
			chunks <- OutputChunk{Done: true, Err: err}
			return
		}

		// Start the command
		if err := cmd.Start(); err != nil {
			chunks <- OutputChunk{Done: true, Err: err}
			return
		}

		// Use a WaitGroup to wait for both readers to finish
		var wg sync.WaitGroup
		wg.Add(2)

		// Read from stdout
		go func() {
			defer wg.Done()
			readPipe(stdout, chunks, ctx)
		}()

		// Read from stderr
		go func() {
			defer wg.Done()
			readPipe(stderr, chunks, ctx)
		}()

		// Wait for both readers to finish
		wg.Wait()

		// Wait for command to complete
		err = cmd.Wait()

		// Send done message
		chunks <- OutputChunk{Done: true, Err: err}
	}()

	return chunks, cancel
}

// readPipe reads from a pipe and sends chunks to the channel
func readPipe(pipe io.Reader, chunks chan<- OutputChunk, ctx context.Context) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return
		default:
			chunks <- OutputChunk{Data: scanner.Text() + "\n"}
		}
	}
}
