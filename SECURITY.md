# 🔐 Subbox Security Plan

> Version: 1.0 — Last Updated: 2025-06-27  
> Maintainer: Mikhail Stepanov (kpblcaoo@gmail.com)  
> License: GPL-3.0

## Table of Contents

- [Overview](#overview)
- [Security Model](#security-model)
- [Deployment Security](#deployment-security)
- [Permission Verification](#permission-verification)
- [Packaging Security](#packaging-security)
- [Security Restrictions](#security-restrictions)
- [Security Monitoring](#security-monitoring)
- [Deployment Checklist](#deployment-checklist)
- [Security Updates](#security-updates)
- [Vulnerability Reporting](#vulnerability-reporting)

## Overview

This document defines the secure deployment model and permission structure for the `sboxagent` system component. SBoxAgent is a Go daemon that manages sing-box proxy configurations through [sboxctl](https://github.com/kpblcaoo/sboxctl) CLI integration.

### Security Principles

- **Principle of Least Privilege**: Minimal required permissions only
- **Defense in Depth**: Multiple layers of security controls
- **Secure by Default**: Secure configuration out of the box
- **Audit Trail**: Comprehensive logging of all security-relevant actions

## Security Model

### User and Permission Matrix

| Component | User | Trust Level | Permissions | Rationale |
|-----------|------|-------------|-------------|-----------|
| sboxagent | sboxagent (system user) | 🔒 Restricted | Minimal sudo rights | Isolated execution context |
| sboxctl | sboxagent → subprocess | 🟢 Safe | Read-only access | CLI tool execution |
| sing-box | systemd unit | 🔐 System | Service management | Proxy service control |
| Configs | /etc/sing-box/ | ACL/Group | Write access only | Configuration updates |

### Required Permissions

| Action | Required Permission | Command/Path | Security Control |
|--------|-------------------|--------------|------------------|
| Write config.json | File write access | `sudo cp /tmp/config.json /etc/sing-box/` | ACL or group membership |
| Execute sboxctl | User execution | `sboxctl update` | Standard user permissions |
| Restart sing-box | Service control | `sudo systemctl restart sing-box.service` | Specific sudo command |
| Write logs | Directory access | `/var/log/sboxagent/` | Owned by sboxagent |
| Health checks | Network access | `curl`, sockets | Standard user permissions |

## Deployment Security

### 1. System User Creation

```bash
# Create dedicated system user
sudo useradd --system --no-create-home --shell /usr/sbin/nologin sboxagent
```

**Security Benefits:**
- No interactive shell access
- System user privileges only
- Isolated execution context

### 2. Directory Structure and Permissions

```bash
# Create secure directory structure
sudo mkdir -p /etc/sboxagent /var/log/sboxagent /var/lib/sboxagent
sudo chown -R sboxagent:sboxagent /etc/sboxagent /var/log/sboxagent /var/lib/sboxagent
sudo chmod 750 /etc/sboxagent /var/log/sboxagent /var/lib/sboxagent
```

**Security Controls:**
- Owner-only access to sensitive directories
- No world-readable permissions
- Proper separation of concerns

### 3. Sudo Configuration

```bash
# /etc/sudoers.d/sboxagent
sboxagent ALL=(ALL) NOPASSWD: /bin/systemctl restart sing-box.service
sboxagent ALL=(ALL) NOPASSWD: /bin/cp /tmp/config.json /etc/sing-box/config.json
```

**Security Features:**
- Specific command allowlisting
- No password required for automation
- Minimal privilege escalation

### 4. Configuration File Access Control

#### Option A: ACL (Recommended)
```bash
sudo setfacl -m u:sboxagent:rw /etc/sing-box/config.json
```

#### Option B: Group-based Access (Fallback)
```bash
# Create dedicated group
sudo groupadd singbox

# Add users to group
sudo usermod -a -G singbox sboxagent
sudo usermod -a -G singbox root

# Set secure permissions
sudo chown root:singbox /etc/sing-box/config.json
sudo chmod 640 /etc/sing-box/config.json
```

## Permission Verification

### Manual Testing
```bash
# Test configuration generation
sudo -u sboxagent /usr/bin/sboxctl generate --output /tmp/test.json

# Test configuration deployment
sudo -u sboxagent cp /tmp/test.json /etc/sing-box/config.json

# Test service restart
sudo -u sboxagent systemctl restart sing-box.service

# Cleanup
sudo rm /tmp/test.json
```

### Automated Verification Script
```bash
#!/bin/bash
# scripts/verify-permissions.sh

set -euo pipefail

echo "🔍 Subbox Security Verification"

# User existence check
if ! id sboxagent >/dev/null 2>&1; then
    echo "❌ User sboxagent not found"
    exit 1
fi
echo "✅ User sboxagent exists"

# Directory permission verification
declare -a dirs=("/etc/sboxagent" "/var/log/sboxagent" "/var/lib/sboxagent")
for dir in "${dirs[@]}"; do
    if [[ -d "$dir" ]] && [[ "$(stat -c '%U:%G' "$dir")" == "sboxagent:sboxagent" ]]; then
        echo "✅ Directory $dir properly configured"
    else
        echo "❌ Directory $dir misconfigured"
        exit 1
    fi
done

# Sudo permissions verification
if sudo -lU sboxagent | grep -E "systemctl restart sing-box|cp .*config\.json" >/dev/null; then
    echo "✅ Sudo permissions configured"
else
    echo "❌ Sudo permissions missing"
    exit 1
fi

# Configuration file access check
if [[ -f "/etc/sing-box/config.json" ]]; then
    if getfacl "/etc/sing-box/config.json" 2>/dev/null | grep -q "user:sboxagent:rw"; then
        echo "✅ ACL permissions configured"
    elif [[ "$(stat -c '%G' "/etc/sing-box/config.json")" == "singbox" ]]; then
        echo "✅ Group permissions configured"
    else
        echo "❌ Configuration file permissions misconfigured"
        exit 1
    fi
fi

echo "🔍 Security verification completed successfully"
```

## Packaging Security

### Installation Script Requirements

The installation script must implement the following security measures:

1. **User and Group Creation**
   ```bash
   # Create system user with secure defaults
   sudo useradd --system --no-create-home --shell /usr/sbin/nologin sboxagent
   sudo groupadd singbox 2>/dev/null || true
   sudo usermod -a -G singbox sboxagent
   ```

2. **Sudo Configuration**
   ```bash
   # Create secure sudoers configuration
   sudo tee /etc/sudoers.d/sboxagent > /dev/null <<EOF
   sboxagent ALL=(ALL) NOPASSWD: /bin/systemctl restart sing-box.service
   sboxagent ALL=(ALL) NOPASSWD: /bin/cp /tmp/config.json /etc/sing-box/config.json
   EOF
   sudo chmod 440 /etc/sudoers.d/sboxagent
   ```

3. **Directory and Permission Setup**
   ```bash
   # Create secure directory structure
   sudo mkdir -p /etc/sboxagent /var/log/sboxagent /var/lib/sboxagent
   sudo chown -R sboxagent:sboxagent /etc/sboxagent /var/log/sboxagent /var/lib/sboxagent
   sudo chmod 750 /etc/sboxagent /var/log/sboxagent /var/lib/sboxagent
   
   # Configure configuration file access
   if command -v setfacl >/dev/null 2>&1; then
       sudo setfacl -m u:sboxagent:rw /etc/sing-box/config.json
   else
       sudo chown root:singbox /etc/sing-box/config.json
       sudo chmod 640 /etc/sing-box/config.json
   fi
   ```

4. **Installation Logging**
   ```bash
   # Log all installation actions
   sudo mkdir -p /var/log/sboxagent
   sudo tee /var/log/sboxagent/install.log > /dev/null <<EOF
   $(date -u): SBoxAgent installation
   - User: sboxagent (system)
   - Groups: sboxagent, singbox
   - Directories: /etc/sboxagent, /var/log/sboxagent, /var/lib/sboxagent
   - Sudo permissions: configured
   - Configuration access: configured
   - Installer: $(whoami)@$(hostname)
   EOF
   ```

5. **Post-Installation Verification**
   ```bash
   # Verify installation security
   ./scripts/verify-permissions.sh
   ```

## Security Restrictions

### Prohibited Actions

- ❌ **No root execution**: Never run sboxagent as root
- ❌ **No broad sudo access**: Avoid wildcard sudo permissions
- ❌ **No world-writable files**: Prevent unauthorized modifications
- ❌ **No network exposure**: Don't expose internal APIs publicly
- ❌ **No credential storage**: Don't store secrets in plain text

### Recommended Security Practices

- ✅ **Principle of least privilege**: Grant only necessary permissions
- ✅ **Secure defaults**: Secure configuration out of the box
- ✅ **Comprehensive logging**: Log all security-relevant actions
- ✅ **Regular audits**: Periodic security assessments
- ✅ **Configuration validation**: Validate all configuration changes

## Security Monitoring

### Logging Strategy

| Log Type | Location | Retention | Purpose |
|----------|----------|-----------|---------|
| Application logs | `/var/log/sboxagent/` | 30 days | Operational monitoring |
| Sudo commands | `/var/log/auth.log` | 90 days | Privilege escalation audit |
| Systemd logs | `journalctl -u sboxagent` | 30 days | Service monitoring |
| Installation logs | `/var/log/sboxagent/install.log` | Permanent | Deployment audit |

### Security Auditing

- **Regular permission checks**: Monthly verification of sudo permissions
- **Configuration integrity**: Monitor for unauthorized config changes
- **Binary integrity**: Verify sboxagent binary checksums
- **Access pattern analysis**: Monitor for unusual access patterns

## Deployment Checklist

### Pre-deployment
- [ ] Security requirements reviewed
- [ ] Target environment assessed
- [ ] Installation script tested
- [ ] Rollback plan prepared

### Installation
- [ ] System user created (sboxagent)
- [ ] Directories created with proper permissions
- [ ] Sudo configuration applied
- [ ] Configuration file access configured
- [ ] Installation logged

### Post-installation
- [ ] Security verification completed
- [ ] Service functionality tested
- [ ] Logging verified
- [ ] Documentation updated

## Security Updates

### Version Management

- **Security changes require versioning**: All security modifications must be versioned
- **Change documentation**: Document all security changes in security-changelog.md
- **Staging testing**: Test security changes in staging environment
- **Rollback procedures**: Maintain rollback procedures for security changes

### Incident Response

- **Quick rollback**: Procedures for rapid security rollback
- **Configuration backup**: Backup security configuration
- **Recovery plan**: Plan for security incident recovery

## Vulnerability Reporting

### Responsible Disclosure

If you discover a security vulnerability in SBoxAgent, please follow responsible disclosure practices:

1. **DO NOT** create a public issue
2. **DO** email: kpblcaoo@gmail.com
3. **Include** in your report:
   - Vulnerability description
   - Reproduction steps
   - Potential impact assessment
   - Suggested remediation

### Response Timeline

- **Critical vulnerabilities**: 24-48 hours
- **High severity**: 3-5 business days
- **Medium severity**: 1-2 weeks
- **Low severity**: 2-4 weeks

### Security Acknowledgments

We acknowledge and thank all security researchers who help make SBoxAgent more secure through responsible disclosure.

---

**Note:** This document is a living document and will be updated as the project evolves. The latest version is always available in the repository. 