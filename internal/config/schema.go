package config

// TunnelConfig represents the cloudflared tunnel configuration YAML.
// This struct maps 1:1 to cloudflared's native config format so it can
// be passed directly via --config.
type TunnelConfig struct {
	Tunnel          string         `yaml:"tunnel"`
	CredentialsFile string         `yaml:"credentials-file"`
	Ingress         []IngressRule  `yaml:"ingress"`
	OriginRequest   *OriginRequest `yaml:"originRequest,omitempty"`
	WarpRouting     *WarpRouting   `yaml:"warp-routing,omitempty"`
}

// IngressRule represents a single ingress entry in cloudflared config.
type IngressRule struct {
	Hostname      string         `yaml:"hostname,omitempty"`
	Service       string         `yaml:"service"`
	Path          string         `yaml:"path,omitempty"`
	OriginRequest *OriginRequest `yaml:"originRequest,omitempty"`
}

// OriginRequest holds origin-specific settings for an ingress rule.
type OriginRequest struct {
	ConnectTimeout     string `yaml:"connectTimeout,omitempty"`
	TLSTimeout         string `yaml:"tlsTimeout,omitempty"`
	TCPKeepAlive       string `yaml:"tcpKeepAlive,omitempty"`
	NoHappyEyeballs    bool   `yaml:"noHappyEyeballs,omitempty"`
	KeepAliveTimeout   string `yaml:"keepAliveTimeout,omitempty"`
	KeepAliveConnCount int    `yaml:"keepAliveConnections,omitempty"`
	HTTPHostHeader     string `yaml:"httpHostHeader,omitempty"`
	OriginServerName   string `yaml:"originServerName,omitempty"`
	NoTLSVerify        bool   `yaml:"noTLSVerify,omitempty"`
	DisableChunked     bool   `yaml:"disableChunkedEncoding,omitempty"`
	ProxyAddress       string `yaml:"proxyAddress,omitempty"`
	ProxyPort          int    `yaml:"proxyPort,omitempty"`
	ProxyType          string `yaml:"proxyType,omitempty"`
}

// WarpRouting enables routing private network traffic through the tunnel.
type WarpRouting struct {
	Enabled bool `yaml:"enabled"`
}

// IsCatchAll returns true if this is a catch-all rule (no hostname).
func (r IngressRule) IsCatchAll() bool {
	return r.Hostname == "" && r.Path == ""
}

// TunnelInfo represents parsed tunnel information from cloudflared output.
type TunnelInfo struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CreatedAt  string `json:"created_at"`
	DeletedAt  string `json:"deleted_at,omitempty"`
	ConnectorID string `json:"connector_id,omitempty"`
}
