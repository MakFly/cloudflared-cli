package config

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestEnv(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	envDir := filepath.Join(tmpDir, envDirName)
	if err := os.MkdirAll(envDir, 0o755); err != nil {
		t.Fatal(err)
	}
	return tmpDir
}

func TestLoadSaveTunnelConfig(t *testing.T) {
	baseDir := setupTestEnv(t)

	cfg := &TunnelConfig{
		Tunnel:          "test-uuid",
		CredentialsFile: "/path/to/creds.json",
		Ingress: []IngressRule{
			{Hostname: "app.example.com", Service: "http://localhost:3000"},
			{Service: "http_status:404"},
		},
	}

	if err := SaveTunnelConfig(baseDir, "dev", cfg); err != nil {
		t.Fatalf("SaveTunnelConfig: %v", err)
	}

	loaded, err := LoadTunnelConfig(baseDir, "dev")
	if err != nil {
		t.Fatalf("LoadTunnelConfig: %v", err)
	}

	if loaded.Tunnel != cfg.Tunnel {
		t.Errorf("Tunnel = %q, want %q", loaded.Tunnel, cfg.Tunnel)
	}
	if len(loaded.Ingress) != 2 {
		t.Errorf("Ingress count = %d, want 2", len(loaded.Ingress))
	}
}

func TestUpdateTunnelID(t *testing.T) {
	baseDir := setupTestEnv(t)

	cfg := &TunnelConfig{
		Tunnel:  "old-id",
		Ingress: []IngressRule{{Service: "http_status:404"}},
	}
	if err := SaveTunnelConfig(baseDir, "dev", cfg); err != nil {
		t.Fatal(err)
	}

	if err := UpdateTunnelID(baseDir, "dev", "new-uuid"); err != nil {
		t.Fatalf("UpdateTunnelID: %v", err)
	}

	loaded, err := LoadTunnelConfig(baseDir, "dev")
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Tunnel != "new-uuid" {
		t.Errorf("Tunnel = %q, want %q", loaded.Tunnel, "new-uuid")
	}
}
