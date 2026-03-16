package cmd

import (
	"fmt"
	"os"

	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgEnv          string
	projectName     string
	projectDirFlag  string
	verbose         bool
	cloudflaredPath string

	// projectDir is resolved at runtime from --project / --project-dir / local detection.
	projectDir string
)

var rootCmd = &cobra.Command{
	Use:   "cloudflared-project",
	Short: "CLI wrapper for managing Cloudflare Tunnel projects",
	Long: `cloudflared-project simplifies Cloudflare Tunnel management by wrapping
the cloudflared binary with project-aware configuration, multi-environment
support, and streamlined workflows.

Projects are stored in ~/.cloudflared/projects/ by default.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Resolve project directory (skip for init and version)
		if cmd.Name() == "init" || cmd.Name() == "version" {
			return
		}
		projectDir = config.ResolveProjectDir(projectDirFlag, projectName)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&cfgEnv, "env", "e", "dev", "target environment (dev/staging/prod)")
	rootCmd.PersistentFlags().StringVarP(&projectName, "project", "p", "", "project name (resolved from ~/.cloudflared/projects/)")
	rootCmd.PersistentFlags().StringVar(&projectDirFlag, "project-dir", "", "explicit project directory path (overrides --project)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().StringVar(&cloudflaredPath, "cloudflared-path", "", "explicit path to cloudflared binary")

	_ = viper.BindPFlag("env", rootCmd.PersistentFlags().Lookup("env"))
	_ = viper.BindPFlag("project", rootCmd.PersistentFlags().Lookup("project"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("cloudflared_path", rootCmd.PersistentFlags().Lookup("cloudflared-path"))
}

func initConfig() {
	viper.SetEnvPrefix("CLOUDFLARED_PROJECT")
	viper.AutomaticEnv()

	if cloudflaredPath == "" {
		cloudflaredPath = viper.GetString("cloudflared_path")
	}
}
