package cmd

import (
	"context"
	"fmt"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate tunnel configuration",
	Long:  "Runs local validation checks and optionally validates with cloudflared.",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)

		cfg, err := config.LoadTunnelConfig(projectDir, env)
		if err != nil {
			return fmt.Errorf("failed to load config for env %q: %w", env, err)
		}

		ui.Info("Validating configuration [%s]...", env)

		// Local validation
		result := config.ValidateTunnelConfig(cfg)
		if !result.IsValid() {
			ui.Error("Local validation failed:")
			for _, e := range result.Errors {
				ui.Warn("  %s", e)
			}
			return fmt.Errorf("configuration has %d error(s)", len(result.Errors))
		}
		ui.Success("Local validation passed")

		// cloudflared ingress validate (if available)
		detect := cloudflared.Detect(cloudflaredPath)
		if detect.Found {
			executor := cloudflared.NewExecutor(detect.Path)
			configPath := config.EnvConfigPath(projectDir, env)
			res, err := executor.RunCommand(
				context.Background(),
				"tunnel", "ingress", "validate", "--config", configPath,
			)
			if err != nil {
				ui.Warn("cloudflared validation: %s", res.Stderr)
			} else {
				ui.Success("cloudflared ingress validation passed")
			}
		} else {
			ui.Dim("Skipped cloudflared validation (binary not found)")
		}

		return nil
	},
}

func init() {
	configCmd.AddCommand(configValidateCmd)
}
