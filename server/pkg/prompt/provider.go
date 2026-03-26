// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package prompt

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type mcpServerProvider struct {
	server *mcp.Server
}

// Server returns the underlying MCP server instance.
// Summary: Retrieves the MCP server.
// Returns:
//   - *mcp.Server: The MCP server instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Server returns the underlying MCP server instance.
// Summary: Retrieves the MCP server.
// Returns:
//   - *mcp.Server: The MCP server instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	return p.server
}

// NewMCPServerProvider creates a new MCPServerProvider.
// Summary: Initializes a provider for the MCP server.
// Parameters:
//   - server: *mcp.Server. The server instance to wrap.
//
// Returns:
//   - MCPServerProvider: The initialized provider.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// NewMCPServerProvider creates a new MCPServerProvider.
// Summary: Initializes a provider for the MCP server.
// Parameters:
//   - server: *mcp.Server. The server instance to wrap.
//
// Returns:
//   - MCPServerProvider: The initialized provider.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
	return &mcpServerProvider{server: server}
}
