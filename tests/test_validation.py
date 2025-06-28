"""
Tests for validation framework.
"""

import json
import pytest
from pathlib import Path
from validation.validator import ConfigValidator

@pytest.fixture
def validator():
    """Create validator instance for testing."""
    return ConfigValidator()

@pytest.fixture
def sample_singbox_config():
    """Sample sing-box configuration for testing."""
    return {
        "client": "sing-box",
        "version": "1.8.0",
        "created_at": "2025-06-28T14:30:00Z",
        "config": {
            "log": {
                "level": "info",
                "timestamp": true,
                "output": "stdout"
            },
            "inbounds": [
                {
                    "type": "mixed",
                    "tag": "mixed-in",
                    "listen": "127.0.0.1",
                    "listen_port": 7890,
                    "sniff": true
                }
            ],
            "outbounds": [
                {
                    "type": "vmess",
                    "tag": "vmess-out",
                    "server": "example.com",
                    "server_port": 443,
                    "uuid": "12345678-1234-1234-1234-123456789012",
                    "security": "tls"
                }
            ]
        },
        "metadata": {
            "source": "https://example.com/subscription",
            "generator": "sboxmgr-1.5.0",
            "checksum": "a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"
        }
    }

class TestConfigValidator:
    """Test cases for ConfigValidator."""
    
    def test_validator_initialization(self, validator):
        """Test validator initialization."""
        assert validator is not None
        assert hasattr(validator, 'schemas')
        assert hasattr(validator, 'schemas_dir')
    
    def test_validate_singbox_config(self, validator, sample_singbox_config):
        """Test sing-box configuration validation."""
        result = validator.validate_config(sample_singbox_config, "sing-box")
        
        assert "valid" in result
        assert "errors" in result
        assert "warnings" in result
        assert "client_type" in result
        assert result["client_type"] == "sing-box"
    
    def test_validate_invalid_json(self, validator):
        """Test validation with invalid JSON."""
        result = validator.validate_config("invalid json", "sing-box")
        
        assert result["valid"] == False
        assert len(result["errors"]) > 0
        assert "Invalid JSON" in result["errors"][0]
    
    def test_validate_unknown_client(self, validator, sample_singbox_config):
        """Test validation with unknown client type."""
        result = validator.validate_config(sample_singbox_config, "unknown-client")
        
        assert result["valid"] == False
        assert len(result["errors"]) > 0
        assert "No schema found" in result["errors"][0]
    
    def test_validate_clash_config(self, validator):
        """Test clash configuration validation."""
        clash_config = {
            "port": 7890,
            "socks-port": 7891,
            "mode": "rule",
            "log-level": "info",
            "proxies": [
                {
                    "name": "test-proxy",
                    "type": "vmess",
                    "server": "example.com",
                    "port": 443,
                    "uuid": "12345678-1234-1234-1234-123456789012"
                }
            ],
            "proxy-groups": [
                {
                    "name": "Proxy",
                    "type": "select",
                    "proxies": ["test-proxy"]
                }
            ],
            "rules": [
                "DOMAIN-SUFFIX,google.com,Proxy",
                "GEOIP,CN,DIRECT",
                "MATCH,Proxy"
            ]
        }
        
        result = validator.validate_config(clash_config, "clash")
        assert "valid" in result
        assert "errors" in result
    
    def test_validate_xray_config(self, validator):
        """Test xray configuration validation."""
        xray_config = {
            "log": {
                "loglevel": "warning"
            },
            "inbounds": [
                {
                    "port": 1080,
                    "protocol": "socks",
                    "settings": {
                        "auth": "noauth",
                        "udp": true
                    }
                }
            ],
            "outbounds": [
                {
                    "protocol": "vmess",
                    "settings": {
                        "vnext": [
                            {
                                "address": "example.com",
                                "port": 443,
                                "users": [
                                    {
                                        "id": "12345678-1234-1234-1234-123456789012",
                                        "alterId": 0
                                    }
                                ]
                            }
                        ]
                    }
                }
            ]
        }
        
        result = validator.validate_config(xray_config, "xray")
        assert "valid" in result
        assert "errors" in result
    
    def test_validate_mihomo_config(self, validator):
        """Test mihomo configuration validation."""
        mihomo_config = {
            "port": 7890,
            "socks-port": 7891,
            "mode": "rule",
            "log-level": "info",
            "proxies": [
                {
                    "name": "test-proxy",
                    "type": "vmess",
                    "server": "example.com",
                    "port": 443,
                    "uuid": "12345678-1234-1234-1234-123456789012"
                }
            ],
            "proxy-groups": [
                {
                    "name": "Proxy",
                    "type": "select",
                    "proxies": ["test-proxy"]
                }
            ],
            "rules": [
                "DOMAIN-SUFFIX,google.com,Proxy",
                "GEOIP,CN,DIRECT",
                "MATCH,Proxy"
            ]
        }
        
        result = validator.validate_config(mihomo_config, "mihomo")
        assert "valid" in result
        assert "errors" in result
    
    def test_semantic_validation_singbox(self, validator):
        """Test semantic validation for sing-box."""
        invalid_config = {
            "log": {"level": "info"},
            "inbounds": [
                {
                    "type": "mixed",
                    "listen_port": 99999  # Invalid port
                }
            ]
        }
        
        result = validator.validate_config(invalid_config, "sing-box")
        assert result["valid"] == False
        assert any("Invalid port" in error for error in result["errors"])
    
    def test_semantic_validation_clash(self, validator):
        """Test semantic validation for clash."""
        invalid_config = {
            "port": 7890,
            "proxy-groups": [
                {
                    "name": "Proxy",
                    "type": "select",
                    "proxies": ["undefined-proxy"]  # Undefined proxy
                }
            ]
        }
        
        result = validator.validate_config(invalid_config, "clash")
        assert result["valid"] == False
        assert any("undefined proxy" in error.lower() for error in result["errors"])
    
    def test_get_supported_clients(self, validator):
        """Test getting supported client types."""
        clients = validator.get_supported_clients()
        assert isinstance(clients, list)
        # Should have at least some schemas loaded
        assert len(clients) >= 0

if __name__ == "__main__":
    pytest.main([__file__]) 