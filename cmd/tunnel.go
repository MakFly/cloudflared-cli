package cmd

import (
	"fmt"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/spf13/cobra"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Manage cloudflared tunnels",
}

func init() {
	rootCmd.AddCommand(tunnelCmd)
}

// getExecutor creates an Executor by detecting the cloudflared binary.
func getExecutor() (cloudflared.Executor, error) {
	result := cloudflared.Detect(cloudflaredPath)
	if !result.Found {
		return nil, fmt.Errorf("%s", cloudflared.InstallGuide())
	}
	return cloudflared.NewExecutor(result.Path), nil
}
