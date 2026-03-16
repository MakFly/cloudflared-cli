package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage tunnel configuration",
	Long:  "View, modify, and validate cloudflared tunnel configuration for each environment.",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
