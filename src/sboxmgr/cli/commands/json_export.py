"""
Enhanced CLI commands with JSON output support for sboxagent integration.

This module provides CLI commands that generate standardized JSON output
for integration with sboxagent according to ADR-0001 architecture.
"""

import json
import typer
import logging
from typing import Optional, List
from pathlib import Path
from datetime import datetime, timezone

from ...subscription.exporters.json_exporter import JSONExportFramework
from ...subscription.parser import parse_subscription
from ...server.exclusions import get_excluded_servers
from ...utils.version import get_version

logger = logging.getLogger(__name__)

json_app = typer.Typer(help="JSON export commands for sboxagent integration")

@json_app.command("generate")
def generate_json_config(
    url: str = typer.Option(..., "-u", "--url", help="Subscription URL"),
    client_type: str = typer.Option("sing-box", "-c", "--client", help="Target client type (sing-box, clash, xray, mihomo)"),
    output: Optional[Path] = typer.Option(None, "-o", "--output", help="Output file path (default: stdout)"),
    client_version: Optional[str] = typer.Option(None, "--version", help="Client version for compatibility"),
    exclude: Optional[str] = typer.Option(None, "--exclude", help="Comma-separated server names to exclude"),
    include: Optional[str] = typer.Option(None, "--include", help="Comma-separated server names to include"),
    pretty: bool = typer.Option(True, "--pretty/--compact", help="Pretty-print JSON output"),
    metadata: bool = typer.Option(True, "--metadata/--no-metadata", help="Include metadata in output"),
    debug: int = typer.Option(0, "-d", "--debug", help="Debug verbosity level (0-2)")
):
    """Generate standardized JSON configuration for sboxagent.
    
    Creates a JSON configuration with metadata that can be consumed by sboxagent
    for client management. The output follows the ADR-0001 interface protocol.
    
    Args:
        url: Subscription URL to fetch servers from
        client_type: Target client type (sing-box, clash, xray, mihomo)
        output: Output file path (default: stdout)
        client_version: Client version for compatibility
        exclude: Comma-separated server names to exclude
        include: Comma-separated server names to include
        pretty: Pretty-print JSON output
        metadata: Include metadata in output
        debug: Debug verbosity level
    """
    try:
        # Set debug level
        if debug > 0:
            logging.getLogger().setLevel(logging.DEBUG if debug > 1 else logging.INFO)
        
        # Parse subscription
        logger.info(f"Fetching subscription from: {url}")
        servers = parse_subscription(url)
        logger.info(f"Parsed {len(servers)} servers from subscription")
        
        # Apply exclusions
        if exclude:
            exclude_list = [s.strip() for s in exclude.split(",") if s.strip()]
            excluded_servers = get_excluded_servers()
            exclude_list.extend(excluded_servers)
            servers = [s for s in servers if s.name not in exclude_list]
            logger.info(f"Excluded {len(exclude_list)} servers, {len(servers)} remaining")
        
        # Apply inclusions
        if include:
            include_list = [s.strip() for s in include.split(",") if s.strip()]
            servers = [s for s in servers if s.name in include_list]
            logger.info(f"Included {len(include_list)} servers, {len(servers)} remaining")
        
        # Prepare options
        options = {}
        if exclude:
            options["exclude_servers"] = [s.strip() for s in exclude.split(",") if s.strip()]
        if include:
            options["include_servers"] = [s.strip() for s in include.split(",") if s.strip()]
        
        # Generate configuration
        framework = JSONExportFramework()
        config = framework.generate_config(
            servers=servers,
            client_type=client_type,
            subscription_url=url,
            client_version=client_version,
            options=options
        )
        
        # Remove metadata if not requested
        if not metadata:
            config.pop("metadata", None)
        
        # Format output
        indent = 2 if pretty else None
        json_str = json.dumps(config, indent=indent, ensure_ascii=False)
        
        # Output
        if output:
            output.write_text(json_str, encoding='utf-8')
            typer.echo(f"Configuration written to: {output}")
        else:
            typer.echo(json_str)
            
    except Exception as e:
        logger.error(f"Failed to generate configuration: {e}")
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(1)

@json_app.command("validate")
def validate_json_config(
    config_file: Path = typer.Option(..., "-f", "--file", help="Configuration file to validate"),
    client_type: str = typer.Option("sing-box", "-c", "--client", help="Client type for validation"),
    schema_dir: Optional[Path] = typer.Option(None, "--schema-dir", help="Custom schema directory"),
    debug: int = typer.Option(0, "-d", "--debug", help="Debug verbosity level (0-2)")
):
    """Validate JSON configuration against client schema.
    
    Validates a JSON configuration file against the appropriate client schema
    to ensure compatibility and correctness.
    
    Args:
        config_file: Configuration file to validate
        client_type: Client type for validation
        schema_dir: Custom schema directory
        debug: Debug verbosity level
    """
    try:
        # Set debug level
        if debug > 0:
            logging.getLogger().setLevel(logging.DEBUG if debug > 1 else logging.INFO)
        
        # Load configuration
        logger.info(f"Loading configuration from: {config_file}")
        config_data = json.loads(config_file.read_text(encoding='utf-8'))
        
        # Import validator (from sbox-common)
        try:
            import sys
            sys.path.append(str(Path(__file__).parent.parent.parent.parent / "sbox-common"))
            from validation.validator import ConfigValidator
            
            validator = ConfigValidator(schema_dir)
            result = validator.validate_config(config_data, client_type)
            
            # Output results
            if result["valid"]:
                typer.echo("✅ Configuration is valid")
                if result["warnings"]:
                    typer.echo("⚠️  Warnings:")
                    for warning in result["warnings"]:
                        typer.echo(f"   - {warning}")
            else:
                typer.echo("❌ Configuration is invalid")
                typer.echo("Errors:")
                for error in result["errors"]:
                    typer.echo(f"   - {error}")
                raise typer.Exit(1)
                
        except ImportError:
            typer.echo("⚠️  sbox-common validation not available, skipping schema validation")
            typer.echo("✅ Configuration file is valid JSON")
            
    except json.JSONDecodeError as e:
        typer.echo(f"❌ Invalid JSON: {e}", err=True)
        raise typer.Exit(1)
    except Exception as e:
        logger.error(f"Validation failed: {e}")
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(1)

