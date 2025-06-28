"""
Validation framework for sbox-common.

This package provides JSON schema validation utilities for all subbox client configurations.
"""

__version__ = "1.0.0"
__author__ = "Subbox Team"

from .validator import ConfigValidator
from .schemas import SchemaRegistry

__all__ = ["ConfigValidator", "SchemaRegistry"] 