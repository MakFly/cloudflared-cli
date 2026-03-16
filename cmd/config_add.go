package cmd

import (
	"fmt"

	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var configAddCmd = &cobra.Command{
	Use:   "add-ingress",
	Short: "Add an ingress rule before the catch-all",
	RunE: func(cmd *cobra.Command, args []string) error {
		env := config.ResolveEnv(projectDir, cfgEnv)
		hostname, _ := cmd.Flags().GetString("hostname")
		service, _ := cmd.Flags().GetString("service")
		path, _ := cmd.Flags().GetString("path")

		if service == "" {
			return fmt.Errorf("--service is required")
		}

		cfg, err := config.LoadTunnelConfig(projectDir, env)
		if err != nil {
			return fmt.Errorf("failed to load config for env %q: %w", env, err)
		}

		rule := config.IngressRule{
			Hostname: hostname,
			Service:  service,
			Path:     path,
		}

		// Insert before the catch-all (last rule)
		if len(cfg.Ingress) == 0 {
			cfg.Ingress = []config.IngressRule{rule, {Service: "http_status:404"}}
		} else {
			// Insert before the last rule (catch-all)
			newIngress := make([]config.IngressRule, 0, len(cfg.Ingress)+1)
			newIngress = append(newIngress, cfg.Ingress[:len(cfg.Ingress)-1]...)
			newIngress = append(newIngress, rule)
			newIngress = append(newIngress, cfg.Ingress[len(cfg.Ingress)-1])
			cfg.Ingress = newIngress
		}

		if err := config.SaveTunnelConfig(projectDir, env, cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		ui.Success("Added ingress rule: %s → %s [%s]", hostname, service, env)
		return nil
	},
}

func init() {
	configAddCmd.Flags().String("hostname", "", "hostname for the ingress rule")
	configAddCmd.Flags().String("service", "", "backend service URL (required)")
	configAddCmd.Flags().String("path", "", "path pattern (optional)")
	configCmd.AddCommand(configAddCmd)
}
