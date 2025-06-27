#!/bin/bash

# SboxAgent uninstallation script

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

# Check if running as root
if [[ $EUID -ne 0 ]]; then
   echo -e "${RED}This script must be run as root${NC}"
   exit 1
fi

# Stop and disable service
if systemctl is-active --quiet sboxagent.service; then
    echo -e "${YELLOW}Stopping service...${NC}"
    systemctl stop sboxagent.service
fi

# Disable service
if systemctl is-enabled --quiet sboxagent.service; then
    echo -e "${YELLOW}Disabling service...${NC}"
    systemctl disable sboxagent.service
fi

# Remove systemd service file
if [ -f $SERVICE_DIR/sboxagent.service ]; then
    echo -e "${YELLOW}Removing systemd service...${NC}"
    rm -f $SERVICE_DIR/sboxagent.service
    systemctl daemon-reload
fi

# Remove binary
if [ -f $INSTALL_DIR/$BINARY_NAME ]; then
    echo -e "${YELLOW}Removing binary...${NC}"
    rm -f $INSTALL_DIR/$BINARY_NAME
fi

# Remove configuration (optional)
read -p "Remove configuration directory $CONFIG_DIR? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Removing configuration...${NC}"
    rm -rf $CONFIG_DIR
fi

# Remove user and group (optional)
read -p "Remove user $USER_NAME and group $GROUP_NAME? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Removing user and group...${NC}"
    userdel $USER_NAME 2>/dev/null || true
    groupdel $GROUP_NAME 2>/dev/null || true
fi

echo -e "${GREEN}Uninstallation complete!${NC}"
