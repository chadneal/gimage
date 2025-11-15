# gimage-deploy

A comprehensive CLI tool with Bubbletea TUI for managing gimage Lambda deployments and API Gateway API keys.

## Features

- **Deployment Management**: Deploy, update, monitor, and destroy gimage Lambda instances
- **API Key Management**: Full CRUD operations for API Gateway API keys
- **Interactive TUI**: Beautiful terminal UI powered by Bubbletea
- **Headless CLI**: Automation-friendly commands for CI/CD
- **Real-time Monitoring**: CloudWatch metrics, logs, and health checks
- **Secure Storage**: Encrypted local storage for API keys (AES-256-GCM)

## Installation

```bash
# Clone the repository
git clone https://github.com/apresai/gimage-deploy.git
cd gimage-deploy

# Build and install
make install
```

## Quick Start

### Interactive TUI Mode

```bash
# Launch interactive terminal UI
gimage-deploy tui
```

### CLI Mode

```bash
# List all deployments
gimage-deploy list

# Create a new deployment
gimage-deploy deploy --id prod-001 --stage prod --region us-east-1

# List API keys
gimage-deploy keys list

# Create an API key
gimage-deploy keys create --name web-app --deployment prod-001

# View configuration
gimage-deploy config get
```

## Commands

- `gimage-deploy tui` - Launch interactive TUI
- `gimage-deploy deploy` - Create new deployment
- `gimage-deploy list` - List all deployments
- `gimage-deploy keys` - Manage API keys
  - `keys list` - List all API keys
  - `keys create` - Create new API key
  - `keys delete <id>` - Delete API key
- `gimage-deploy config` - Manage configuration
  - `config get [key]` - View configuration
  - `config set <key> <value>` - Update configuration
  - `config reset` - Reset to defaults
- `gimage-deploy version` - Show version

## Configuration

Configuration is stored in `~/.gimage-deploy/config.json`.

Default settings:
- AWS Profile: `default`
- Default Region: `us-east-1`
- Default Stage: `production`
- Default Memory: `512 MB`
- Default Timeout: `30 seconds`
- Default Concurrency: `10`

## Development

```bash
# Install dependencies
make deps

# Run tests
make test

# Run tests with coverage
make test-coverage

# Build
make build

# Build for all platforms
make build-all

# Run linter
make lint

# Format code
make fmt
```

## Architecture

```
gimage-deploy/
├── cmd/gimage-deploy/      # CLI entrypoint
├── internal/
│   ├── aws/                # AWS SDK wrappers
│   ├── deploy/             # Deployment management
│   ├── apikeys/            # API key management
│   ├── monitoring/         # Metrics and logs
│   ├── storage/            # Local storage
│   ├── cli/                # CLI commands
│   ├── tui/                # Bubbletea TUI
│   └── models/             # Data models
├── pkg/utils/              # Utilities
└── test/                   # Tests
```

## Security

- API keys are encrypted using AES-256-GCM before storage
- Configuration files are created with 0600 permissions (owner read/write only)
- Supports AWS IAM roles and profiles
- Never logs sensitive credentials

## License

MIT

## Contributing

Contributions welcome! Please open an issue or submit a pull request.
