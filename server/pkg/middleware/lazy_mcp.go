// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// LazyMCPConfig defines the configuration for On-Demand Discovery filtering.
//
// Summary: Represents the configuration for the LazyMCP middleware.
type LazyMCPConfig struct {
	Enabled   bool    `json:"enabled"`
	Threshold float64 `json:"threshold"`
}

// LazyMCPMiddleware filters tools based on a simplistic similarity logic to prevent context pollution.
//
// Summary: Represents the middleware for filtering tools based on intent.
type LazyMCPMiddleware struct {
	config LazyMCPConfig
}

// NewLazyMCPMiddleware creates a new LazyMCPMiddleware.
//
// Summary: Creates a new instance of LazyMCPMiddleware.
//
// Parameters:
//   - config (LazyMCPConfig): The configuration settings for the middleware.
//
// Returns:
//   - *LazyMCPMiddleware: A new instance of the middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewLazyMCPMiddleware(config LazyMCPConfig) *LazyMCPMiddleware {
	return &LazyMCPMiddleware{
		config: config,
	}
}

// FilterTools takes an original tool list result and an intent string (from the context or header),
// and returns a new list of tools that are "similar" to the intent.
//
// Summary: Filters a list of tools based on their similarity to the provided intent.
//
// Parameters:
//   - res (*mcp.ListToolsResult): The original list of tools to filter.
//   - intent (string): The user's intent to match against tool names and descriptions.
//
// Returns:
//   - *mcp.ListToolsResult: A new list of tools that match the given intent.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
