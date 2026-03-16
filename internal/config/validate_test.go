package config

import (
	"testing"
)

func TestValidateTunnelConfig_Valid(t *testing.T) {
	cfg := &TunnelConfig{
		Tunnel:          "my-tunnel",
		CredentialsFile: "~/.cloudflared/abc.json",
		Ingress: []IngressRule{
			{Hostname: "app.example.com", Service: "http://localhost:3000"},
			{Service: "http_status:404"},
		},
	}

	result := ValidateTunnelConfig(cfg)
	if !result.IsValid() {
		t.Errorf("expected valid config, got errors: %s", result.Error())
	}
}

func TestValidateTunnelConfig_MissingTunnel(t *testing.T) {
	cfg := &TunnelConfig{
		CredentialsFile: "~/.cloudflared/abc.json",
		Ingress: []IngressRule{
			{Service: "http_status:404"},
		},
	}

	result := ValidateTunnelConfig(cfg)
	if result.IsValid() {
		t.Error("expected validation error for missing tunnel")
	}
}

func TestValidateTunnelConfig_MissingCatchAll(t *testing.T) {
	cfg := &TunnelConfig{
		Tunnel:          "my-tunnel",
		CredentialsFile: "~/.cloudflared/abc.json",
		Ingress: []IngressRule{
			{Hostname: "app.example.com", Service: "http://localhost:3000"},
		},
	}

	result := ValidateTunnelConfig(cfg)
	if result.IsValid() {
		t.Error("expected validation error for missing catch-all")
	}

	found := false
	for _, e := range result.Errors {
		if e.Field == "ingress" {
			found = true
		}
	}
	if !found {
		t.Error("expected ingress validation error")
	}
}

func TestValidateTunnelConfig_EmptyIngress(t *testing.T) {
	cfg := &TunnelConfig{
		Tunnel:          "my-tunnel",
		CredentialsFile: "~/.cloudflared/abc.json",
	}

	result := ValidateTunnelConfig(cfg)
	if result.IsValid() {
		t.Error("expected validation error for empty ingress")
	}
}

func TestValidateTunnelConfig_ValidUUID(t *testing.T) {
	cfg := &TunnelConfig{
		Tunnel:          "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
		CredentialsFile: "/path/to/creds.json",
		Ingress: []IngressRule{
			{Service: "http_status:404"},
		},
	}

	result := ValidateTunnelConfig(cfg)
	if !result.IsValid() {
		t.Errorf("expected valid config with UUID, got: %s", result.Error())
	}
}

func TestIsCatchAll(t *testing.T) {
	tests := []struct {
		name     string
		rule     IngressRule
		expected bool
	}{
		{"catch-all", IngressRule{Service: "http_status:404"}, true},
		{"with hostname", IngressRule{Hostname: "app.example.com", Service: "http://localhost:3000"}, false},
		{"with path", IngressRule{Path: "/api/*", Service: "http://localhost:3000"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rule.IsCatchAll(); got != tt.expected {
				t.Errorf("IsCatchAll() = %v, want %v", got, tt.expected)
			}
		})
	}
}
