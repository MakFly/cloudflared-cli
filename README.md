# cloudflared-project

[![Go Version](https://img.shields.io/github/go-mod/go-version/MakFly/cloudflared-cli)](https://go.dev/)
[![License](https://img.shields.io/github/license/MakFly/cloudflared-cli)](LICENSE)
[![Release](https://img.shields.io/github/v/release/MakFly/cloudflared-cli)](https://github.com/MakFly/cloudflared-cli/releases)

> Production-grade CLI wrapper for Cloudflare Tunnel management.

`cloudflared-project` simplifies multi-environment Cloudflare Tunnel workflows — project scaffolding, per-env configs, one-command deploys, and DNS routing — all backed by the official `cloudflared` binary.

## Features

- **Multi-environment** — Manage `dev`, `staging`, and `prod` tunnel configs in one project
- **Project scaffolding** — `init` generates a complete project structure with sensible defaults
- **Config management** — Add/remove ingress rules, set tunnel parameters, validate before deploy
- **One-command deploy** — Validate, route DNS, and start the tunnel in a single step
- **Background mode** — Run tunnels detached with built-in log tailing
- **Zero lock-in** — Generates standard `cloudflared` config files you can use directly

## Quick Install

```bash
# One-line installer
curl -fsSL https://raw.githubusercontent.com/MakFly/cloudflared-cli/main/scripts/install.sh | bash

# Or via go install
go install github.com/MakFly/cloudflared-cli@latest

# Or from source
git clone https://github.com/MakFly/cloudflared-cli.git
cd cloudflared-cli
make build
```

## Quick Start

```bash
# 1. Authenticate with Cloudflare
cloudflared-project login

# 2. Initialize a project
cloudflared-project init myapp --domain app.example.com

# 3. Create a tunnel
cloudflared-project -p myapp tunnel create myapp

# 4. Review configuration
cloudflared-project -p myapp config show

# 5. Deploy with DNS routing
cloudflared-project -p myapp deploy --route-dns
```

## Commands

| Command | Description |
|---------|-------------|
| `init <name>` | Initialize a new project |
| `login` | Authenticate with Cloudflare |
| `tunnel create <name>` | Create a new tunnel |
| `tunnel list` | List all tunnels |
| `tunnel delete <name>` | Delete a tunnel |
| `tunnel info <name>` | Show tunnel details |
| `config show` | Display current config |
| `config set <key> <value>` | Set a config value |
| `config add-ingress` | Add an ingress rule |
| `config remove-ingress` | Remove an ingress rule |
| `config validate` | Validate tunnel config |
| `deploy` | Validate, route DNS, and start tunnel |
| `status` | Show tunnel status |
| `logs` | Tail tunnel logs |
| `version` | Print version info |

## Documentation

Full documentation is available at **[cloudflared.pulseview.app](https://cloudflared.pulseview.app)**.

## Prerequisites

- [Go 1.26+](https://go.dev/dl/) (for building from source)
- [`cloudflared`](https://developers.cloudflare.com/cloudflare-one/connections/connect-networks/downloads/) binary installed and in `PATH`
- A Cloudflare account with a configured domain

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

```bash
# Run tests
make test

# Run linter
make lint

# Build for all platforms
make cross
```

## License

[MIT](LICENSE)
