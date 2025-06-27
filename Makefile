# Makefile for sboxagent
# Version: 0.1.0-alpha

# Variables
BINARY_NAME=sboxagent
BINARY_UNIX=$(BINARY_NAME)_unix
VERSION=$(shell cat VERSION)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD)
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Directories
BIN_DIR=bin
INSTALL_DIR=/usr/local/bin
SERVICE_DIR=/etc/systemd/system
CONFIG_DIR=/etc/sboxagent
USER_NAME=sboxagent
GROUP_NAME=sboxagent
DOCS_DIR=docs

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GODOC=$(GOCMD) doc
BINARY_PATH=$(BIN_DIR)/$(BINARY_NAME)

# Default target
.DEFAULT_GOAL := build

# Build the application
.PHONY: build
build: clean
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BIN_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) ./cmd/sboxagent
	@echo "Build complete: $(BINARY_PATH)"

# Build for Linux
.PHONY: build-linux
build-linux: clean
	@echo "Building $(BINARY_NAME) for Linux v$(VERSION)..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_UNIX) ./cmd/sboxagent
	@echo "Build complete: $(BIN_DIR)/$(BINARY_UNIX)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BIN_DIR)
	@$(GOCLEAN)
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...
	@echo "Tests complete"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	@echo "Coverage report:"
	@go tool cover -func=coverage.out | tail -1
	@echo "Tests with coverage complete"

# Run tests with coverage and generate HTML report
.PHONY: test-coverage-html
test-coverage-html: test-coverage
	@echo "Generating HTML coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "HTML coverage report: coverage.html"

# Run integration tests
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -v ./tests/integration/...
	@echo "Integration tests complete"

# Run benchmarks
.PHONY: benchmark
benchmark:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. ./...
	@echo "Benchmarks complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy
	@echo "Dependencies installed"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Code formatting complete"

# Run linter
.PHONY: lint
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found, skipping linting"; \
	fi
	@echo "Linting complete"

