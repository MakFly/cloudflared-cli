package runner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
)

// PIDFile manages the tunnel process PID file.
const pidFileName = "tunnel.pid"

// Runner manages the cloudflared tunnel process lifecycle.
type Runner struct {
	Executor   cloudflared.Executor
	ProjectDir string
	Env        string
}

// NewRunner creates a new Runner.
func NewRunner(executor cloudflared.Executor, projectDir, env string) *Runner {
	return &Runner{
		Executor:   executor,
		ProjectDir: projectDir,
		Env:        env,
	}
}

// RunForeground starts the tunnel in the foreground, blocking until the context
// is cancelled or the process exits.
func (r *Runner) RunForeground(ctx context.Context) error {
	configPath := config.EnvConfigPath(r.ProjectDir, r.Env)
	return r.Executor.RunAttached(ctx, "tunnel", "--config", configPath, "run")
}

// RunDetached starts the tunnel in the background, writing a PID file.
func (r *Runner) RunDetached(ctx context.Context) (int, error) {
	configPath := config.EnvConfigPath(r.ProjectDir, r.Env)

	logDir := filepath.Join(r.ProjectDir, "logs")
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return 0, fmt.Errorf("failed to create logs directory: %w", err)
	}

	logFile, err := os.OpenFile(
		filepath.Join(logDir, r.Env+".log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0o644,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to open log file: %w", err)
	}

	exec, ok := r.Executor.(*cloudflared.RealExecutor)
	if !ok {
		return 0, fmt.Errorf("detached mode requires a real executor")
	}

	proc, err := exec.RunDetached(ctx, logFile, "tunnel", "--config", configPath, "run")
	if err != nil {
		logFile.Close()
		return 0, err
	}

	pid := proc.Pid
	if err := r.writePID(pid); err != nil {
		ui.Warn("Failed to write PID file: %s", err)
	}

	return pid, nil
}

// Stop stops the running tunnel by reading the PID file and sending SIGTERM.
func (r *Runner) Stop() error {
	pid, err := r.readPID()
	if err != nil {
		return fmt.Errorf("no running tunnel found: %w", err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		r.cleanPID()
		return fmt.Errorf("process %d not found", pid)
	}

	if err := proc.Signal(os.Interrupt); err != nil {
		r.cleanPID()
		return fmt.Errorf("failed to stop process %d: %w", pid, err)
	}

	r.cleanPID()
	return nil
}

// IsRunning checks if a tunnel process is currently running.
func (r *Runner) IsRunning() (bool, int) {
	pid, err := r.readPID()
	if err != nil {
		return false, 0
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		r.cleanPID()
		return false, 0
	}

	// On Unix, FindProcess always succeeds. Check if process is alive.
	if err := proc.Signal(os.Signal(nil)); err != nil {
		r.cleanPID()
		return false, 0
	}

	return true, pid
}

func (r *Runner) pidPath() string {
	return filepath.Join(r.ProjectDir, pidFileName)
}

func (r *Runner) writePID(pid int) error {
	return os.WriteFile(r.pidPath(), []byte(strconv.Itoa(pid)), 0o644)
}

func (r *Runner) readPID() (int, error) {
	data, err := os.ReadFile(r.pidPath())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func (r *Runner) cleanPID() {
	os.Remove(r.pidPath())
}

// RouteDNS creates a DNS route for the tunnel.
func (r *Runner) RouteDNS(ctx context.Context, hostname string) error {
	cfg, err := config.LoadTunnelConfig(r.ProjectDir, r.Env)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	result, err := r.Executor.RunCommand(ctx, "tunnel", "route", "dns", cfg.Tunnel, hostname)
	if err != nil {
		return fmt.Errorf("failed to route DNS: %s\n%s", err, result.Stderr)
	}

	return nil
}
