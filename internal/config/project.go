package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ProjectConfig represents the project-level metadata stored in project.yaml.
type ProjectConfig struct {
	Version    string `yaml:"version"`
	Name       string `yaml:"name"`
	DefaultEnv string `yaml:"default_env"`
}

// LoadProject reads and parses the project.yaml from the given project directory.
func LoadProject(projectDir string) (*ProjectConfig, error) {
	path := filepath.Join(projectDir, projectFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// SaveProject writes the project config to project.yaml in the given project directory.
func SaveProject(projectDir string, cfg *ProjectConfig) error {
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(projectDir, projectFile), data, 0o644)
}