# Check code quality
.PHONY: check
check: fmt lint test
	@echo "Code quality check complete"

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@mkdir -p $(DOCS_DIR)
	@echo "# SboxAgent Documentation" > $(DOCS_DIR)/README.md
	@echo "" >> $(DOCS_DIR)/README.md
	@echo "Generated on: $(shell date)" >> $(DOCS_DIR)/README.md
	@echo "Version: $(VERSION)" >> $(DOCS_DIR)/README.md
	@echo "" >> $(DOCS_DIR)/README.md
	@echo "## Package Documentation" >> $(DOCS_DIR)/README.md
	@echo "" >> $(DOCS_DIR)/README.md
	@echo "### Main Package" >> $(DOCS_DIR)/README.md
	@echo '```' >> $(DOCS_DIR)/README.md
	@$(GODOC) ./cmd/sboxagent >> $(DOCS_DIR)/README.md 2>/dev/null || echo "No documentation available" >> $(DOCS_DIR)/README.md
	@echo '```' >> $(DOCS_DIR)/README.md
	@echo "" >> $(DOCS_DIR)/README.md
	@echo "### Internal Packages" >> $(DOCS_DIR)/README.md
	@echo "" >> $(DOCS_DIR)/README.md
	@for pkg in internal/*; do \
		if [ -d "$$pkg" ]; then \
			echo "#### $$(basename $$pkg)" >> $(DOCS_DIR)/README.md; \
			echo '```' >> $(DOCS_DIR)/README.md; \
			$(GODOC) ./$$pkg >> $(DOCS_DIR)/README.md 2>/dev/null || echo "No documentation available" >> $(DOCS_DIR)/README.md; \
			echo '```' >> $(DOCS_DIR)/README.md; \
			echo "" >> $(DOCS_DIR)/README.md; \
		fi; \
	done
	@echo "Documentation generated: $(DOCS_DIR)/README.md"

# Check documentation coverage
.PHONY: docs-check
docs-check:
	@echo "Checking documentation coverage..."
	@echo "Checking exported symbols without documentation..."
	@for pkg in internal/*; do \
		if [ -d "$$pkg" ]; then \
			echo "Package: $$(basename $$pkg)"; \
			$(GODOC) -short ./$$pkg 2>/dev/null | grep -E "^[A-Z]" | while read symbol; do \
				if ! $(GODOC) ./$$pkg | grep -q "$$symbol"; then \
					echo "  WARNING: $$symbol has no documentation"; \
				fi; \
			done; \
		fi; \
	done
	@echo "Documentation check complete"

# Start documentation server
.PHONY: docs-serve
docs-serve:
	@echo "Starting documentation server at http://localhost:6060"
	@echo "Press Ctrl+C to stop"
	@echo "Note: If godoc fails, use 'go doc ./...' for command-line documentation"
	@godoc -http=:6060 2>/dev/null || (echo "godoc failed, starting alternative server..." && python3 -m http.server 6060 --directory docs/ 2>/dev/null || echo "Please install python3 or use 'go doc ./...' for documentation")

# Create systemd service file
.PHONY: service
service:
	@echo "Creating systemd service file..."
	@mkdir -p scripts
	@echo '[Unit]' > scripts/sboxagent.service
	@echo 'Description=SboxAgent - sing-box proxy configuration manager' >> scripts/sboxagent.service
	@echo 'Documentation=https://github.com/kpblcaoo/sboxagent' >> scripts/sboxagent.service
	@echo 'After=network.target' >> scripts/sboxagent.service
	@echo '' >> scripts/sboxagent.service
	@echo '[Service]' >> scripts/sboxagent.service
	@echo 'Type=simple' >> scripts/sboxagent.service
	@echo 'User=$(USER_NAME)' >> scripts/sboxagent.service
	@echo 'Group=$(GROUP_NAME)' >> scripts/sboxagent.service
	@echo 'ExecStart=$(INSTALL_DIR)/$(BINARY_NAME)' >> scripts/sboxagent.service
	@echo 'Restart=always' >> scripts/sboxagent.service
	@echo 'RestartSec=5' >> scripts/sboxagent.service
	@echo 'StandardOutput=journal' >> scripts/sboxagent.service
	@echo 'StandardError=journal' >> scripts/sboxagent.service
	@echo 'SyslogIdentifier=sboxagent' >> scripts/sboxagent.service
	@echo '' >> scripts/sboxagent.service
	@echo '# Security settings' >> scripts/sboxagent.service
	@echo 'NoNewPrivileges=true' >> scripts/sboxagent.service
	@echo 'PrivateTmp=true' >> scripts/sboxagent.service
	@echo 'ProtectSystem=strict' >> scripts/sboxagent.service
	@echo 'ProtectHome=true' >> scripts/sboxagent.service
	@echo 'ReadWritePaths=$(CONFIG_DIR)' >> scripts/sboxagent.service
	@echo '' >> scripts/sboxagent.service
	@echo '# Resource limits' >> scripts/sboxagent.service
	@echo 'LimitNOFILE=65536' >> scripts/sboxagent.service
	@echo 'LimitNPROC=4096' >> scripts/sboxagent.service
	@echo '' >> scripts/sboxagent.service
	@echo '[Install]' >> scripts/sboxagent.service
	@echo 'WantedBy=multi-user.target' >> scripts/sboxagent.service
	@echo "Service file created: scripts/sboxagent.service"

# Create install script
.PHONY: install-script
install-script: service
	@echo "Creating install script..."
	@mkdir -p scripts
	@echo '#!/bin/bash' > scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# SboxAgent installation script' >> scripts/install.sh
	@echo '# Version: $(VERSION)' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo 'set -e' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Colors for output' >> scripts/install.sh
	@echo 'RED="\033[0;31m"' >> scripts/install.sh
	@echo 'GREEN="\033[0;32m"' >> scripts/install.sh
	@echo 'YELLOW="\033[1;33m"' >> scripts/install.sh
	@echo 'NC="\033[0m"' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Configuration' >> scripts/install.sh
	@echo 'BINARY_NAME="$(BINARY_NAME)"' >> scripts/install.sh
	@echo 'INSTALL_DIR="$(INSTALL_DIR)"' >> scripts/install.sh
	@echo 'SERVICE_DIR="$(SERVICE_DIR)"' >> scripts/install.sh
	@echo 'CONFIG_DIR="$(CONFIG_DIR)"' >> scripts/install.sh
	@echo 'USER_NAME="$(USER_NAME)"' >> scripts/install.sh
	@echo 'GROUP_NAME="$(GROUP_NAME)"' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo 'echo -e "$${GREEN}Installing SboxAgent v$(VERSION)...$${NC}"' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Check if running as root' >> scripts/install.sh
	@echo 'if [[ $$EUID -ne 0 ]]; then' >> scripts/install.sh
	@echo '   echo -e "$${RED}This script must be run as root$${NC}"' >> scripts/install.sh
	@echo '   exit 1' >> scripts/install.sh
	@echo 'fi' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Create user and group if they do not exist' >> scripts/install.sh
	@echo 'if ! id "$$USER_NAME" &>/dev/null; then' >> scripts/install.sh
	@echo '    echo -e "$${YELLOW}Creating user $$USER_NAME...$${NC}"' >> scripts/install.sh
	@echo '    useradd --system --no-create-home --shell /bin/false $$USER_NAME' >> scripts/install.sh
	@echo 'else' >> scripts/install.sh
	@echo '    echo -e "$${GREEN}User $$USER_NAME already exists$${NC}"' >> scripts/install.sh
	@echo 'fi' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo 'if ! getent group "$$GROUP_NAME" &>/dev/null; then' >> scripts/install.sh
	@echo '    echo -e "$${YELLOW}Creating group $$GROUP_NAME...$${NC}"' >> scripts/install.sh
	@echo '    groupadd --system $$GROUP_NAME' >> scripts/install.sh
	@echo '    usermod -a -G $$GROUP_NAME $$USER_NAME' >> scripts/install.sh
	@echo 'else' >> scripts/install.sh
	@echo '    echo -e "$${GREEN}Group $$GROUP_NAME already exists$${NC}"' >> scripts/install.sh
	@echo 'fi' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Create configuration directory' >> scripts/install.sh
	@echo 'echo -e "$${YELLOW}Creating configuration directory...$${NC}"' >> scripts/install.sh
	@echo 'mkdir -p $$CONFIG_DIR' >> scripts/install.sh
	@echo 'chown $$USER_NAME:$$GROUP_NAME $$CONFIG_DIR' >> scripts/install.sh
	@echo 'chmod 755 $$CONFIG_DIR' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Copy binary' >> scripts/install.sh
	@echo 'echo -e "$${YELLOW}Installing binary...$${NC}"' >> scripts/install.sh
	@echo 'cp bin/$$BINARY_NAME $$INSTALL_DIR/' >> scripts/install.sh
	@echo 'chown $$USER_NAME:$$GROUP_NAME $$INSTALL_DIR/$$BINARY_NAME' >> scripts/install.sh
	@echo 'chmod 755 $$INSTALL_DIR/$$BINARY_NAME' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Install systemd service' >> scripts/install.sh
	@echo 'echo -e "$${YELLOW}Installing systemd service...$${NC}"' >> scripts/install.sh
	@echo 'cp scripts/sboxagent.service $$SERVICE_DIR/' >> scripts/install.sh
	@echo 'systemctl daemon-reload' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Create default configuration if it does not exist' >> scripts/install.sh
	@echo 'if [ ! -f $$CONFIG_DIR/agent.yaml ]; then' >> scripts/install.sh
	@echo '    echo -e "$${YELLOW}Creating default configuration...$${NC}"' >> scripts/install.sh
	@echo '    cat > $$CONFIG_DIR/agent.yaml << EOF' >> scripts/install.sh
	@echo '# SboxAgent Configuration' >> scripts/install.sh
	@echo '# Version: $(VERSION)' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Agent configuration' >> scripts/install.sh
	@echo 'agent:' >> scripts/install.sh
	@echo '  name: "sboxagent"' >> scripts/install.sh
	@echo '  version: "$(VERSION)"' >> scripts/install.sh
	@echo '  log_level: "info"' >> scripts/install.sh
	@echo '  log_format: "json"' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Sboxctl service configuration' >> scripts/install.sh
	@echo 'sboxctl:' >> scripts/install.sh
	@echo '  command: ["sboxctl", "status"]' >> scripts/install.sh
	@echo '  interval: "30s"' >> scripts/install.sh
	@echo '  timeout: "10s"' >> scripts/install.sh
	@echo '  stdout_capture: true' >> scripts/install.sh
	@echo '  health_check:' >> scripts/install.sh
	@echo '    enabled: true' >> scripts/install.sh
	@echo '    interval: "60s"' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Log aggregator configuration' >> scripts/install.sh
	@echo 'aggregator:' >> scripts/install.sh
	@echo '  max_entries: 1000' >> scripts/install.sh
	@echo '  max_age: "24h"' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Health checker configuration' >> scripts/install.sh
	@echo 'health:' >> scripts/install.sh
	@echo '  check_interval: "30s"' >> scripts/install.sh
	@echo '  timeout: "5s"' >> scripts/install.sh
	@echo 'EOF' >> scripts/install.sh
	@echo '    chown $$USER_NAME:$$GROUP_NAME $$CONFIG_DIR/agent.yaml' >> scripts/install.sh
	@echo '    chmod 644 $$CONFIG_DIR/agent.yaml' >> scripts/install.sh
	@echo 'fi' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Enable and start service' >> scripts/install.sh
	@echo 'echo -e "$${YELLOW}Enabling and starting service...$${NC}"' >> scripts/install.sh
	@echo 'systemctl enable sboxagent.service' >> scripts/install.sh
	@echo 'systemctl start sboxagent.service' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo '# Check service status' >> scripts/install.sh
	@echo 'if systemctl is-active --quiet sboxagent.service; then' >> scripts/install.sh
	@echo '    echo -e "$${GREEN}SboxAgent service is running$${NC}"' >> scripts/install.sh
	@echo 'else' >> scripts/install.sh
	@echo '    echo -e "$${RED}SboxAgent service failed to start$${NC}"' >> scripts/install.sh
	@echo '    systemctl status sboxagent.service' >> scripts/install.sh
	@echo '    exit 1' >> scripts/install.sh
	@echo 'fi' >> scripts/install.sh
	@echo '' >> scripts/install.sh
	@echo 'echo -e "$${GREEN}Installation complete!$${NC}"' >> scripts/install.sh
	@echo 'echo -e "$${YELLOW}Configuration: $$CONFIG_DIR/agent.yaml$${NC}"' >> scripts/install.sh
	@echo 'echo -e "$${YELLOW}Service: systemctl status sboxagent$${NC}"' >> scripts/install.sh
	@echo 'echo -e "$${YELLOW}Logs: journalctl -u sboxagent -f$${NC}"' >> scripts/install.sh
	@chmod +x scripts/install.sh
	@echo "Install script created: scripts/install.sh"

# Create uninstall script
.PHONY: uninstall-script
uninstall-script:
	@echo "Creating uninstall script..."
	@mkdir -p scripts
	@echo '#!/bin/bash' > scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# SboxAgent uninstallation script' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo 'set -e' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Colors for output' >> scripts/uninstall.sh
	@echo 'RED="\033[0;31m"' >> scripts/uninstall.sh
	@echo 'GREEN="\033[0;32m"' >> scripts/uninstall.sh
	@echo 'YELLOW="\033[1;33m"' >> scripts/uninstall.sh
	@echo 'NC="\033[0m"' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Configuration' >> scripts/uninstall.sh
	@echo 'BINARY_NAME="$(BINARY_NAME)"' >> scripts/uninstall.sh
	@echo 'INSTALL_DIR="$(INSTALL_DIR)"' >> scripts/uninstall.sh
	@echo 'SERVICE_DIR="$(SERVICE_DIR)"' >> scripts/uninstall.sh
	@echo 'CONFIG_DIR="$(CONFIG_DIR)"' >> scripts/uninstall.sh
	@echo 'USER_NAME="$(USER_NAME)"' >> scripts/uninstall.sh
	@echo 'GROUP_NAME="$(GROUP_NAME)"' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo 'echo -e "$${YELLOW}Uninstalling SboxAgent...$${NC}"' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Check if running as root' >> scripts/uninstall.sh
	@echo 'if [[ $$EUID -ne 0 ]]; then' >> scripts/uninstall.sh
	@echo '   echo -e "$${RED}This script must be run as root$${NC}"' >> scripts/uninstall.sh
	@echo '   exit 1' >> scripts/uninstall.sh
	@echo 'fi' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Stop and disable service' >> scripts/uninstall.sh
	@echo 'if systemctl is-active --quiet sboxagent.service; then' >> scripts/uninstall.sh
	@echo '    echo -e "$${YELLOW}Stopping service...$${NC}"' >> scripts/uninstall.sh
	@echo '    systemctl stop sboxagent.service' >> scripts/uninstall.sh
	@echo 'fi' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Disable service' >> scripts/uninstall.sh
	@echo 'if systemctl is-enabled --quiet sboxagent.service; then' >> scripts/uninstall.sh
	@echo '    echo -e "$${YELLOW}Disabling service...$${NC}"' >> scripts/uninstall.sh
	@echo '    systemctl disable sboxagent.service' >> scripts/uninstall.sh
	@echo 'fi' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Remove systemd service file' >> scripts/uninstall.sh
	@echo 'if [ -f $$SERVICE_DIR/sboxagent.service ]; then' >> scripts/uninstall.sh
	@echo '    echo -e "$${YELLOW}Removing systemd service...$${NC}"' >> scripts/uninstall.sh
	@echo '    rm -f $$SERVICE_DIR/sboxagent.service' >> scripts/uninstall.sh
	@echo '    systemctl daemon-reload' >> scripts/uninstall.sh
	@echo 'fi' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Remove binary' >> scripts/uninstall.sh
	@echo 'if [ -f $$INSTALL_DIR/$$BINARY_NAME ]; then' >> scripts/uninstall.sh
	@echo '    echo -e "$${YELLOW}Removing binary...$${NC}"' >> scripts/uninstall.sh
	@echo '    rm -f $$INSTALL_DIR/$$BINARY_NAME' >> scripts/uninstall.sh
	@echo 'fi' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Remove configuration (optional)' >> scripts/uninstall.sh
	@echo 'read -p "Remove configuration directory $$CONFIG_DIR? (y/N): " -n 1 -r' >> scripts/uninstall.sh
	@echo 'echo' >> scripts/uninstall.sh
	@echo 'if [[ $$REPLY =~ ^[Yy]$$ ]]; then' >> scripts/uninstall.sh
	@echo '    echo -e "$${YELLOW}Removing configuration...$${NC}"' >> scripts/uninstall.sh
	@echo '    rm -rf $$CONFIG_DIR' >> scripts/uninstall.sh
	@echo 'fi' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo '# Remove user and group (optional)' >> scripts/uninstall.sh
	@echo 'read -p "Remove user $$USER_NAME and group $$GROUP_NAME? (y/N): " -n 1 -r' >> scripts/uninstall.sh
	@echo 'echo' >> scripts/uninstall.sh
	@echo 'if [[ $$REPLY =~ ^[Yy]$$ ]]; then' >> scripts/uninstall.sh
	@echo '    echo -e "$${YELLOW}Removing user and group...$${NC}"' >> scripts/uninstall.sh
	@echo '    userdel $$USER_NAME 2>/dev/null || true' >> scripts/uninstall.sh
	@echo '    groupdel $$GROUP_NAME 2>/dev/null || true' >> scripts/uninstall.sh
	@echo 'fi' >> scripts/uninstall.sh
	@echo '' >> scripts/uninstall.sh
	@echo 'echo -e "$${GREEN}Uninstallation complete!$${NC}"' >> scripts/uninstall.sh
	@chmod +x scripts/uninstall.sh
	@echo "Uninstall script created: scripts/uninstall.sh"

# Install locally (for development)
.PHONY: install-local
install-local: build
	@echo "Installing locally..."
	@mkdir -p $(HOME)/.local/bin
	@cp $(BINARY_PATH) $(HOME)/.local/bin/
	@echo "Installed to $(HOME)/.local/bin/$(BINARY_NAME)"

# Uninstall locally
.PHONY: uninstall-local
uninstall-local:
	@echo "Uninstalling locally..."
	@rm -f $(HOME)/.local/bin/$(BINARY_NAME)
	@echo "Uninstalled from $(HOME)/.local/bin/$(BINARY_NAME)"

# Run the application
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	@$(BINARY_PATH)

# Run with specific config
.PHONY: run-config
run-config: build
	@echo "Running $(BINARY_NAME) with config..."
	@$(BINARY_PATH) -config examples/agent.yaml

# Show help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  build-linux        - Build for Linux"
	@echo "  clean              - Clean build artifacts"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage"
	@echo "  test-integration   - Run integration tests"
	@echo "  benchmark          - Run benchmarks"
	@echo "  deps               - Install dependencies"
	@echo "  fmt                - Format code"
	@echo "  lint               - Run linter"
	@echo "  check              - Run fmt, lint, and test"
	@echo "  docs               - Generate documentation"
	@echo "  docs-check         - Check documentation coverage"
	@echo "  docs-serve         - Start documentation server"
	@echo "  service            - Create systemd service file"
	@echo "  install-script     - Create install script"
	@echo "  uninstall-script   - Create uninstall script"
	@echo "  install-local      - Install locally"
	@echo "  uninstall-local    - Uninstall locally"
	@echo "  run                - Run the application"
	@echo "  run-config         - Run with config"
	@echo "  help               - Show this help"

# Development targets
.PHONY: dev-setup
dev-setup: deps fmt lint
	@echo "Development setup complete"

# Release preparation
.PHONY: release-prep
release-prep: clean test-coverage build-linux service install-script uninstall-script docs
	@echo "Release preparation complete"
	@echo "Files ready for release:"
	@echo "  - $(BIN_DIR)/$(BINARY_UNIX)"
	@echo "  - scripts/sboxagent.service"
	@echo "  - scripts/install.sh"
	@echo "  - scripts/uninstall.sh"
	@echo "  - $(DOCS_DIR)/README.md" 