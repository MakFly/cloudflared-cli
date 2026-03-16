package config

import (
	"testing"
)

func TestLoadSaveProject(t *testing.T) {
	projDir := t.TempDir()

	cfg := &ProjectConfig{
		Version:    "1",
		Name:       "test-project",
		DefaultEnv: "staging",
	}

	if err := SaveProject(projDir, cfg); err != nil {
		t.Fatalf("SaveProject: %v", err)
	}

	loaded, err := LoadProject(projDir)
	if err != nil {
		t.Fatalf("LoadProject: %v", err)
	}

	if loaded.Name != cfg.Name {
		t.Errorf("Name = %q, want %q", loaded.Name, cfg.Name)
	}
	if loaded.DefaultEnv != cfg.DefaultEnv {
		t.Errorf("DefaultEnv = %q, want %q", loaded.DefaultEnv, cfg.DefaultEnv)
	}
}
