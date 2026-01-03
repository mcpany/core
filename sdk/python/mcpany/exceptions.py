# Copyright 2025 Author(s) of MCP Any
# SPDX-License-Identifier: Apache-2.0

"""
This module defines the exceptions used by the MCP Any Python Client SDK.
"""

class MCPAnyError(Exception):
    """Base exception for MCPAny SDK"""

class ConnectionError(MCPAnyError):
    """Raised when connection to server fails"""

class ToolExecutionError(MCPAnyError):
    """Raised when a tool execution fails"""
