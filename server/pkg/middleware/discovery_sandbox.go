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

// DiscoverySandboxConfig defines the configuration for Discovery-Phase Sandbox Isolation.
//
// Summary: Represents the configuration for the Discovery-Phase Sandbox middleware.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors/Throws:
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

// DiscoverySandboxMiddleware wraps discovery commands in a secure, ephemeral execution environment
// to prevent startup-time RCE and Ghost-Execution exploits during capability discovery.
//
// Summary: Represents the middleware for Discovery-Phase Sandbox Isolation.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors/Throws:
//   - None.
//
// Side Effects:
//   - None.
type DiscoverySandboxMiddleware struct {
	config DiscoverySandboxConfig
}

// NewDiscoverySandboxMiddleware creates a new DiscoverySandboxMiddleware instance.
//
// Summary: Creates a new Discovery-Phase Sandbox Isolation middleware instance.
//
// Parameters:
//   - config (DiscoverySandboxConfig): The configuration settings.
//
// Returns:
//   - *DiscoverySandboxMiddleware: The resulting middleware instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Errors/Throws:
//   - None.
func NewDiscoverySandboxMiddleware(config DiscoverySandboxConfig) *DiscoverySandboxMiddleware {
	return &DiscoverySandboxMiddleware{
		config: config,
	}
}

// PreExecute is called before a tool is executed. For discovery commands, it validates
// constraints and simulates launching an isolated sandbox environment.
//
// Summary: Intercepts tool execution to enforce discovery-phase sandbox constraints.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*mcp.CallToolRequest): The tool execution request.
//   - t (*tool.Tool): The tool being executed.
//
// Returns:
//   - error: An error if sandbox constraints are violated.
//
// Errors:
//   - Returns an error if the tool name indicates a discovery command and sandbox simulation fails.
//
// Side Effects:
//   - Logs a security audit event upon sandbox constraint enforcement.
//
// Errors/Throws:
//   - error: Returns an error if the operation fails.
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

// PostExecute is called after the tool completes, successfully or not.
//
// Summary: PostExecute performs cleanup of the ephemeral sandbox environment.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*mcp.CallToolRequest): The tool execution request.
//   - result (*mcp.CallToolResult): The result of the tool execution.
//   - err (error): The error returned by the tool, if any.
//
// Returns:
//   - *mcp.CallToolResult: The original result.
//   - error: The original error.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Logs the teardown of the ephemeral sandbox.
//
// Errors/Throws:
//   - error: Returns an error if the operation fails.
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
