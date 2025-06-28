"""
Configuration validator for subbox clients.

This module provides validation utilities for sing-box, clash, xray, and mihomo configurations.
"""

import json
import logging
from pathlib import Path
from typing import Dict, List, Optional, Union, Any
from jsonschema import validate, ValidationError, SchemaError
from jsonschema.validators import Draft202012Validator

logger = logging.getLogger(__name__)

class ConfigValidator:
    """Validator for subbox client configurations."""
    
    def __init__(self, schemas_dir: Optional[Path] = None):
        """Initialize validator with schema directory.
        
        Args:
            schemas_dir: Directory containing JSON schemas. Defaults to ./schemas
        """
        self.schemas_dir = schemas_dir or Path(__file__).parent.parent / "schemas"
        self.schemas: Dict[str, Dict[str, Any]] = {}
        self._load_schemas()
    
    def _load_schemas(self) -> None:
        """Load all JSON schemas from schemas directory."""
        if not self.schemas_dir.exists():
            logger.warning(f"Schemas directory not found: {self.schemas_dir}")
            return
            
        for schema_file in self.schemas_dir.glob("*.schema.json"):
            try:
                with open(schema_file, 'r', encoding='utf-8') as f:
                    schema = json.load(f)
                    schema_id = schema.get('$id', schema_file.stem)
                    self.schemas[schema_id] = schema
                    logger.debug(f"Loaded schema: {schema_id}")
            except Exception as e:
                logger.error(f"Failed to load schema {schema_file}: {e}")
    
    def validate_config(self, config_data: Union[Dict, str], client_type: str) -> Dict[str, Any]:
        """Validate configuration data against client schema.
        
        Args:
            config_data: Configuration data (dict or JSON string)
            client_type: Client type (sing-box, clash, xray, mihomo)
            
        Returns:
            Validation result with success status and errors/warnings
        """
        result = {
            "valid": False,
            "errors": [],
            "warnings": [],
            "client_type": client_type
        }
        
        # Parse config data if it's a string
        if isinstance(config_data, str):
            try:
                config_data = json.loads(config_data)
            except json.JSONDecodeError as e:
                result["errors"].append(f"Invalid JSON: {e}")
                return result
        
        # Find appropriate schema
        schema = self._get_schema_for_client(client_type)
        if not schema:
            result["errors"].append(f"No schema found for client type: {client_type}")
            return result
        
        # Validate against schema
        try:
            validate(instance=config_data, schema=schema, cls=Draft202012Validator)
            result["valid"] = True
            logger.debug(f"Configuration validated successfully for {client_type}")
        except ValidationError as e:
            result["errors"].append(f"Validation error: {e.message}")
            logger.warning(f"Configuration validation failed for {client_type}: {e.message}")
        except SchemaError as e:
            result["errors"].append(f"Schema error: {e.message}")
            logger.error(f"Schema error for {client_type}: {e.message}")
        except Exception as e:
            result["errors"].append(f"Unexpected error: {e}")
            logger.error(f"Unexpected validation error for {client_type}: {e}")
        
        # Additional semantic validation
        semantic_errors = self._validate_semantics(config_data, client_type)
        result["errors"].extend(semantic_errors)
        
        return result
    
    def _get_schema_for_client(self, client_type: str) -> Optional[Dict[str, Any]]:
        """Get schema for specific client type.
        
        Args:
            client_type: Client type (sing-box, clash, xray, mihomo)
            
        Returns:
            Schema dictionary or None if not found
        """
        schema_mapping = {
            "sing-box": "https://schemas.subbox.dev/sing-box.schema.json",
            "clash": "https://schemas.subbox.dev/clash.schema.json", 
            "xray": "https://schemas.subbox.dev/xray.schema.json",
            "mihomo": "https://schemas.subbox.dev/mihomo.schema.json"
        }
        
        schema_id = schema_mapping.get(client_type)
        if schema_id and schema_id in self.schemas:
            return self.schemas[schema_id]
        
        # Fallback to file-based lookup
        schema_file = self.schemas_dir / f"{client_type.replace('-', '_')}.schema.json"
        if schema_file.exists():
            try:
                with open(schema_file, 'r', encoding='utf-8') as f:
                    return json.load(f)
            except Exception as e:
                logger.error(f"Failed to load schema from {schema_file}: {e}")
        
        return None
    
    def _validate_semantics(self, config_data: Dict[str, Any], client_type: str) -> List[str]:
        """Perform semantic validation beyond JSON schema.
        
        Args:
            config_data: Configuration data
            client_type: Client type
            
        Returns:
            List of semantic validation errors
        """
        errors = []
        
        if client_type == "sing-box":
            errors.extend(self._validate_singbox_semantics(config_data))
        elif client_type == "clash":
            errors.extend(self._validate_clash_semantics(config_data))
        elif client_type == "xray":
            errors.extend(self._validate_xray_semantics(config_data))
        elif client_type == "mihomo":
            errors.extend(self._validate_mihomo_semantics(config_data))
        
        return errors
    
    def _validate_singbox_semantics(self, config: Dict[str, Any]) -> List[str]:
        """Validate sing-box specific semantics."""
        errors = []
        
        # Check for required outbounds
        if "outbounds" not in config or not config["outbounds"]:
            errors.append("sing-box configuration must have at least one outbound")
        
        # Check for valid port ranges
        if "inbounds" in config:
            for i, inbound in enumerate(config["inbounds"]):
                if "listen_port" in inbound:
                    port = inbound["listen_port"]
                    if not (1 <= port <= 65535):
                        errors.append(f"Invalid port {port} in inbound {i}")
        
        return errors
    
    def _validate_clash_semantics(self, config: Dict[str, Any]) -> List[str]:
        """Validate clash specific semantics."""
        errors = []
        
        # Check for required proxies
        if "proxies" not in config or not config["proxies"]:
            errors.append("clash configuration must have at least one proxy")
        
        # Check proxy group references
        if "proxy-groups" in config:
            proxy_names = {p["name"] for p in config.get("proxies", [])}
            for group in config["proxy-groups"]:
                for proxy in group.get("proxies", []):
                    if proxy not in proxy_names:
                        errors.append(f"Proxy group '{group['name']}' references undefined proxy '{proxy}'")
        
        return errors
    
    def _validate_xray_semantics(self, config: Dict[str, Any]) -> List[str]:
        """Validate xray specific semantics."""
        errors = []
        
        # Check for required outbounds
        if "outbounds" not in config or not config["outbounds"]:
            errors.append("xray configuration must have at least one outbound")
        
        # Check for valid port ranges
        if "inbounds" in config:
            for i, inbound in enumerate(config["inbounds"]):
                if "port" in inbound:
                    port = inbound["port"]
                    if not (1 <= port <= 65535):
                        errors.append(f"Invalid port {port} in inbound {i}")
        
        return errors
    
    def _validate_mihomo_semantics(self, config: Dict[str, Any]) -> List[str]:
        """Validate mihomo specific semantics."""
        errors = []
        
        # Check for required proxies
        if "proxies" not in config or not config["proxies"]:
            errors.append("mihomo configuration must have at least one proxy")
        
        # Check proxy group references
        if "proxy-groups" in config:
            proxy_names = {p["name"] for p in config.get("proxies", [])}
            for group in config["proxy-groups"]:
                for proxy in group.get("proxies", []):
                    if proxy not in proxy_names:
                        errors.append(f"Proxy group '{group['name']}' references undefined proxy '{proxy}'")
        
        return errors
    
    def get_supported_clients(self) -> List[str]:
        """Get list of supported client types.
        
        Returns:
            List of supported client types
        """
        return list(self.schemas.keys()) 