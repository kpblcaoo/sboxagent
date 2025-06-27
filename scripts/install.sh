#!/bin/bash

# SboxAgent installation script
# Version: 0.1.0-alpha 

set -e

# Colors for output
RED="[0;31m"
GREEN="[0;32m"
YELLOW="[1;33m"
NC="[0m"

# Configuration
BINARY_NAME="sboxagent"
INSTALL_DIR="/usr/local/bin"
SERVICE_DIR="/etc/systemd/system"
CONFIG_DIR="/etc/sboxagent"
USER_NAME="sboxagent"
GROUP_NAME="sboxagent"

echo -e "${GREEN}Installing SboxAgent v0.1.0-alpha ...${NC}"

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

# Create user and group if they do not exist
if ! id "$USER_NAME" &>/dev/null; then
    echo -e "${YELLOW}Creating user $USER_NAME...${NC}"
    useradd --system --no-create-home --shell /bin/false $USER_NAME
else
    echo -e "${GREEN}User $USER_NAME already exists${NC}"
fi

if ! getent group "$GROUP_NAME" &>/dev/null; then
    echo -e "${YELLOW}Creating group $GROUP_NAME...${NC}"
    groupadd --system $GROUP_NAME
    usermod -a -G $GROUP_NAME $USER_NAME
else
    echo -e "${GREEN}Group $GROUP_NAME already exists${NC}"
fi

# Create configuration directory
echo -e "${YELLOW}Creating configuration directory...${NC}"
mkdir -p $CONFIG_DIR
chown $USER_NAME:$GROUP_NAME $CONFIG_DIR
chmod 755 $CONFIG_DIR

# Copy binary
echo -e "${YELLOW}Installing binary...${NC}"
cp bin/$BINARY_NAME $INSTALL_DIR/
chown $USER_NAME:$GROUP_NAME $INSTALL_DIR/$BINARY_NAME
chmod 755 $INSTALL_DIR/$BINARY_NAME

# Install systemd service
echo -e "${YELLOW}Installing systemd service...${NC}"
cp scripts/sboxagent.service $SERVICE_DIR/
systemctl daemon-reload

# Create default configuration if it does not exist
if [ ! -f $CONFIG_DIR/agent.yaml ]; then
    echo -e "${YELLOW}Creating default configuration...${NC}"
    cat > $CONFIG_DIR/agent.yaml << EOF
# SboxAgent Configuration
# Version: 0.1.0-alpha 

# Agent configuration
agent:
  name: "sboxagent"
  version: "0.1.0-alpha "
  log_level: "info"
  log_format: "json"

# Sboxctl service configuration
sboxctl:
  command: ["sboxctl", "status"]
  interval: "30s"
  timeout: "10s"
  stdout_capture: true
  health_check:
    enabled: true
    interval: "60s"

# Log aggregator configuration
aggregator:
  max_entries: 1000
  max_age: "24h"

# Health checker configuration
health:
  check_interval: "30s"
  timeout: "5s"
EOF
    chown $USER_NAME:$GROUP_NAME $CONFIG_DIR/agent.yaml
    chmod 644 $CONFIG_DIR/agent.yaml
fi

# Enable and start service
echo -e "${YELLOW}Enabling and starting service...${NC}"
systemctl enable sboxagent.service
systemctl start sboxagent.service

# Check service status
if systemctl is-active --quiet sboxagent.service; then
    echo -e "${GREEN}SboxAgent service is running${NC}"
else
    echo -e "${RED}SboxAgent service failed to start${NC}"
    systemctl status sboxagent.service
    exit 1
fi

echo -e "${GREEN}Installation complete!${NC}"
echo -e "${YELLOW}Configuration: $CONFIG_DIR/agent.yaml${NC}"
echo -e "${YELLOW}Service: systemctl status sboxagent${NC}"
echo -e "${YELLOW}Logs: journalctl -u sboxagent -f${NC}"
