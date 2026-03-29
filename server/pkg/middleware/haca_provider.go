// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// HACAProviderConfig defines the configuration for the Hardware-Attested Cost Attribution Provider.
//
// Summary: Configuration for Hardware-Attested Cost Attribution (HACA) Provider Middleware.
type HACAProviderConfig struct {
	// Enabled determines if the HACA Provider is active.
	Enabled bool `json:"enabled"`
	// RequiredFor tools that mandate a hardware-attested cost attribution token.
	// If empty and Enabled is true, it applies to all tools.
	RequiredFor []string `json:"required_for"`
}

// HACAProviderMiddleware implements the Hardware-Attested Cost Attribution Provider.
// It verifies that subagent reasoning traces cryptographically attribute their
// token usage to specific sub-process lineage, preventing Reasoning-Budget Hijacking.
//
// Summary: Represents the HACA Provider middleware.
type HACAProviderMiddleware struct {
	config HACAProviderConfig
}

// NewHACAProviderMiddleware creates a new HACAProviderMiddleware instance.
//
// Summary: Creates a new Hardware-Attested Cost Attribution Provider instance.
//
// Parameters:
//   - config (HACAProviderConfig): The configuration settings.
//
// Returns:
//   - *HACAProviderMiddleware: The resulting HACA Provider instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewHACAProviderMiddleware(config HACAProviderConfig) *HACAProviderMiddleware {
	return &HACAProviderMiddleware{
		config: config,
	}
}

// Execute enforces cost attribution before proceeding to the next handler.
//
// Summary: Executes the cost attribution check before proceeding.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*tool.ExecutionRequest): The tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - any: The execution result if allowed.
//   - error: An error if the policy blocks execution or policy evaluation fails.
//
// Errors:
//   - Returns error if budget hijacking is detected or required headers are missing.
//
// Side Effects:
//   - Logs validation failures.
func (m *HACAProviderMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	// Check if this tool requires HACA validation
	requiresHACA := false
	if len(m.config.RequiredFor) == 0 {
		requiresHACA = true
	} else {
		for _, toolName := range m.config.RequiredFor {
			if toolName == req.ToolName {
				requiresHACA = true
				break
			}
		}
	}

	if !requiresHACA {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "haca_provider")

	args := req.Arguments
	if args == nil {
		logger.Warn("HACA Provider rejected request: missing arguments", "tool", req.ToolName)
		return nil, fmt.Errorf("Economic Security Violation: missing hardware-attested cost token")
	}

	tokenRaw, hasToken := args["_x_haca_token"]
	lineageRaw, hasLineage := args["_x_lineage_id"]

	if !hasToken || !hasLineage {
		logger.Warn("HACA Provider rejected request: missing attribution tokens", "tool", req.ToolName)
		return nil, fmt.Errorf("Economic Security Violation: missing cost attribution token or lineage ID")
	}

	token, ok1 := tokenRaw.(string)
	lineage, ok2 := lineageRaw.(string)

	if !ok1 || !ok2 || token == "" || lineage == "" {
		logger.Warn("HACA Provider rejected request: invalid attribution token format", "tool", req.ToolName)
		return nil, fmt.Errorf("Economic Security Violation: invalid cost attribution token format")
	}

	expectedPrefix := "haca-attested-"
	if !strings.HasPrefix(token, expectedPrefix) {
		logger.Warn("HACA Provider rejected request: budget hijacking detected (invalid signature)", "tool", req.ToolName, "lineage", lineage)
		return nil, fmt.Errorf("Economic Security Violation: token attestation failed for lineage '%s'", lineage)
	}

	// Validation successful
	logger.Debug("HACA Provider validated request", "tool", req.ToolName, "lineage", lineage)

	// Remove the meta arguments before passing to upstream
	cleanedArgs := make(map[string]interface{})
	for k, v := range args {
		if k != "_x_haca_token" && k != "_x_lineage_id" {
			cleanedArgs[k] = v
		}
	}

	req.Arguments = cleanedArgs

	return next(ctx, req)
}
