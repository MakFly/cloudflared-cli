package cloudflared

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"
)

// CommandResult holds the output of a cloudflared command execution.
type CommandResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// Executor defines the interface for running cloudflared commands.
// This interface enables mocking in tests.
type Executor interface {
	// RunCommand executes a cloudflared command and captures output.
	// It applies a default timeout of 30 seconds.
	RunCommand(ctx context.Context, args ...string) (CommandResult, error)

	// RunAttached executes a cloudflared command with stdin/stdout/stderr
	// connected directly to the terminal. Used for long-running commands
	// like `tunnel run`.
	RunAttached(ctx context.Context, args ...string) error
}

// RealExecutor implements Executor using the actual cloudflared binary.
type RealExecutor struct {
	BinaryPath string
	Timeout    time.Duration
}

// NewExecutor creates a new RealExecutor with the given binary path.
func NewExecutor(binaryPath string) *RealExecutor {
	return &RealExecutor{
		BinaryPath: binaryPath,
		Timeout:    30 * time.Second,
	}
}

func (e *RealExecutor) RunCommand(ctx context.Context, args ...string) (CommandResult, error) {
	timeout := e.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, e.BinaryPath, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	result := CommandResult{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}

	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}

	if ctx.Err() == context.DeadlineExceeded {
		return result, fmt.Errorf("command timed out after %s", timeout)
	}

	return result, err
}

func (e *RealExecutor) RunAttached(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, e.BinaryPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunDetached starts a cloudflared command in the background and returns
// the process. The caller is responsible for managing the process lifecycle.
func (e *RealExecutor) RunDetached(ctx context.Context, logWriter io.Writer, args ...string) (*os.Process, error) {
	cmd := exec.CommandContext(ctx, e.BinaryPath, args...)
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start cloudflared: %w", err)
	}

	return cmd.Process, nil
}
