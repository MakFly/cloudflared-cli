package cmd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var tunnelListOutput string

var tunnelListCmd = &cobra.Command{
	Use:   "list",
	Short: "List cloudflared tunnels",
	RunE:  runTunnelList,
}

func init() {
	tunnelListCmd.Flags().StringVarP(&tunnelListOutput, "output", "o", "table", "output format (table|json)")
	tunnelCmd.AddCommand(tunnelListCmd)
}

func runTunnelList(cmd *cobra.Command, args []string) error {
	exec, err := getExecutor()
	if err != nil {
		return err
	}

	result, err := exec.RunCommand(context.Background(), "tunnel", "list", "--output", "json")
	if err != nil {
		ui.Error("Failed to list tunnels: %s", result.Stderr)
		return fmt.Errorf("tunnel list failed: %w", err)
	}

	tunnels, err := cloudflared.ParseTunnelList(result.Stdout)
	if err != nil {
		return err
	}

	switch tunnelListOutput {
	case "json":
		data, err := json.MarshalIndent(tunnels, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		fmt.Println(string(data))
	default:
		headers := []string{"ID", "NAME", "CREATED", "STATUS"}
		rows := make([][]string, len(tunnels))
		for i, t := range tunnels {
			status := "active"
			if t.DeletedAt != "" {
				status = "deleted"
			}
			rows[i] = []string{t.ID, t.Name, t.CreatedAt, status}
		}
		ui.Table(headers, rows)
	}

	return nil
}
