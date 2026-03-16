package cmd

import (
	"context"
	"fmt"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with Cloudflare (opens browser)",
	Long:  `Runs 'cloudflared login' to authenticate with your Cloudflare account. A browser window will open for authorization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if cloudflared.IsAuthenticated() {
			ui.Success("Already authenticated")
			ui.Info("Certificate: %s", cloudflared.CertPath())
			return nil
		}

		detect := cloudflared.Detect(cloudflaredPath)
		if !detect.Found {
			return fmt.Errorf("%s", cloudflared.InstallGuide())
		}

		executor := cloudflared.NewExecutor(detect.Path)

		ui.Info("Opening browser for Cloudflare authentication...")
		if err := executor.RunAttached(context.Background(), "login"); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		if cloudflared.IsAuthenticated() {
			ui.Success("Authentication successful")
			ui.Info("Certificate saved: %s", cloudflared.CertPath())
			return nil
		}

		return fmt.Errorf("login completed but certificate not found at %s", cloudflared.CertPath())
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
