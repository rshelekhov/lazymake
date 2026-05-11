package executor

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"sort"
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

// ExecutionOptions controls how ExecuteStreaming invokes `make`.
//
// All fields are optional. The zero value is equivalent to a normal,
// no-extra-env, no-extra-flags execution.
type ExecutionOptions struct {
	// DryRun adds `-n` (`--just-print`) so make prints recipe lines
	// without actually running them.
	DryRun bool

	// Env are additional environment variables applied on top of
	// os.Environ() in the spawned process. Make picks them up the same
	// way as a regular shell would. Keys are sorted before applying
	// so command construction is deterministic.
	Env map[string]string

	// Flags are extra arguments inserted between the makefile path and
	// the target, e.g. ["-j4", "-k"]. They are passed through exec
	// argv (not the shell), so no escaping is performed and no shell
	// metacharacters are interpreted.
	Flags []string
}

// buildArgs constructs the argv (excluding the leading "make") for the
// given options and target. Order is:
//
//	-f <makefilePath> [-n] [user flags...] <target>
//
// `-n` is injected before user flags so users can still override or
// extend dry-run with additional flags like `-j1`.
func buildArgs(target, makefilePath string, opts ExecutionOptions) []string {
	args := make([]string, 0, 3+len(opts.Flags))
	args = append(args, "-f", makefilePath)
	if opts.DryRun {
		// `-n` (--just-print / --dry-run): print recipe commands without executing them.
		// MAKEFLAGS propagates -n to recursive sub-makes automatically.
		args = append(args, "-n")
	}
	args = append(args, opts.Flags...)
	args = append(args, target)
	return args
}

// buildEnv returns the environment for the spawned process: os.Environ()
// plus opts.Env, with deterministic ordering (keys sorted). Returns nil
// when opts.Env is empty so cmd.Env can stay at the os/exec default.
func buildEnv(opts ExecutionOptions) []string {
	if len(opts.Env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(opts.Env))
	for k := range opts.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	env := os.Environ()
	for _, k := range keys {
		env = append(env, k+"="+opts.Env[k])
	}
	return env
}

// ExecuteStreaming runs a make target and streams output via channel.
//
// Options control dry-run mode, additional environment variables, and
// extra make flags. The command is invoked via exec argv (no shell), so
// values are not subject to shell interpretation.
//
// Returns: channel for output chunks, cancel function.
func ExecuteStreaming(target, makefilePath string, opts ExecutionOptions) (<-chan OutputChunk, func()) {
	chunks := make(chan OutputChunk, 100)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(chunks)

		args := buildArgs(target, makefilePath, opts)
		cmd := exec.CommandContext(ctx, "make", args...)
		if env := buildEnv(opts); env != nil {
			cmd.Env = env
		}

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
