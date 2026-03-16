package cmd

import (
	"context"
	"fmt"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var tunnelInfoCmd = &cobra.Command{
	Use:   "info <name-or-id>",
	Short: "Show details of a cloudflared tunnel",
	Args:  cobra.ExactArgs(1),
	RunE:  runTunnelInfo,
}

func init() {
	tunnelCmd.AddCommand(tunnelInfoCmd)
}

func runTunnelInfo(cmd *cobra.Command, args []string) error {
	nameOrID := args[0]

	exec, err := getExecutor()
	if err != nil {
		return err
	}

	result, err := exec.RunCommand(context.Background(), "tunnel", "info", nameOrID)
	if err != nil {
		ui.Error("Failed to get tunnel info: %s", result.Stderr)
		return fmt.Errorf("tunnel info failed: %w", err)
	}

	output := result.Stdout
	if output == "" {
		output = result.Stderr
	}

	info, err := cloudflared.ParseTunnelInfo(output)
	if err != nil {
		return err
	}

	ui.Bold("Tunnel Information")
	fmt.Println()
	ui.KeyValue("ID:", info.ID)
	ui.KeyValue("Name:", info.Name)
	ui.KeyValue("Created:", info.CreatedAt)
	if info.ConnectorID != "" {
		ui.KeyValue("Connector ID:", info.ConnectorID)
	}
	if info.DeletedAt != "" {
		ui.KeyValue("Deleted:", info.DeletedAt)
	}

	return nil
}
