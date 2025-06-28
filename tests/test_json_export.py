"""
Tests for JSON Export Framework.
"""

import json
import pytest
from pathlib import Path
from unittest.mock import Mock, patch
from src.sboxmgr.subscription.exporters.json_exporter import JSONExportFramework, JSONExporter
from src.sboxmgr.subscription.models import ParsedServer

@pytest.fixture
def sample_servers():
    """Sample servers for testing."""
    return [
        ParsedServer(
            name="test-vmess",
            protocol="vmess",
            address="example.com",
            port=443,
            uuid="12345678-1234-1234-1234-123456789012",
            security="auto",
            network="ws",
            tls=True
        ),
        ParsedServer(
            name="test-ss",
            protocol="ss",
            address="example2.com",
            port=8388,
            security="aes-256-gcm",
            password="password123"
        )
    ]

@pytest.fixture
def framework():
    """Create JSONExportFramework instance for testing."""
    return JSONExportFramework()

class TestJSONExportFramework:
    """Test cases for JSONExportFramework."""
    
    def test_initialization(self, framework):
        """Test framework initialization."""
        assert framework is not None
        assert hasattr(framework, 'supported_clients')
        assert len(framework.supported_clients) == 4
        assert "sing-box" in framework.supported_clients
        assert "clash" in framework.supported_clients
        assert "xray" in framework.supported_clients
        assert "mihomo" in framework.supported_clients
    
    def test_generate_config_singbox(self, framework, sample_servers):
        """Test sing-box configuration generation."""
        config = framework.generate_config(
            servers=sample_servers,
            client_type="sing-box",
            subscription_url="https://example.com/subscription"
        )
        
        assert "client" in config
        assert config["client"] == "sing-box"
        assert "version" in config
        assert "created_at" in config
        assert "config" in config
        assert "metadata" in config
        assert "checksum" in config["metadata"]
    
    def test_generate_config_clash(self, framework, sample_servers):
        """Test clash configuration generation."""
        config = framework.generate_config(
            servers=sample_servers,
            client_type="clash",
            subscription_url="https://example.com/subscription"
        )
        
        assert config["client"] == "clash"
        assert "config" in config
        clash_config = config["config"]
        assert "port" in clash_config
        assert "proxies" in clash_config
        assert "proxy-groups" in clash_config
        assert "rules" in clash_config
    
    def test_generate_config_xray(self, framework, sample_servers):
        """Test xray configuration generation."""
        config = framework.generate_config(
            servers=sample_servers,
            client_type="xray",
            subscription_url="https://example.com/subscription"
        )
        
        assert config["client"] == "xray"
        assert "config" in config
        xray_config = config["config"]
        assert "log" in xray_config
        assert "inbounds" in xray_config
        assert "outbounds" in xray_config
        assert "routing" in xray_config
    
    def test_generate_config_mihomo(self, framework, sample_servers):
        """Test mihomo configuration generation."""
        config = framework.generate_config(
            servers=sample_servers,
            client_type="mihomo",
            subscription_url="https://example.com/subscription"
        )
        
        assert config["client"] == "mihomo"
        assert "config" in config
        mihomo_config = config["config"]
        assert "port" in mihomo_config
        assert "proxies" in mihomo_config
    
    def test_unsupported_client_type(self, framework, sample_servers):
        """Test error handling for unsupported client type."""
        with pytest.raises(ValueError, match="Unsupported client type"):
            framework.generate_config(
                servers=sample_servers,
                client_type="unsupported",
                subscription_url="https://example.com/subscription"
            )
    
    def test_metadata_generation(self, framework, sample_servers):
        """Test metadata generation."""
        config = framework.generate_config(
            servers=sample_servers,
            client_type="sing-box",
            subscription_url="https://example.com/subscription"
        )
        
        metadata = config["metadata"]
        assert "source" in metadata
        assert metadata["source"] == "https://example.com/subscription"
        assert "generator" in metadata
        assert "checksum" in metadata
        assert "subscription_info" in metadata
        
        subscription_info = metadata["subscription_info"]
        assert "total_servers" in subscription_info
        assert "filtered_servers" in subscription_info
        assert "excluded_servers" in subscription_info
        assert subscription_info["filtered_servers"] == 2
    
    def test_options_handling(self, framework, sample_servers):
        """Test options handling in configuration generation."""
        options = {
            "exclude_servers": ["server1", "server2"],
            "include_servers": ["test-vmess"]
        }
        
        config = framework.generate_config(
            servers=sample_servers,
            client_type="sing-box",
            subscription_url="https://example.com/subscription",
            options=options
        )
        
        metadata = config["metadata"]
        subscription_info = metadata["subscription_info"]
        assert subscription_info["excluded_servers"] == 2
    
    def test_checksum_calculation(self, framework, sample_servers):
        """Test checksum calculation."""
        config = framework.generate_config(
            servers=sample_servers,
            client_type="sing-box",
            subscription_url="https://example.com/subscription"
        )
        
        checksum = config["metadata"]["checksum"]
        assert len(checksum) == 64  # SHA-256 hex length
        assert all(c in "0123456789abcdef" for c in checksum)
    
    def test_convert_to_clash_proxy(self, framework, sample_servers):
        """Test conversion to clash proxy format."""
        vmess_server = sample_servers[0]
        proxy = framework._convert_to_clash_proxy(vmess_server, "test-proxy")
        
        assert proxy is not None
        assert proxy["name"] == "test-proxy"
        assert proxy["type"] == "vmess"
        assert proxy["server"] == "example.com"
        assert proxy["port"] == 443
        assert proxy["uuid"] == "12345678-1234-1234-1234-123456789012"
    
    def test_convert_to_xray_outbound(self, framework, sample_servers):
        """Test conversion to xray outbound format."""
        vmess_server = sample_servers[0]
        outbound = framework._convert_to_xray_outbound(vmess_server)
        
        assert outbound is not None
        assert outbound["protocol"] == "vmess"
        assert "settings" in outbound
        assert "streamSettings" in outbound
        
        settings = outbound["settings"]
        assert "vnext" in settings
        assert len(settings["vnext"]) == 1
        assert settings["vnext"][0]["address"] == "example.com"

class TestJSONExporter:
    """Test cases for JSONExporter."""
    
    def test_initialization(self):
        """Test JSONExporter initialization."""
        exporter = JSONExporter("sing-box")
        assert exporter.client_type == "sing-box"
        assert exporter.framework is not None
    
    def test_export(self, sample_servers):
        """Test JSON export functionality."""
        exporter = JSONExporter("sing-box")
        result = exporter.export(sample_servers)
        
        # Should be valid JSON
        config = json.loads(result)
        assert "client" in config
        assert config["client"] == "sing-box"
        assert "config" in config
    
    def test_export_with_different_client(self, sample_servers):
        """Test export with different client types."""
        for client_type in ["clash", "xray", "mihomo"]:
            exporter = JSONExporter(client_type)
            result = exporter.export(sample_servers)
            config = json.loads(result)
            assert config["client"] == client_type

if __name__ == "__main__":
    pytest.main([__file__]) 