@json_app.command("list-clients")
def list_supported_clients(
    output: str = typer.Option("table", "-o", "--output", help="Output format (table, json)"),
    debug: int = typer.Option(0, "-d", "--debug", help="Debug verbosity level (0-2)")
):
    """List supported client types and their capabilities.
    
    Displays information about supported client types and their
    capabilities for configuration generation.
    
    Args:
        output: Output format (table, json)
        debug: Debug verbosity level
    """
    try:
        # Set debug level
        if debug > 0:
            logging.getLogger().setLevel(logging.DEBUG if debug > 1 else logging.INFO)
        
        framework = JSONExportFramework()
        clients_info = {
            "sing-box": {
                "name": "Sing-box",
                "description": "Universal proxy platform",
                "supported_protocols": ["vmess", "vless", "trojan", "ss", "wireguard", "hysteria2", "tuic", "shadowtls"],
                "version_range": {"min": "1.0.0", "max": "1.8.0"}
            },
            "clash": {
                "name": "Clash",
                "description": "Rule-based proxy in Go",
                "supported_protocols": ["vmess", "ss", "ssr", "trojan", "snell"],
                "version_range": {"min": "1.0.0", "max": "1.18.0"}
            },
            "xray": {
                "name": "Xray",
                "description": "Platform for building proxies",
                "supported_protocols": ["vmess", "vless", "trojan", "shadowsocks"],
                "version_range": {"min": "1.0.0", "max": "1.8.0"}
            },
            "mihomo": {
                "name": "Mihomo",
                "description": "Clash fork with enhanced features",
                "supported_protocols": ["vmess", "ss", "ssr", "trojan", "snell", "hysteria", "tuic"],
                "version_range": {"min": "1.0.0", "max": "1.18.8"}
            }
        }
        
        if output == "json":
            result = {
                "clients": [
                    {
                        "type": client_type,
                        **info
                    }
                    for client_type, info in clients_info.items()
                ]
            }
            typer.echo(json.dumps(result, indent=2, ensure_ascii=False))
        else:
            # Table output
            typer.echo("Supported Client Types:")
            typer.echo("=" * 80)
            for client_type, info in clients_info.items():
                typer.echo(f"Type: {client_type}")
                typer.echo(f"Name: {info['name']}")
                typer.echo(f"Description: {info['description']}")
                typer.echo(f"Protocols: {', '.join(info['supported_protocols'])}")
                typer.echo(f"Version Range: {info['version_range']['min']} - {info['version_range']['max']}")
                typer.echo("-" * 40)
                
    except Exception as e:
        logger.error(f"Failed to list clients: {e}")
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(1)

@json_app.command("info")
def show_json_info(
    debug: int = typer.Option(0, "-d", "--debug", help="Debug verbosity level (0-2)")
):
    """Show JSON export framework information.
    
    Displays information about the JSON export framework including
    version, supported features, and integration capabilities.
    
    Args:
        debug: Debug verbosity level
    """
    try:
        # Set debug level
        if debug > 0:
            logging.getLogger().setLevel(logging.DEBUG if debug > 1 else logging.INFO)
        
        info = {
            "framework": "JSON Export Framework",
            "version": get_version(),
            "architecture": "ADR-0001 compliant",
            "supported_clients": ["sing-box", "clash", "xray", "mihomo"],
            "features": [
                "Standardized JSON output",
                "Metadata generation",
                "Checksum calculation",
                "Client-specific configuration",
                "Exclusion/inclusion filtering",
                "Schema validation support"
            ],
            "integration": {
                "sboxagent": "JSON interface protocol",
                "sbox-common": "Schema validation",
                "cli": "Enhanced commands with JSON output"
            },
            "timestamp": datetime.now(timezone.utc).isoformat()
        }
        
        typer.echo("JSON Export Framework Information:")
        typer.echo("=" * 50)
        typer.echo(f"Framework: {info['framework']}")
        typer.echo(f"Version: {info['version']}")
        typer.echo(f"Architecture: {info['architecture']}")
        typer.echo(f"Supported Clients: {', '.join(info['supported_clients'])}")
        typer.echo("\nFeatures:")
        for feature in info["features"]:
            typer.echo(f"  - {feature}")
        typer.echo("\nIntegration:")
        for component, description in info["integration"].items():
            typer.echo(f"  - {component}: {description}")
        typer.echo(f"\nTimestamp: {info['timestamp']}")
        
    except Exception as e:
        logger.error(f"Failed to show info: {e}")
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(1) 