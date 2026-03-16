package cmd

import (
	"fmt"
	"strconv"

	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long: `Set a top-level configuration key. Supported keys:
  tunnel          - Tunnel UUID or name
  credentials-file - Path to credentials JSON file`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)
		key := args[0]
		value := args[1]

		cfg, err := config.LoadTunnelConfig(projectDir, env)
		if err != nil {
			return fmt.Errorf("failed to load config for env %q: %w", env, err)
		}

		switch key {
		case "tunnel":
			cfg.Tunnel = value
		case "credentials-file":
			cfg.CredentialsFile = value
		case "warp-routing":
			enabled, err := strconv.ParseBool(value)
			if err != nil {
				return fmt.Errorf("warp-routing must be true or false")
			}
			if cfg.WarpRouting == nil {
				cfg.WarpRouting = &config.WarpRouting{}
			}
			cfg.WarpRouting.Enabled = enabled
		default:
			return fmt.Errorf("unsupported key %q, supported: tunnel, credentials-file, warp-routing", key)
		}

		if err := config.SaveTunnelConfig(projectDir, env, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		ui.Success("Set %s = %s [%s]", key, value, env)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
