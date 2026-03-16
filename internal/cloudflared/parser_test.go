package cloudflared

import (
	"testing"
)

func TestParseTunnelList_JSON(t *testing.T) {
	input := `[{"id":"abc-123","name":"my-tunnel","created_at":"2024-01-01T00:00:00Z"},{"id":"def-456","name":"other","created_at":"2024-02-01T00:00:00Z"}]`

	tunnels, err := ParseTunnelList(input)
	if err != nil {
		t.Fatalf("ParseTunnelList: %v", err)
	}

	if len(tunnels) != 2 {
		t.Fatalf("expected 2 tunnels, got %d", len(tunnels))
	}

	if tunnels[0].Name != "my-tunnel" {
		t.Errorf("first tunnel name = %q, want %q", tunnels[0].Name, "my-tunnel")
	}
}

func TestParseTunnelCreate(t *testing.T) {
	input := `Created tunnel my-tunnel with id a1b2c3d4-e5f6-7890-abcd-ef1234567890`

	info, err := ParseTunnelCreate(input)
	if err != nil {
		t.Fatalf("ParseTunnelCreate: %v", err)
	}

	if info.ID != "a1b2c3d4-e5f6-7890-abcd-ef1234567890" {
		t.Errorf("ID = %q, want UUID", info.ID)
	}
	if info.Name != "my-tunnel" {
		t.Errorf("Name = %q, want %q", info.Name, "my-tunnel")
	}
}

func TestParseTunnelCreate_NoMatch(t *testing.T) {
	_, err := ParseTunnelCreate("some random output")
	if err == nil {
		t.Error("expected error for unparseable output")
	}
}
