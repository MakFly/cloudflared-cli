package cmd

import (
	"fmt"

	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Display resolved configuration for an environment",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)

		if !config.IsProjectInitialized(projectDir) {
			return fmt.Errorf("project not initialized, run 'cloudflared-project init' first")
		}

		cfg, err := config.LoadTunnelConfig(projectDir, env)
		if err != nil {
			return fmt.Errorf("failed to load config for env %q: %w", env, err)
		}

		ui.Bold("Configuration [%s]", env)
		fmt.Println()

		data, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		fmt.Print(string(data))
		return nil
	},
}

func init() {
	configCmd.AddCommand(configShowCmd)
}
