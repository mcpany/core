// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/tool"
)

// PolicyFirewallConfig defines the configuration for the Policy Firewall Engine.
//
// Summary: Represents the configuration for the Policy Firewall.
type PolicyFirewallConfig struct {
	Enabled       bool     `json:"enabled"`
	BlockedTools  []string `json:"blocked_tools"`
	AllowedTools  []string `json:"allowed_tools"`
	DefaultAction string   `json:"default_action"` // "allow" or "deny"
}

// PolicyFirewallMiddleware enforces execution policies for tool calls.
//
// Summary: Represents middleware that enforces policy rules on tool execution.
type PolicyFirewallMiddleware struct {
	config PolicyFirewallConfig
}

// NewPolicyFirewallMiddleware creates a new PolicyFirewallMiddleware.
//
// Summary: Creates a new instance of PolicyFirewallMiddleware.
//
// Parameters:
//   - config (PolicyFirewallConfig): The configuration settings specifying allowed/blocked tools.
//
// Returns:
//   - *PolicyFirewallMiddleware: A new instance of the middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewPolicyFirewallMiddleware(config PolicyFirewallConfig) *PolicyFirewallMiddleware {
	if config.DefaultAction == "" {
		config.DefaultAction = "deny" // secure by default
	}
	return &PolicyFirewallMiddleware{
		config: config,
	}
}

// Execute checks if the tool is allowed by the firewall policy.
//
// Summary: Validates that the requested tool is allowed by the firewall policy before execution.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*tool.ExecutionRequest): The execution request detailing the tool to be called.
//   - next (tool.ExecutionFunc): The next execution function in the middleware chain.
//
// Returns:
//   - any: The result of the tool execution if permitted.
//   - error: An error if access is denied or execution fails.
//
// Errors:
//   - Returns an error if the requested tool is explicitly blocked.
//   - Returns an error if the requested tool is not explicitly allowed and default action is deny.
//
// Side Effects:
//   - Executes the next function in the chain if access is granted.
func (m *PolicyFirewallMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	// 1. Check Blocklist (Highest Priority)
	for _, blocked := range m.config.BlockedTools {
		if req.ToolName == blocked || (strings.HasSuffix(blocked, ".*") && strings.HasPrefix(req.ToolName, strings.TrimSuffix(blocked, ".*"))) {
			return nil, fmt.Errorf("policy firewall: access denied for tool '%s' (blocked)", req.ToolName)
		}
	}

	// 2. Check Allowlist
	isAllowed := false
	for _, allowed := range m.config.AllowedTools {
		if req.ToolName == allowed || (strings.HasSuffix(allowed, ".*") && strings.HasPrefix(req.ToolName, strings.TrimSuffix(allowed, ".*"))) {
			isAllowed = true
			break
		}
	}

	if isAllowed {
		return next(ctx, req)
	}

	// 3. Fallback to Default Action
	if strings.ToLower(m.config.DefaultAction) == "allow" {
		return next(ctx, req)
	}

	return nil, fmt.Errorf("policy firewall: access denied for tool '%s' (default deny)", req.ToolName)
}
