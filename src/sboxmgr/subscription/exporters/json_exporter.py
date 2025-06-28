"""
JSON Export Framework for sboxmgr.

This module provides standardized JSON export functionality for all subbox clients
following the ADR-0001 architecture. It generates JSON configurations with metadata
for sboxagent integration.
"""

import json
import hashlib
import logging
from datetime import datetime, timezone
from typing import List, Dict, Any, Optional
from ..models import ParsedServer, ClientProfile
from ..base_exporter import BaseExporter
from ..registry import register

logger = logging.getLogger(__name__)

class JSONExportFramework:
    """Framework for generating standardized JSON configurations."""
    
    def __init__(self):
        """Initialize the JSON export framework."""
        self.supported_clients = ["sing-box", "clash", "xray", "mihomo"]
    
    def generate_config(
        self,
        servers: List[ParsedServer],
        client_type: str,
        subscription_url: str,
        client_version: Optional[str] = None,
        options: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Generate standardized JSON configuration.
        
        Args:
            servers: List of parsed servers
            client_type: Target client type (sing-box, clash, xray, mihomo)
            subscription_url: Source subscription URL
            client_version: Client version for compatibility
            options: Additional options (exclusions, routing, etc.)
            
        Returns:
            Standardized configuration with metadata
        """
        if client_type not in self.supported_clients:
            raise ValueError(f"Unsupported client type: {client_type}")
        
        # Generate client-specific configuration
        config_data = self._generate_client_config(servers, client_type, options)
        
        # Create standardized structure
        config = {
            "client": client_type,
            "version": client_version or "1.0.0",
            "created_at": datetime.now(timezone.utc).isoformat(),
            "config": config_data,
            "metadata": self._generate_metadata(servers, subscription_url, options)
        }
        
        # Add checksum
        config["metadata"]["checksum"] = self._calculate_checksum(config_data)
        
        return config
    
    def _generate_client_config(
        self,
        servers: List[ParsedServer],
        client_type: str,
        options: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Generate client-specific configuration data."""
        if client_type == "sing-box":
            return self._generate_singbox_config(servers, options)
        elif client_type == "clash":
            return self._generate_clash_config(servers, options)
        elif client_type == "xray":
            return self._generate_xray_config(servers, options)
        elif client_type == "mihomo":
            return self._generate_mihomo_config(servers, options)
        else:
            raise ValueError(f"Unsupported client type: {client_type}")
    
    def _generate_singbox_config(
        self,
        servers: List[ParsedServer],
        options: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Generate sing-box configuration."""
        from .singbox_exporter import singbox_export
        
        # Use existing singbox_exporter
        config = singbox_export(servers, [], None, None, True)
        
        # Apply options if provided
        if options:
            if "inbounds" in options:
                config["inbounds"] = options["inbounds"]
            if "dns" in options:
                config["dns"] = options["dns"]
            if "route" in options:
                config["route"] = options["route"]
        
        return config
    
    def _generate_clash_config(
        self,
        servers: List[ParsedServer],
        options: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Generate Clash configuration."""
        config = {
            "port": 7890,
            "socks-port": 7891,
            "mode": "rule",
            "log-level": "info",
            "external-controller": "127.0.0.1:9090",
            "proxies": [],
            "proxy-groups": [],
            "rules": []
        }
        
        # Convert servers to clash proxies
        for i, server in enumerate(servers):
            proxy = self._convert_to_clash_proxy(server, f"server-{i}")
            if proxy:
                config["proxies"].append(proxy)
        
        # Add default proxy group
        if config["proxies"]:
            config["proxy-groups"].append({
                "name": "Proxy",
                "type": "select",
                "proxies": [p["name"] for p in config["proxies"]]
            })
        
        # Add basic rules
        config["rules"] = [
            "DOMAIN-SUFFIX,google.com,Proxy",
            "GEOIP,CN,DIRECT",
            "MATCH,Proxy"
        ]
        
        return config
    
    def _generate_xray_config(
        self,
        servers: List[ParsedServer],
        options: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Generate Xray configuration."""
        config = {
            "log": {
                "loglevel": "warning"
            },
            "inbounds": [
                {
                    "port": 1080,
                    "protocol": "socks",
                    "settings": {
                        "auth": "noauth",
                        "udp": True
                    }
                }
            ],
            "outbounds": [],
            "routing": {
                "domainStrategy": "AsIs",
                "rules": []
            }
        }
        
        # Convert servers to xray outbounds
        for server in servers:
            outbound = self._convert_to_xray_outbound(server)
            if outbound:
                config["outbounds"].append(outbound)
        
        # Add direct and block outbounds
        config["outbounds"].extend([
            {
                "protocol": "freedom",
                "tag": "direct"
            },
            {
                "protocol": "blackhole",
                "tag": "block"
            }
        ])
        
        return config
    
    def _generate_mihomo_config(
        self,
        servers: List[ParsedServer],
        options: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Generate Mihomo configuration (similar to Clash)."""
        return self._generate_clash_config(servers, options)
    
    def _convert_to_clash_proxy(self, server: ParsedServer, name: str) -> Optional[Dict[str, Any]]:
        """Convert ParsedServer to Clash proxy format."""
        if server.protocol == "vmess":
            return {
                "name": name,
                "type": "vmess",
                "server": server.address,
                "port": server.port,
                "uuid": server.uuid,
                "alterId": getattr(server, "alter_id", 0),
                "cipher": "auto",
                "tls": getattr(server, "tls", False),
                "network": getattr(server, "network", "tcp")
            }
        elif server.protocol == "ss":
            return {
                "name": name,
                "type": "ss",
                "server": server.address,
                "port": server.port,
                "cipher": server.security,
                "password": server.password
            }
        elif server.protocol == "trojan":
            return {
                "name": name,
                "type": "trojan",
                "server": server.address,
                "port": server.port,
                "password": server.password,
                "tls": True
            }
        else:
            logger.warning(f"Unsupported protocol for clash: {server.protocol}")
            return None
    
    def _convert_to_xray_outbound(self, server: ParsedServer) -> Optional[Dict[str, Any]]:
        """Convert ParsedServer to Xray outbound format."""
        if server.protocol == "vmess":
            return {
                "protocol": "vmess",
                "settings": {
                    "vnext": [
                        {
                            "address": server.address,
                            "port": server.port,
                            "users": [
                                {
                                    "id": server.uuid,
                                    "alterId": getattr(server, "alter_id", 0)
                                }
                            ]
                        }
                    ]
                },
                "streamSettings": {
                    "network": getattr(server, "network", "tcp"),
                    "security": "tls" if getattr(server, "tls", False) else "none"
                }
            }
        elif server.protocol == "ss":
            return {
                "protocol": "shadowsocks",
                "settings": {
                    "servers": [
                        {
                            "address": server.address,
                            "port": server.port,
                            "method": server.security,
                            "password": server.password
                        }
                    ]
                }
            }
        elif server.protocol == "trojan":
            return {
                "protocol": "trojan",
                "settings": {
                    "servers": [
                        {
                            "address": server.address,
                            "port": server.port,
                            "password": server.password
                        }
                    ]
                },
                "streamSettings": {
                    "security": "tls"
                }
            }
        else:
            logger.warning(f"Unsupported protocol for xray: {server.protocol}")
            return None
    
    def _generate_metadata(
        self,
        servers: List[ParsedServer],
        subscription_url: str,
        options: Optional[Dict[str, Any]] = None
    ) -> Dict[str, Any]:
        """Generate metadata for configuration."""
        excluded_count = 0
        if options and "exclude_servers" in options:
            excluded_count = len(options["exclude_servers"])
        
        return {
            "source": subscription_url,
            "generator": "sboxmgr-1.5.0",
            "checksum": "",  # Will be calculated later
            "subscription_info": {
                "total_servers": len(servers) + excluded_count,
                "filtered_servers": len(servers),
                "excluded_servers": excluded_count
            }
        }
    
    def _calculate_checksum(self, config_data: Dict[str, Any]) -> str:
        """Calculate SHA-256 checksum of configuration data."""
        config_str = json.dumps(config_data, sort_keys=True, separators=(',', ':'))
        return hashlib.sha256(config_str.encode('utf-8')).hexdigest()

# Global instance
json_framework = JSONExportFramework()

@register("json")
class JSONExporter(BaseExporter):
    """Standardized JSON format configuration exporter.
    
    Implements the BaseExporter interface for generating standardized JSON
    configurations with metadata for sboxagent integration.
    """
    
    def __init__(self, client_type: str = "sing-box"):
        """Initialize JSON exporter.
        
        Args:
            client_type: Target client type (sing-box, clash, xray, mihomo)
        """
        self.client_type = client_type
        self.framework = json_framework
    
    def export(self, servers: List[ParsedServer]) -> str:
        """Export servers to standardized JSON configuration string.
        
        Args:
            servers: List of ParsedServer objects to export.
            
        Returns:
            JSON string containing standardized configuration with metadata.
            
        Raises:
            ValueError: If server data is invalid or cannot be exported.
        """
        config = self.framework.generate_config(
            servers=servers,
            client_type=self.client_type,
            subscription_url="unknown",  # Should be passed from context
            options=None
        )
        return json.dumps(config, indent=2, ensure_ascii=False) 