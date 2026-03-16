package cmd

import (
	"github.com/kev/cloudflared-cli/internal/cloudflared"
	"github.com/kev/cloudflared-cli/internal/config"
	"github.com/kev/cloudflared-cli/internal/project"
	"github.com/kev/cloudflared-cli/internal/ui"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init <name>",
	Short: "Initialize a new cloudflared project",
	Long: `Initialize a new cloudflared project in ~/.cloudflared/projects/<name>/.
Creates project metadata and a dev environment configuration file.

Use --local to create the project in the current directory instead.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		tunnelName, _ := cmd.Flags().GetString("tunnel-name")
		domain, _ := cmd.Flags().GetString("domain")
		force, _ := cmd.Flags().GetBool("force")
		local, _ := cmd.Flags().GetBool("local")

		// Determine target directory
		targetDir := config.ProjectPath(name)
		if local {
			targetDir = ".cloudflared-project"
		}
		if projectDirFlag != "" {
			targetDir = projectDirFlag
		}

		opts := project.ScaffoldOptions{
			ProjectName: name,
			TunnelName:  tunnelName,
			Domain:      domain,
			Force:       force,
		}

		if err := project.Scaffold(targetDir, opts); err != nil {
			ui.Error("Failed to initialize project: %s", err)
			return err
		}

		ui.Success("Project %q initialized", name)
		ui.Info("Location: %s", targetDir)
		ui.Info("Default environment: dev")
		if domain != "" {
			ui.Info("Domain: %s", domain)
		}
		ui.Dim("Use 'cloudflared-project -p %s <command>' to manage this project", name)

		// Check cloudflared installation and authentication
		detect := cloudflared.Detect(cloudflaredPath)
		if detect.Found && !cloudflared.IsAuthenticated() {
			ui.Warn("cloudflared login required before creating tunnels")
			ui.Info("  Run: cloudflared-project login")
		}

		return nil
	},
}

func init() {
	initCmd.Flags().String("tunnel-name", "", "tunnel name (defaults to <project>-<env>)")
	initCmd.Flags().String("domain", "", "domain for ingress rules")
	initCmd.Flags().Bool("force", false, "overwrite existing project")
	initCmd.Flags().Bool("local", false, "create in current directory instead of ~/.cloudflared/projects/")

	rootCmd.AddCommand(initCmd)
}
