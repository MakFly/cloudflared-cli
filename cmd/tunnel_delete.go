package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var (
	tunnelDeleteForce   bool
	tunnelDeleteCleanup bool
)

var tunnelDeleteCmd = &cobra.Command{
	Use:   "delete <name-or-id>",
	Short: "Delete a cloudflared tunnel",
	Args:  cobra.ExactArgs(1),
	RunE:  runTunnelDelete,
}

func init() {
	tunnelDeleteCmd.Flags().BoolVarP(&tunnelDeleteForce, "force", "f", false, "skip confirmation prompt")
	tunnelDeleteCmd.Flags().BoolVar(&tunnelDeleteCleanup, "cleanup", false, "also remove tunnel from environment config")
	tunnelCmd.AddCommand(tunnelDeleteCmd)
}

func runTunnelDelete(cmd *cobra.Command, args []string) error {
	nameOrID := args[0]

	if !tunnelDeleteForce {
		ui.Warn("You are about to delete tunnel %q. This action cannot be undone.", nameOrID)
		fmt.Print("Are you sure? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			ui.Info("Deletion cancelled.")
			return nil
		}
	}

	exec, err := getExecutor()
	if err != nil {
		return err
	}

	ui.Info("Deleting tunnel %q...", nameOrID)

	result, err := exec.RunCommand(context.Background(), "tunnel", "delete", nameOrID)
	if err != nil {
		ui.Error("Failed to delete tunnel: %s", result.Stderr)
		return fmt.Errorf("tunnel delete failed: %w", err)
	}

	ui.Success("Tunnel %q deleted.", nameOrID)

	if tunnelDeleteCleanup {
		cfg, loadErr := config.LoadTunnelConfig(projectDir, cfgEnv)
		if loadErr != nil {
			ui.Warn("Could not load environment config for cleanup: %s", loadErr)
			return nil
		}
		cfg.Tunnel = ""
		cfg.CredentialsFile = ""
		if saveErr := config.SaveTunnelConfig(projectDir, cfgEnv, cfg); saveErr != nil {
			ui.Warn("Failed to clean up environment config: %s", saveErr)
		} else {
			ui.Success("Cleaned up %s environment config.", cfgEnv)
		}
	}

	return nil
}
