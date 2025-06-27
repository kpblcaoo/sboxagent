# SboxMgr

Python CLI tool for managing sing-box proxy subscriptions with sboxagent integration.

## Overview

SboxMgr is a Python CLI tool that manages sing-box proxy subscriptions and integrates with sboxagent daemon via Unix socket IPC for real-time configuration updates.

## Features

- Subscription management for sing-box proxies
- Unix socket IPC with sboxagent daemon
- Framed JSON protocol communication
- Event-driven architecture
- Configuration validation
- Structured logging

## Installation

### Prerequisites
- Python >= 3.8
- sbox-common package
- sboxagent daemon

### Development Install
```bash
# Install sbox-common in editable mode
cd ../sbox-common
pip install -e .

# Install sboxmgr in editable mode
cd ../sboxmgr
pip install -e .
```

### Production Install
```bash
pip install sboxmgr
```

## Architecture

```
sboxmgr/
├── sboxmgr/
│   ├── __init__.py
│   ├── cli.py              # CLI entry point
│   ├── agent/
│   │   ├── __init__.py
│   │   └── ipc/
│   │       ├── __init__.py
│   │       └── socket_client.py  # Unix socket client
│   ├── config/             # Configuration management
│   ├── events/             # Event handling
│   └── utils/              # Utilities
├── tests/                  # Unit and integration tests
└── docs/                   # Documentation
```

## IPC Integration

SboxMgr communicates with sboxagent via Unix socket using the framed JSON protocol:

### Socket Client Usage

```python
from sboxmgr.agent.ipc.socket_client import SocketClient

# Create client
client = SocketClient('/tmp/sboxagent.sock')

# Connect to sboxagent
client.connect()

# Send event message
event_msg = client.protocol.create_event_message({
    "type": "config_updated",
    "data": {"config_id": "test-123"}
})
client.send_message(event_msg)

# Receive response
response = client.recv_message()
print(f"Response ID: {response['id']}")

# Close connection
client.close()
```

### Protocol Features

- **Framed JSON**: Reliable message framing with length headers
- **Event Messages**: Notifications and status updates
- **Command Messages**: Request execution of actions
- **Response Messages**: Replies to commands
- **Heartbeat Messages**: Health and status information

## Usage

### Basic Commands

```bash
# Show help
sboxmgr --help

# List subscriptions
sboxmgr list

# Add subscription
sboxmgr add --name "my-sub" --url "https://example.com/sub"

# Update subscription
sboxmgr update --name "my-sub"

# Remove subscription
sboxmgr remove --name "my-sub"

# Show status
sboxmgr status
```

### Agent Integration

```bash
# Start sboxagent daemon
sboxagent --socket /tmp/sboxagent.sock &

# Use sboxmgr with agent integration
sboxmgr --agent-socket /tmp/sboxagent.sock list

# Stop agent
kill %1
```

## Development

### Dependencies
- Python >= 3.8
- sbox-common (for protocols)
- click (for CLI)
- jsonschema (for validation)

### Testing
```bash
# Unit tests
pytest tests/

# Integration tests
pytest tests/integration/

# With coverage
pytest --cov=sboxmgr tests/
```

### Integration Testing
```bash
# Start sboxagent
cd ../sboxagent
./sboxagent --socket /tmp/test.sock &

# Run integration tests
cd ../sboxmgr
pytest tests/integration/ -v

# Cleanup
kill %1
rm /tmp/test.sock
```

## Configuration

SboxMgr supports configuration via:

- Command-line arguments
- Environment variables
- Configuration files

### Environment Variables

- `SBOX_AGENT_SOCKET`: Default agent socket path
- `SBOX_LOG_LEVEL`: Logging level
- `SBOX_CONFIG_DIR`: Configuration directory

## License

Apache-2.0 