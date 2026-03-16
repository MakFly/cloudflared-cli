package cmd

import (
	"fmt"

	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var configRemoveCmd = &cobra.Command{
	Use:   "remove-ingress",
	Short: "Remove an ingress rule by hostname",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)
		hostname, _ := cmd.Flags().GetString("hostname")

		if hostname == "" {
			return fmt.Errorf("--hostname is required")
		}

		cfg, err := config.LoadTunnelConfig(projectDir, env)
		if err != nil {
			return fmt.Errorf("failed to load config for env %q: %w", env, err)
		}

		found := false
		newIngress := make([]config.IngressRule, 0, len(cfg.Ingress))
		for _, rule := range cfg.Ingress {
			if rule.Hostname == hostname {
				found = true
				continue
			}
			newIngress = append(newIngress, rule)
		}

		if !found {
			return fmt.Errorf("no ingress rule found for hostname %q", hostname)
		}

		cfg.Ingress = newIngress

		if err := config.SaveTunnelConfig(projectDir, env, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		ui.Success("Removed ingress rule for hostname %q [%s]", hostname, env)
		return nil
	},
}

func init() {
	configRemoveCmd.Flags().String("hostname", "", "hostname of the rule to remove (required)")
	configCmd.AddCommand(configRemoveCmd)
}
