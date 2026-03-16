package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadTunnelConfig reads and parses the environment YAML config.
func LoadTunnelConfig(projectDir, env string) (*TunnelConfig, error) {
	path := EnvConfigPath(projectDir, env)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config %s: %w", path, err)
	}

	var cfg TunnelConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config %s: %w", path, err)
	}

	return &cfg, nil
}

// SaveTunnelConfig writes the tunnel config to the environment YAML file.
func SaveTunnelConfig(projectDir, env string, cfg *TunnelConfig) error {
	path := EnvConfigPath(projectDir, env)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}

// UpdateTunnelID updates the tunnel field in the environment config.
func UpdateTunnelID(projectDir, env, tunnelID string) error {
	cfg, err := LoadTunnelConfig(projectDir, env)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &TunnelConfig{}
		} else {
			return err
		}
	}

	cfg.Tunnel = tunnelID
	return SaveTunnelConfig(projectDir, env, cfg)
}

// UpdateCredentialsFile updates the credentials-file field in the environment config.
func UpdateCredentialsFile(projectDir, env, credPath string) error {
	cfg, err := LoadTunnelConfig(projectDir, env)
	if err != nil {
		if os.IsNotExist(err) {
			cfg = &TunnelConfig{}
		} else {
			return err
		}
	}

	cfg.CredentialsFile = credPath
	return SaveTunnelConfig(projectDir, env, cfg)
}
