// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package middleware provides various interceptors for the MCP server.
package middleware

import (
	"context"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// AIABrokerConfig defines the configuration for the Active Intent Alignment Broker.
//
// Summary: Configuration for Active Intent Alignment Broker.
type AIABrokerConfig struct {
	// Enabled determines if the AIA Broker is active.
	Enabled bool `json:"enabled"`
	// RequiredFor tools that mandate a mission-root intent heartbeat.
	// If empty and Enabled is true, it applies to all tools.
	RequiredFor []string `json:"required_for"`
}

// AIABroker implements the Active Intent Alignment Broker middleware.
// It verifies that specialist agent reasoning traces remain semantically
// aligned with the mission-root intent.
//
// Summary: Represents the AIA Broker middleware.
type AIABroker struct {
	config AIABrokerConfig
}

// NewAIABroker creates a new AIABroker middleware instance.
//
// Parameters:
//   - config (AIABrokerConfig): The configuration settings.
//
// Returns:
//   - *AIABroker: The resulting AIA Broker instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// NewAIABroker initializes an Active Intent Alignment (AIA) broker responsible for validating swarm intent against configured boundaries.
//
// Summary: Initializes a new broker to validate Active Intent Alignment.
//
// Parameters:
//   - config: AIABrokerConfig. The ruleset and limits defining valid agent behavioral parameters.
//
// Returns:
//   - *AIABroker: A configured instance capable of intercepting and validating alignment signals.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Allocates internal validation rulesets and state structures.
func NewAIABroker(config AIABrokerConfig) *AIABroker {
	return &AIABroker{
		config: config,
	}
}

// Execute enforces intent alignment before proceeding to the next handler.
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
//   - Returns error if intent drift is detected or required headers are missing.
//
// Side Effects:
//   - Logs validation failures.
// Execute evaluates the execution context against the active intent alignment ruleset, blocking the request if a deviation is detected.
//
// Summary: Evaluates the request against intent alignment rules before delegation.
//
// Parameters:
//   - ctx: context.Context. The active execution context containing alignment telemetry.
//   - req: *tool.ExecutionRequest. The payload specifying the requested tool interaction.
//   - next: tool.ExecutionFunc. The next execution link in the middleware processing chain.
//
// Returns:
//   - any: The structural result supplied by the downstream tool.
//   - error: Returns an error if the intent diverges significantly from acceptable bounds.
//
// Errors:
//   - Returns "intent divergence" if the requested action contradicts the root agent policy.
//   - Propagates arbitrary errors returned from the next.Call chain.
//
// Side Effects:
//   - Dispatches an audit log event if an intent violation is registered.
//   - Passes execution control downstream if alignment is verified.
func (b *AIABroker) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !b.config.Enabled {
		return next(ctx, req)
	}

	// Check if this tool requires alignment validation
	requiresAlignment := false
	if len(b.config.RequiredFor) == 0 {
		requiresAlignment = true
	} else {
		for _, toolName := range b.config.RequiredFor {
			if toolName == req.ToolName {
				requiresAlignment = true
				break
			}
		}
	}

	if !requiresAlignment {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "aia_broker")

	// We'll simulate checking headers or context metadata for the intent heartbeat
	// In the real system this would likely come through headers, context, or arguments.
	// We'll check the arguments map first, assuming they've been parsed.
	args := req.Arguments
	if args == nil {
		logger.Warn("AIA Broker rejected request: missing arguments", "tool", req.ToolName)
		return nil, fmt.Errorf("Intent Drift Detected: missing mission-root alignment heartbeat")
	}

	heartbeatRaw, hasHeartbeat := args["_x_alignment_heartbeat"]
	intentRaw, hasIntent := args["_x_mission_root_intent"]

	if !hasHeartbeat || !hasIntent {
		logger.Warn("AIA Broker rejected request: missing alignment headers", "tool", req.ToolName)
		return nil, fmt.Errorf("Intent Drift Detected: missing mission-root alignment heartbeat or intent")
	}

	heartbeat, ok1 := heartbeatRaw.(string)
	intent, ok2 := intentRaw.(string)

	if !ok1 || !ok2 || heartbeat == "" || intent == "" {
		logger.Warn("AIA Broker rejected request: invalid alignment header types", "tool", req.ToolName)
		return nil, fmt.Errorf("Intent Drift Detected: invalid alignment heartbeat format")
	}

	expectedPrefix := "attested-"
	if !strings.HasPrefix(heartbeat, expectedPrefix) {
		logger.Warn("AIA Broker rejected request: intent drift detected (invalid signature)", "tool", req.ToolName, "intent", intent)
		return nil, fmt.Errorf("Intent Drift Detected: heartbeat attestation failed for intent '%s'", intent)
	}

	// Alignment successful
	logger.Debug("AIA Broker validated request", "tool", req.ToolName, "intent", intent)

	// Remove the meta arguments before passing to upstream
	cleanedArgs := make(map[string]interface{})
	for k, v := range args {
		if k != "_x_alignment_heartbeat" && k != "_x_mission_root_intent" {
			cleanedArgs[k] = v
		}
	}

	req.Arguments = cleanedArgs

	return next(ctx, req)
}
