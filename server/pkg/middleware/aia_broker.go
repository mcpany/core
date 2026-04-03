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

// AIABrokerConfig represents the public AIABrokerConfig entity.
//
// Summary: Defines the structured data model representing a broker config.
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
type AIABrokerConfig struct {
	// Enabled determines if the AIA Broker is active.
	Enabled bool `json:"enabled"`
	// RequiredFor tools that mandate a mission-root intent heartbeat.
	// If empty and Enabled is true, it applies to all tools.
	RequiredFor []string `json:"required_for"`
}

// AIABroker represents the public AIABroker entity.
//
// Summary: Defines the structured data model representing a broker.
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
type AIABroker struct {
	config AIABrokerConfig
}

// NewAIABroker serves as a public interface for interacting with NewAIABroker.
//
// Summary: Constructs and returns an initialized aia broker ready for consumption.
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
func NewAIABroker(config AIABrokerConfig) *AIABroker {
	return &AIABroker{
		config: config,
	}
}

// Execute serves as a public interface for interacting with Execute.
//
// Summary: Execute the  appropriately based on current system conditions.
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
