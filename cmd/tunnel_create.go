package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var tunnelCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new cloudflared tunnel",
	Args:  cobra.ExactArgs(1),
	RunE:  runTunnelCreate,
}

func init() {
	tunnelCmd.AddCommand(tunnelCreateCmd)
}

func runTunnelCreate(cmd *cobra.Command, args []string) error {
	name := args[0]

	exec, err := getExecutor()
	if err != nil {
		return err
	}

	if !cloudflared.IsAuthenticated() {
		return fmt.Errorf("not authenticated, run 'cloudflared-project login' first")
	}

	ui.Info("Creating tunnel %q...", name)

	result, err := exec.RunCommand(context.Background(), "tunnel", "create", name)
	if err != nil {
		ui.Error("Failed to create tunnel: %s", result.Stderr)
		return fmt.Errorf("tunnel create failed: %w", err)
	}

	// Parse the combined output (cloudflared may write to stdout or stderr)
	output := result.Stdout
	if output == "" {
		output = result.Stderr
	}

	info, err := cloudflared.ParseTunnelCreate(output)
	if err != nil {
		ui.Error("Tunnel may have been created but output could not be parsed")
		return err
	}

	ui.Success("Tunnel %q created with ID: %s", info.Name, info.ID)

	// Update environment config with tunnel ID
	if err := config.UpdateTunnelID(projectDir, cfgEnv, info.ID); err != nil {
		ui.Warn("Failed to update environment config: %s", err)
	} else {
		ui.Info("Updated %s environment config with tunnel ID", cfgEnv)
	}

	// Update credentials file path
	homeDir, err := os.UserHomeDir()
	if err == nil {
		credPath := filepath.Join(homeDir, ".cloudflared", info.ID+".json")
		if err := config.UpdateCredentialsFile(projectDir, cfgEnv, credPath); err != nil {
			ui.Warn("Failed to update credentials file path: %s", err)
		} else {
			ui.Info("Updated credentials file path: %s", credPath)
		}
	}

	return nil
}
