package project

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/kev/cloudflared-cli/internal/config"
)

// ScaffoldOptions contains the parameters for creating a new project scaffold.
type ScaffoldOptions struct {
	ProjectName string
	TunnelName  string
	Domain      string
	Force       bool
}

// Scaffold creates the project directory structure with project metadata
// and environment configuration files.
// projectDir is the resolved target directory (e.g. ~/.cloudflared/projects/myapp/).
func Scaffold(projectDir string, opts ScaffoldOptions) error {
	// Check if directory already exists and has a project.yaml
	if _, err := os.Stat(filepath.Join(projectDir, "project.yaml")); err == nil && !opts.Force {
		return fmt.Errorf("project %q already exists at %s (use --force to overwrite)", opts.ProjectName, projectDir)
	}

	// Create directory structure
	envsDir := filepath.Join(projectDir, "environments")
	if err := os.MkdirAll(envsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory structure: %w", err)
	}

	// Write project.yaml
	projectCfg := &config.ProjectConfig{
		Version:    "1",
		Name:       opts.ProjectName,
		DefaultEnv: "dev",
	}
	if err := config.SaveProject(projectDir, projectCfg); err != nil {
		return fmt.Errorf("failed to write project.yaml: %w", err)
	}

	// Write dev environment config
	if err := writeEnvConfig(envsDir, "dev", opts); err != nil {
		return fmt.Errorf("failed to write dev.yaml: %w", err)
	}

	return nil
}

// writeEnvConfig creates a cloudflared-native environment config file.
func writeEnvConfig(envsDir, envName string, opts ScaffoldOptions) error {
	tunnelName := opts.TunnelName
	if tunnelName == "" {
		tunnelName = fmt.Sprintf("%s-%s", opts.ProjectName, envName)
	}

	ingress := []config.IngressRule{
		{Service: "http_status:404"},
	}

	if opts.Domain != "" {
		ingress = []config.IngressRule{
			{
				Hostname: opts.Domain,
				Service:  "http://localhost:8080",
			},
			{Service: "http_status:404"},
		}
	}

	tunnelCfg := config.TunnelConfig{
		Tunnel:  tunnelName,
		Ingress: ingress,
	}

	data, err := yaml.Marshal(tunnelCfg)
	if err != nil {
		return err
	}

	path := filepath.Join(envsDir, envName+".yaml")
	return os.WriteFile(path, data, 0o644)
}
