package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	envDirName  = "environments"
	projectFile = "project.yaml"
)

// DefaultProjectsDir returns the default base directory for all projects.
// Uses ~/.cloudflared/projects/ which leverages the existing ~/.cloudflared/
// directory created by cloudflared on both macOS and Linux.
func DefaultProjectsDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".cloudflared", "projects")
	}
	return filepath.Join(home, ".cloudflared", "projects")
}

// ProjectPath returns the full path to a named project.
func ProjectPath(name string) string {
	return filepath.Join(DefaultProjectsDir(), name)
}

// ResolveProjectDir resolves the project directory using:
// 1. Explicit path (from --project-dir flag, if not default)
// 2. Named project in ~/.cloudflared/projects/<name>/
// 3. Local .cloudflared-project/ in current directory (legacy/override)
func ResolveProjectDir(explicitDir, projectName string) string {
	// Explicit --project-dir takes priority
	if explicitDir != "" {
		return explicitDir
	}

	// Named project in default location
	if projectName != "" {
		p := ProjectPath(projectName)
		if _, err := os.Stat(filepath.Join(p, projectFile)); err == nil {
			return p
		}
	}

	// Local project in current directory
	localDir := filepath.Join(".", ".cloudflared-project")
	if _, err := os.Stat(filepath.Join(localDir, projectFile)); err == nil {
		return localDir
	}

	// If a name was given, return the default path (for init)
	if projectName != "" {
		return ProjectPath(projectName)
	}

	return ""
}

// EnvConfigPath returns the path to a specific environment's config file.
func EnvConfigPath(projectDir, env string) string {
	return filepath.Join(projectDir, envDirName, env+".yaml")
}

// EnvDir returns the path to the environments directory.
func EnvDir(projectDir string) string {
	return filepath.Join(projectDir, envDirName)
}

// ProjectConfigPath returns the path to project.yaml.
func ProjectConfigPath(projectDir string) string {
	return filepath.Join(projectDir, projectFile)
}

// ResolveEnv resolves the environment name using:
// 1. Explicit env parameter (from --env flag)
// 2. Project config default_env
// 3. Fallback to "dev"
func ResolveEnv(projectDir, explicitEnv string) string {
	if explicitEnv != "" {
		return explicitEnv
	}

	cfg, err := LoadProject(projectDir)
	if err == nil && cfg.DefaultEnv != "" {
		return cfg.DefaultEnv
	}

	return "dev"
}

// ListEnvironments returns all available environment names.
func ListEnvironments(projectDir string) ([]string, error) {
	envPath := EnvDir(projectDir)
	entries, err := os.ReadDir(envPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read environments directory: %w", err)
	}

	var envs []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext == ".yaml" || ext == ".yml" {
			envs = append(envs, name[:len(name)-len(ext)])
		}
	}
	return envs, nil
}

// ListProjects returns all project names in the default projects directory.
func ListProjects() ([]string, error) {
	dir := DefaultProjectsDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var projects []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Verify it's a valid project (has project.yaml)
		if _, err := os.Stat(filepath.Join(dir, entry.Name(), projectFile)); err == nil {
			projects = append(projects, entry.Name())
		}
	}
	return projects, nil
}

// EnvExists checks if an environment config file exists.
func EnvExists(projectDir, env string) bool {
	_, err := os.Stat(EnvConfigPath(projectDir, env))
	return err == nil
}

// IsProjectInitialized checks if the project has been initialized.
func IsProjectInitialized(projectDir string) bool {
	if projectDir == "" {
		return false
	}
	_, err := os.Stat(ProjectConfigPath(projectDir))
	return err == nil
}
