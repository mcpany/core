// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LazyMCPConfig represents the public LazyMCPConfig entity.
//
// Summary: Defines the structured data model representing a mcp config.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type LazyMCPConfig struct {
	Enabled   bool    `json:"enabled"`
	Threshold float64 `json:"threshold"`
	CacheTTL  int     `json:"cache_ttl"`
}

// LazyMCPMiddleware represents the public LazyMCPMiddleware entity.
//
// Summary: Defines the structured data model representing a mcp middleware.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type LazyMCPMiddleware struct {
	config LazyMCPConfig
}

// NewLazyMCPMiddleware serves as a public interface for interacting with NewLazyMCPMiddleware.
//
// Summary: Constructs and returns an initialized lazy mcp middleware ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewLazyMCPMiddleware(config LazyMCPConfig) *LazyMCPMiddleware {
	return &LazyMCPMiddleware{
		config: config,
	}
}

// FilterTools serves as a public interface for interacting with FilterTools.
//
// Summary: Filter the tools appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *LazyMCPMiddleware) FilterTools(res *mcp.ListToolsResult, intent string) *mcp.ListToolsResult {
	if !m.config.Enabled || intent == "" {
		return res
	}

	var filtered []*mcp.Tool
	intentLower := strings.ToLower(intent)

	for _, tool := range res.Tools {
		// A naive substring check acts as a stand-in for the "similarity-based approach".
		// In a production system this would use Levenshtein distance or cosine similarity over embeddings.
		if strings.Contains(strings.ToLower(tool.Name), intentLower) ||
			strings.Contains(strings.ToLower(tool.Description), intentLower) {
			filtered = append(filtered, tool)
		}
	}

	return &mcp.ListToolsResult{Tools: filtered}
}
