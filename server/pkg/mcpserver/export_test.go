// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import "github.com/modelcontextprotocol/go-sdk/mcp"

// GetRouter returns the server's router. This is for testing purposes only.
// GetRouter returns the server's router. This is for testing purposes only.
// Summary: GetRouter
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return s.router
}

// RouterMiddleware ...
// Summary: RouterMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return s.routerMiddleware(next)
}

// ToolListFilteringMiddleware ...
// Summary: ToolListFilteringMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return s.toolListFilteringMiddleware(next)
}

// ResourceListFilteringMiddleware ...
// Summary: ResourceListFilteringMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return s.resourceListFilteringMiddleware(next)
}

// PromptListFilteringMiddleware ...
// Summary: PromptListFilteringMiddleware
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return s.promptListFilteringMiddleware(next)
}
