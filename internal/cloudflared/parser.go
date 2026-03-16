package cloudflared

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/kev/cloudflared-cli/internal/config"
)

var uuidPattern = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// ParseTunnelList parses the output of `cloudflared tunnel list --output json`.
// It expects JSON array output but falls back to text parsing if JSON fails.
func ParseTunnelList(output string) ([]config.TunnelInfo, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, nil
	}

	// Try JSON parsing first
	var tunnels []config.TunnelInfo
	if err := json.Unmarshal([]byte(output), &tunnels); err == nil {
		return tunnels, nil
	}

	// Fallback: extract UUIDs and names from text output
	lines := strings.Split(output, "\n")
	var result []config.TunnelInfo
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		uuid := uuidPattern.FindString(line)
		if uuid == "" {
			continue
		}

		info := config.TunnelInfo{ID: uuid}

		// Try to extract a name: typically the field after the UUID
		parts := strings.Fields(line)
		for i, p := range parts {
			if p == uuid && i+1 < len(parts) {
				info.Name = parts[i+1]
				break
			}
		}
		result = append(result, info)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("failed to parse tunnel list output")
	}
	return result, nil
}

// ParseTunnelCreate parses the output of `cloudflared tunnel create` to
// extract the tunnel UUID and name.
func ParseTunnelCreate(output string) (config.TunnelInfo, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return config.TunnelInfo{}, fmt.Errorf("empty tunnel create output")
	}

	// Try JSON first
	var info config.TunnelInfo
	if err := json.Unmarshal([]byte(output), &info); err == nil && info.ID != "" {
		return info, nil
	}

	// Fallback: extract UUID from text
	uuid := uuidPattern.FindString(output)
	if uuid == "" {
		return config.TunnelInfo{}, fmt.Errorf("could not find tunnel UUID in output: %s", output)
	}

	info = config.TunnelInfo{ID: uuid}

	// Try to extract the tunnel name from common patterns like:
	// "Created tunnel <name> with id <uuid>"
	nameRe := regexp.MustCompile(`(?i)created tunnel\s+(\S+)\s+with id`)
	if m := nameRe.FindStringSubmatch(output); len(m) > 1 {
		info.Name = m[1]
	}

	return info, nil
}

// ParseTunnelInfo parses the output of `cloudflared tunnel info`.
func ParseTunnelInfo(output string) (*config.TunnelInfo, error) {
	output = strings.TrimSpace(output)
	if output == "" {
		return nil, fmt.Errorf("empty tunnel info output")
	}

	// Try JSON first
	var info config.TunnelInfo
	if err := json.Unmarshal([]byte(output), &info); err == nil && info.ID != "" {
		return &info, nil
	}

	// Fallback: extract UUID from text
	uuid := uuidPattern.FindString(output)
	if uuid == "" {
		return nil, fmt.Errorf("could not find tunnel UUID in output: %s", output)
	}

	info = config.TunnelInfo{ID: uuid}

	// Try to find name
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "name:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.Name = strings.TrimSpace(parts[1])
			}
		}
		if strings.Contains(lower, "created:") || strings.Contains(lower, "createdat:") || strings.Contains(lower, "created_at:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				info.CreatedAt = strings.TrimSpace(parts[1])
			}
		}
	}

	return &info, nil
}
