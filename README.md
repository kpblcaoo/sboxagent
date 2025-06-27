# Subbox Agent (sboxagent)

A Go daemon for automatic subscription management, logging, and service monitoring in the Subbox ecosystem.

## 🎯 Purpose

Subbox Agent (`sboxagent`) is a lightweight Go daemon that provides:

- **Automatic Updates** - Scheduled execution of Subbox Manager (`sboxctl`) commands
- **Service Management** - Monitoring and control of VPN clients (sing-box, xray, clash, hysteria)
- **Logging & Monitoring** - Structured logging with aggregation and health checks
- **HTTP API** - REST API for configuration updates and status monitoring
- **Security** - Sandboxed execution with configurable access controls

## 🏗 Architecture

```
sboxagent/
├── cmd/sboxagent/        # Main application entry point
├── internal/             # Private application code
│   ├── agent/           # Core agent logic
│   ├── api/             # HTTP API handlers
│   ├── config/          # Configuration management
│   ├── logger/          # Structured logging
│   ├── services/        # Service management
│   └── security/        # Security and sandboxing
├── pkg/                 # Public packages
│   ├── schemas/         # Embedded JSON schemas
│   └── protocols/       # API protocol definitions
├── scripts/             # Build and deployment scripts
├── examples/            # Configuration examples
├── docs/                # Documentation
└── tests/               # Test files
```

## 🚀 Quick Start

### Installation

```bash
# Build from source
go build -o sboxagent ./cmd/sboxagent

# Or install directly
go install ./cmd/sboxagent@latest
```

### Configuration

Create a configuration file `config.yaml`:

```yaml
agent:
  name: "home-server"
  version: "0.1.0"
  log_level: "info"

server:
  port: 8080
  host: "127.0.0.1"
  timeout: "30s"

services:
  sboxctl:
    enabled: true
    command: ["sboxctl", "update"]
    interval: "30m"
    timeout: "5m"
    stdout_capture: true
    health_check:
      enabled: true
      interval: "1m"
      timeout: "10s"

clients:
  sing-box:
    enabled: true
    binary_path: "/usr/local/bin/sing-box"
    config_path: "/etc/sing-box/config.json"

logging:
  stdout_capture: true
  aggregation: true
  retention_days: 30
  max_entries: 1000

security:
  allow_remote_api: false
  api_token: "your-secure-token-here"
  allowed_hosts: ["127.0.0.1", "::1"]
  tls_enabled: false
```

### Running

```bash
# Start with configuration file
./sboxagent -config config.yaml

# Start with default settings
./sboxagent

# Run in foreground for debugging
./sboxagent -debug
```

## 🔧 Configuration

### Agent Settings

- `name` - Agent identifier for logging and monitoring
- `version` - Agent version for compatibility checks
- `log_level` - Logging verbosity (debug, info, warn, error)

### Server Configuration

- `port` - HTTP API server port (1-65535)
- `host` - Server bind address
- `timeout` - Request timeout (e.g., "30s", "5m")

### Service Management

- `sboxctl.enabled` - Enable Subbox Manager integration
- `sboxctl.command` - Command and arguments to execute
- `sboxctl.interval` - Update frequency (e.g., "30m", "1h")
- `sboxctl.timeout` - Command execution timeout
- `sboxctl.health_check` - Health monitoring settings

### VPN Client Support

Supported clients with automatic configuration management:

- **sing-box** - Universal proxy platform
- **xray** - Xray-core proxy
- **clash** - Clash proxy
- **hysteria** - Hysteria proxy

Each client can be enabled/disabled and configured with custom paths.

### Security Settings

- `allow_remote_api` - Allow external API access
- `api_token` - Authentication token for API requests
- `allowed_hosts` - List of permitted client IPs
- `tls_enabled` - Enable TLS encryption
- `tls_cert_file` - TLS certificate path
- `tls_key_file` - TLS private key path

## 🌐 HTTP API

### Endpoints

- `GET /api/v1/status` - Get agent status and health
- `POST /api/v1/config` - Update agent configuration
- `GET /api/v1/logs` - Retrieve aggregated logs
- `GET /api/v1/health` - Health check endpoint

### Authentication

API requests require authentication via:
- `Authorization: Bearer <api_token>` header
- Or `X-API-Token: <api_token>` header

### Example Requests

```bash
# Get status
curl -H "Authorization: Bearer your-token" \
     http://localhost:8080/api/v1/status

# Update configuration
curl -X POST \
     -H "Authorization: Bearer your-token" \
     -H "Content-Type: application/json" \
     -d @config.json \
     http://localhost:8080/api/v1/config
```

## 📊 Monitoring

### Health Checks

- **Agent Health** - Overall daemon status
- **Service Health** - Managed service status
- **API Health** - HTTP endpoint availability
- **Resource Usage** - Memory and CPU monitoring

### Logging

- **Structured Logs** - JSON format with metadata
- **Log Aggregation** - Centralized log collection
- **Retention Policy** - Configurable log retention
- **Log Levels** - Debug, info, warn, error levels

### Metrics

- **Service Uptime** - Service availability tracking
- **Update Success Rate** - Command execution statistics
- **API Request Metrics** - Request/response statistics
- **Resource Metrics** - System resource utilization

## 🔒 Security

### Sandboxing

- **Process Isolation** - Commands run in isolated environment
- **Resource Limits** - CPU and memory constraints
- **File System Access** - Restricted file system access
- **Network Access** - Controlled network connectivity

### Access Control

- **API Authentication** - Token-based authentication
- **IP Whitelisting** - Configurable client IP restrictions
- **TLS Encryption** - Optional transport encryption
- **Audit Logging** - Security event logging

## 🧪 Development

### Building

```bash
# Build for current platform
go build -o sboxagent ./cmd/sboxagent

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o sboxagent ./cmd/sboxagent

# Build with debug symbols
go build -gcflags="all=-N -l" -o sboxagent ./cmd/sboxagent
```

### Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

### Development Mode

```bash
# Run with hot reload (requires air)
air

# Run with debug logging
./sboxagent -debug -log-level=debug

# Run with custom config
./sboxagent -config dev-config.yaml
```

## 📦 Deployment

### Systemd Service

Create `/etc/systemd/system/sboxagent.service`:

```ini
[Unit]
Description=Subbox Agent
After=network.target

[Service]
Type=simple
User=sboxagent
Group=sboxagent
ExecStart=/usr/local/bin/sboxagent -config /etc/sboxagent/config.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o sboxagent ./cmd/sboxagent

FROM alpine:latest
RUN addgroup -g 1000 sboxagent && \
    adduser -D -s /bin/sh -u 1000 -G sboxagent sboxagent
COPY --from=builder /app/sboxagent /usr/local/bin/
USER sboxagent
EXPOSE 8080
ENTRYPOINT ["sboxagent"]
```

## 🔗 Integration

### With Subbox Manager (sboxctl)

sboxagent integrates with Subbox Manager for automatic updates:

```yaml
services:
  sboxctl:
    enabled: true
    command: ["sboxctl", "update", "--auto"]
    interval: "30m"
```

### With Monitoring Systems

- **Prometheus** - Metrics endpoint at `/metrics`
- **Grafana** - Pre-built dashboards available
- **AlertManager** - Alerting integration
- **ELK Stack** - Log aggregation support

## 📚 Documentation

- [Configuration Reference](docs/configuration.md)
- [API Reference](docs/api.md)
- [Security Guide](docs/security.md)
- [Deployment Guide](docs/deployment.md)
- [Troubleshooting](docs/troubleshooting.md)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## 📄 License

GPL-3.0 - see [LICENSE](LICENSE) file. 