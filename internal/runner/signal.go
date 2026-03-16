package runner

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/kev/cloudflared-cli/internal/ui"
)

// WithGracefulShutdown returns a context that cancels on SIGINT or SIGTERM.
// It handles double-signal for force quit.
func WithGracefulShutdown(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(sigCh)
		select {
		case <-sigCh:
			ui.Warn("Shutting down gracefully... (press Ctrl+C again to force)")
			cancel()
			// Wait for second signal for force quit
			select {
			case <-sigCh:
				ui.Error("Force shutdown")
				os.Exit(1)
			case <-ctx.Done():
			}
		case <-ctx.Done():
		}
	}()

	return ctx, cancel
}
