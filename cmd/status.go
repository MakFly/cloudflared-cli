package cmd

import (
	"fmt"
	"time"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/runner"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show tunnel status",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)
		watch, _ := cmd.Flags().GetBool("watch")

		if !config.IsProjectInitialized(projectDir) {
			return fmt.Errorf("project not initialized, run 'cloudflared-project init' first")
		}

		detect := cloudflared.Detect(cloudflaredPath)
		executor := cloudflared.NewExecutor(detect.Path)

		r := runner.NewRunner(executor, projectDir, env)

		printStatus := func() {
			ui.Bold("Tunnel Status [%s]", env)
			fmt.Println()

			// Check local process
			running, pid := r.IsRunning()
			if running {
				ui.KeyValue("Process:", fmt.Sprintf("running (PID %d)", pid))
			} else {
				ui.KeyValue("Process:", "not running")
			}

			// Load and show config info
			cfg, err := config.LoadTunnelConfig(projectDir, env)
			if err == nil {
				ui.KeyValue("Tunnel:", cfg.Tunnel)
				ui.KeyValue("Credentials:", cfg.CredentialsFile)
				ui.KeyValue("Ingress rules:", fmt.Sprintf("%d", len(cfg.Ingress)))

				for _, rule := range cfg.Ingress {
					if rule.Hostname != "" {
						ui.Dim("  %s → %s", rule.Hostname, rule.Service)
					} else {
						ui.Dim("  (catch-all) → %s", rule.Service)
					}
				}
			}

			// Show cloudflared version
			if detect.Found {
				ui.KeyValue("cloudflared:", detect.Version)
			} else {
				ui.KeyValue("cloudflared:", "not found")
			}
		}

		if watch {
			for {
				fmt.Print("\033[H\033[2J") // Clear screen
				printStatus()
				fmt.Println()
				ui.Dim("Refreshing every 5s... (Ctrl+C to stop)")
				time.Sleep(5 * time.Second)
			}
		}

		printStatus()
		return nil
	},
}

func init() {
	statusCmd.Flags().Bool("watch", false, "continuously refresh status")
	rootCmd.AddCommand(statusCmd)
}
