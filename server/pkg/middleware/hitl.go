// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// HITLConfig defines the configuration for Human-In-The-Loop approval flows.
//
// Summary: Represents the configuration for Human-In-The-Loop (HITL) approval flows.
type HITLConfig struct {
	Enabled        bool     `json:"enabled"`
	SensitiveTools []string `json:"sensitive_tools"`
	RequireMFA     bool     `json:"require_mfa"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

// HITLApprovalRequest represents a request for human approval published to the bus.
//
// Summary: Represents a request for human approval published to the message bus.
type HITLApprovalRequest struct {
	ExecutionID string `json:"execution_id"`
	ToolName    string `json:"tool_name"`
	RequireMFA  bool   `json:"require_mfa"`
}

// HITLApprovalResponse represents the response from the human operator.
//
// Summary: Represents the response from a human operator regarding a tool execution approval request.
type HITLApprovalResponse struct {
	ExecutionID string `json:"execution_id"`
	Approved    bool   `json:"approved"`
	MFAToken    string `json:"mfa_token,omitempty"`
}

// HITLMiddleware enforces Human-In-The-Loop approvals for sensitive actions.
//
// Summary: Represents middleware that enforces Human-In-The-Loop (HITL) approvals for sensitive actions.
type HITLMiddleware struct {
	config HITLConfig
	bus    *bus.Provider
}

// NewHITLMiddleware creates a new HITLMiddleware.
//
// Summary: Creates a new instance of HITLMiddleware.
//
// Parameters:
//   - config (HITLConfig): The configuration settings for HITL approvals.
//   - busProvider (*bus.Provider): The message bus provider used to publish and subscribe to approval events.
//
// Returns:
//   - *HITLMiddleware: A new instance of the middleware.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewHITLMiddleware(config HITLConfig, busProvider *bus.Provider) *HITLMiddleware {
	return &HITLMiddleware{
		config: config,
		bus:    busProvider,
	}
}

// Execute checks if the tool requires HITL approval before proceeding.
//
// Summary: Validates whether a requested tool execution requires human approval and manages the approval flow.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*tool.ExecutionRequest): The execution request detailing the tool to be called.
//   - next (tool.ExecutionFunc): The next execution function in the middleware chain.
//
// Returns:
//   - any: The result of the tool execution if permitted.
//   - error: An error if approval is denied, times out, or fails.
//
// Errors:
//   - Returns an error if the request to get the hitl request bus fails.
//   - Returns an error if the request to get the hitl response bus fails.
//   - Returns an error if publishing the approval request fails.
//   - Returns an error if the execution times out or context is cancelled.
//   - Returns an error if the human operator denies the request.
//
// Side Effects:
//   - Publishes approval requests to the message bus.
//   - Subscribes to response topics on the message bus.
//   - Blocks execution pending a response from the message bus or a timeout.
func (m *HITLMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	isSensitive := false
	for _, t := range m.config.SensitiveTools {
		if t == req.ToolName || (strings.HasSuffix(t, ".*") && strings.HasPrefix(req.ToolName, strings.TrimSuffix(t, ".*"))) {
			isSensitive = true
			break
		}
	}

	if !isSensitive {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "hitl_middleware")
	logger.Warn("HITL Middleware intercepted sensitive tool execution. Suspending for user approval.", "tool", req.ToolName, "require_mfa", m.config.RequireMFA)

	executionID := uuid.New().String()

	// 1. Get bus topics
	reqBus, err := bus.GetBus[HITLApprovalRequest](m.bus, "hitl.requests")
	if err != nil {
		return nil, fmt.Errorf("failed to get hitl request bus: %w", err)
	}

	resBus, err := bus.GetBus[HITLApprovalResponse](m.bus, "hitl.responses."+executionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get hitl response bus: %w", err)
	}

	// 2. Set up channel to receive the response
	responseCh := make(chan HITLApprovalResponse, 1)

	timeout := time.Duration(m.config.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second // Default 5 minutes
	}
	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 3. Subscribe to the unique response topic for this execution
	unsubscribe := resBus.SubscribeOnce(subCtx, "hitl.responses."+executionID, func(res HITLApprovalResponse) {
		responseCh <- res
	})
	defer unsubscribe()

	// 4. Publish the approval request
	err = reqBus.Publish(ctx, "hitl.requests", HITLApprovalRequest{
		ExecutionID: executionID,
		ToolName:    req.ToolName,
		RequireMFA:  m.config.RequireMFA,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to publish hitl approval request: %w", err)
	}

	// 5. Wait for approval, timeout, or context cancellation
	select {
	case <-subCtx.Done():
		return nil, fmt.Errorf("execution suspended for HITL approval: timeout reached or context cancelled")
	case response := <-responseCh:
		if !response.Approved {
			return nil, fmt.Errorf("execution suspended for HITL approval: human denied request")
		}
		if m.config.RequireMFA && response.MFAToken == "" {
			return nil, fmt.Errorf("execution suspended for HITL approval: missing required MFA token")
		}
		// In a real implementation we would validate the MFA token here.
		// For the purpose of this demonstration, we just check if it's non-empty.

		// Proceed if approved
		return next(ctx, req)
	}
}
