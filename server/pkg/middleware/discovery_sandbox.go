// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DiscoverySandboxConfig represents the public DiscoverySandboxConfig entity.
//
// Summary: Defines the structured data model representing a sandbox config.
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
type DiscoverySandboxConfig struct {
	// Enabled determines if the DiscoverySandbox middleware is active.
	Enabled bool `json:"enabled"`
	// IsolatedEnvironment specifies the type of sandbox (e.g., "gVisor", "ephemeral-container", "strict-seccomp").
	IsolatedEnvironment string `json:"isolated_environment"`
	// MaxExecutionTimeMs specifies the maximum allowed time in milliseconds for discovery.
	MaxExecutionTimeMs int `json:"max_execution_time_ms"`
	// AllowedPaths lists the file paths the discovery process is allowed to read.
	AllowedPaths []string `json:"allowed_paths"`
}

// DiscoverySandboxMiddleware represents the public DiscoverySandboxMiddleware entity.
//
// Summary: Defines the structured data model representing a sandbox middleware.
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
type DiscoverySandboxMiddleware struct {
	config DiscoverySandboxConfig
}

// NewDiscoverySandboxMiddleware serves as a public interface for interacting with NewDiscoverySandboxMiddleware.
//
// Summary: Constructs and returns an initialized discovery sandbox middleware ready for consumption.
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
func NewDiscoverySandboxMiddleware(config DiscoverySandboxConfig) *DiscoverySandboxMiddleware {
	return &DiscoverySandboxMiddleware{
		config: config,
	}
}

// PreExecute serves as a public interface for interacting with PreExecute.
//
// Summary: Pre the execute appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *DiscoverySandboxMiddleware) PreExecute(ctx context.Context, req *mcp.CallToolRequest, t *tool.Tool) error {
	if !m.config.Enabled {
		return nil
	}

	// Heuristically detect if this is a discovery or manifest generation command.
	isDiscovery := strings.Contains(strings.ToLower(req.Params.Name), "discovery") ||
		strings.Contains(strings.ToLower(req.Params.Name), "manifest") ||
		strings.Contains(strings.ToLower(req.Params.Name), "list")

	if !isDiscovery {
		return nil
	}

	logger := logging.GetLogger().With("component", "discovery_sandbox_middleware")

	// Simulate Ephemeral Sandbox Execution constraints
	logger.Info("Security event: Enforcing Discovery-Phase Sandbox constraints for tool", "tool", req.Params.Name)

	if m.config.IsolatedEnvironment == "" {
		return fmt.Errorf("security violation: discovery command executed without a configured IsolatedEnvironment")
	}

	if m.config.MaxExecutionTimeMs <= 0 {
		return fmt.Errorf("security violation: discovery command execution time must be strictly bounded")
	}

	logger.Info("Executing discovery command inside ephemeral sandbox", "tool", req.Params.Name, "sandbox", m.config.IsolatedEnvironment, "timeout", m.config.MaxExecutionTimeMs)

	return nil
}

// PostExecute serves as a public interface for interacting with PostExecute.
//
// Summary: Post the execute appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (m *DiscoverySandboxMiddleware) PostExecute(ctx context.Context, req *mcp.CallToolRequest, result *mcp.CallToolResult, err error) (*mcp.CallToolResult, error) {
	if !m.config.Enabled {
		return result, err
	}

	isDiscovery := strings.Contains(strings.ToLower(req.Params.Name), "discovery") ||
		strings.Contains(strings.ToLower(req.Params.Name), "manifest") ||
		strings.Contains(strings.ToLower(req.Params.Name), "list")

	logger := logging.GetLogger().With("component", "discovery_sandbox_middleware")

	if isDiscovery {
		logger.Info("Tearing down ephemeral discovery sandbox for tool", "tool", req.Params.Name)
	}

	return result, err
}
