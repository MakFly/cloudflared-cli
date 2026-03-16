package cloudflared

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// DetectResult holds information about a detected cloudflared binary.
type DetectResult struct {
	Path    string
	Version string
	Found   bool
}

// Detect searches for the cloudflared binary using the following priority:
// 1. Explicit path (from --cloudflared-path flag)
// 2. CLOUDFLARED_PATH environment variable (handled by caller via Viper)
// 3. System PATH
// 4. Common installation paths
func Detect(explicitPath string) DetectResult {
	if explicitPath != "" {
		if ver, err := getVersion(explicitPath); err == nil {
			return DetectResult{Path: explicitPath, Version: ver, Found: true}
		}
	}

	if path, err := exec.LookPath("cloudflared"); err == nil {
		if ver, err := getVersion(path); err == nil {
			return DetectResult{Path: path, Version: ver, Found: true}
		}
	}

	for _, p := range commonPaths() {
		if ver, err := getVersion(p); err == nil {
			return DetectResult{Path: p, Version: ver, Found: true}
		}
	}

	return DetectResult{}
}

func getVersion(binaryPath string) (string, error) {
	out, err := exec.Command(binaryPath, "version").CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func commonPaths() []string {
	paths := []string{
		"/usr/local/bin/cloudflared",
		"/usr/bin/cloudflared",
	}
	if runtime.GOOS == "darwin" {
		paths = append(paths, "/opt/homebrew/bin/cloudflared")
	}
	return paths
}

// CertPath returns the path to the cloudflared certificate file.
func CertPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".cloudflared", "cert.pem")
}

// IsAuthenticated checks whether the user has logged in to cloudflared
// by verifying that ~/.cloudflared/cert.pem exists.
func IsAuthenticated() bool {
	certPath := CertPath()
	if certPath == "" {
		return false
	}
	_, err := os.Stat(certPath)
	return err == nil
}

// InstallGuide returns OS-specific installation instructions.
func InstallGuide() string {
	switch runtime.GOOS {
	case "darwin":
		return fmt.Sprintf(`cloudflared not found. Install it with:
  brew install cloudflared

Or download from: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/`)
	case "linux":
		return fmt.Sprintf(`cloudflared not found. Install it with:
  # Debian/Ubuntu
  curl -fsSL https://pkg.cloudflare.com/cloudflare-main.gpg | sudo tee /usr/share/keyrings/cloudflare-main.gpg >/dev/null
  echo "deb [signed-by=/usr/share/keyrings/cloudflare-main.gpg] https://pkg.cloudflare.com/cloudflared $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/cloudflared.list
  sudo apt-get update && sudo apt-get install cloudflared

Or download from: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/`)
	default:
		return "cloudflared not found. Download from: https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/"
	}
}
