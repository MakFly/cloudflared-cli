package cmd

import (
	"context"
	"fmt"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/runner"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var deployCmd = &cobra.Command{
	Use:   "deploy",
	Short: "Validate config, route DNS, and run the tunnel",
	Long: `Deploy validates the tunnel configuration, optionally creates DNS routes,
and starts the cloudflared tunnel in foreground or detached mode.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)
		detach, _ := cmd.Flags().GetBool("detach")
		routeDNS, _ := cmd.Flags().GetBool("route-dns")

		if !config.IsProjectInitialized(projectDir) {
			return fmt.Errorf("project not initialized, run 'cloudflared-project init' first")
		}

		// Detect cloudflared
		detect := cloudflared.Detect(cloudflaredPath)
		if !detect.Found {
			return fmt.Errorf("%s", cloudflared.InstallGuide())
		}

		if !cloudflared.IsAuthenticated() {
			return fmt.Errorf("not authenticated, run 'cloudflared-project login' first")
		}

		executor := cloudflared.NewExecutor(detect.Path)

		// Validate config
		cfg, err := config.LoadTunnelConfig(projectDir, env)
		if err != nil {
			return fmt.Errorf("failed to load config for env %q: %w", env, err)
		}

		result := config.ValidateTunnelConfig(cfg)
		if !result.IsValid() {
			ui.Error("Configuration validation failed:")
			for _, e := range result.Errors {
				ui.Warn("  %s", e)
			}
			return fmt.Errorf("fix configuration errors before deploying")
		}
		ui.Success("Configuration validated [%s]", env)

		r := runner.NewRunner(executor, projectDir, env)

		// Route DNS if requested
		if routeDNS {
			for _, rule := range cfg.Ingress {
				if rule.Hostname != "" {
					ui.Info("Routing DNS: %s → tunnel %s", rule.Hostname, cfg.Tunnel)
					if err := r.RouteDNS(context.Background(), rule.Hostname); err != nil {
						ui.Warn("DNS routing failed for %s: %s", rule.Hostname, err)
					} else {
						ui.Success("DNS routed: %s", rule.Hostname)
					}
				}
			}
		}

		// Run tunnel
		if detach {
			pid, err := r.RunDetached(context.Background())
			if err != nil {
				return fmt.Errorf("failed to start tunnel: %w", err)
			}
			ui.Success("Tunnel started in background (PID: %d) [%s]", pid, env)
			ui.Info("Use 'cloudflared-project status' to check tunnel state")
			ui.Info("Use 'cloudflared-project logs --follow' to tail logs")
			return nil
		}

		// Foreground mode with graceful shutdown
		ui.Info("Starting tunnel [%s] (Ctrl+C to stop)...", env)
		ctx, cancel := runner.WithGracefulShutdown(context.Background())
		defer cancel()

		if err := r.RunForeground(ctx); err != nil {
			if ctx.Err() != nil {
				ui.Success("Tunnel stopped gracefully")
				return nil
			}
			return fmt.Errorf("tunnel exited with error: %w", err)
		}

		return nil
	},
}

func init() {
	deployCmd.Flags().Bool("detach", false, "run tunnel in background")
	deployCmd.Flags().Bool("route-dns", false, "create DNS routes for ingress hostnames")
	rootCmd.AddCommand(deployCmd)
}
