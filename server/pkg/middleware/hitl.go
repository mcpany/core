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
// Summary: Represents the configuration for Human-In-The-Loop approval constraints.
type HITLConfig struct {
	Enabled        bool     `json:"enabled"`
	SensitiveTools []string `json:"sensitive_tools"`
	RequireMFA     bool     `json:"require_mfa"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

// HITLApprovalRequest represents a request for human approval published to the bus.
//
// Summary: Represents an event payload requesting human approval for a specific execution.
type HITLApprovalRequest struct {
	ExecutionID string `json:"execution_id"`
	ToolName    string `json:"tool_name"`
	RequireMFA  bool   `json:"require_mfa"`
}

// HITLApprovalResponse represents the response from the human operator.
//
// Summary: Represents the event payload indicating the outcome of a human approval request.
type HITLApprovalResponse struct {
	ExecutionID string `json:"execution_id"`
	Approved    bool   `json:"approved"`
}

// HITLMiddleware enforces Human-In-The-Loop approvals for sensitive actions.
//
// Summary: Implements the middleware that suspends tool execution pending out-of-band human approval.
type HITLMiddleware struct {
	config HITLConfig
	bus    *bus.Provider
}

// NewHITLMiddleware creates a new HITLMiddleware.
//
// Summary: Initializes and returns a new Human-in-the-Loop middleware instance.
//
// Parameters:
//   - config (HITLConfig): The configuration governing which tools require approval.
//   - busProvider (*bus.Provider): The message bus used to publish and subscribe to approval events.
//
// Returns:
//   - *HITLMiddleware: A new instance of the middleware.
func NewHITLMiddleware(config HITLConfig, busProvider *bus.Provider) *HITLMiddleware {
	return &HITLMiddleware{
		config: config,
		bus:    busProvider,
	}
}

// Execute checks if the tool requires HITL approval before proceeding.
//
// Summary: Suspends execution of sensitive tools and awaits explicit out-of-band human approval via the message bus.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - req (*tool.ExecutionRequest): The details of the tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the execution chain.
//
// Returns:
//   - any: The result of the next handler, typically the tool's response.
//   - error: An error if approval is denied, timeouts occur, or the next handler errors out.
//
// Errors:
//   - Returns an error containing "execution denied by human operator" if the request is rejected.
//   - Returns context.DeadlineExceeded if the approval request times out.
//   - Returns an error if the bus fails to subscribe or decode the response.
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
	responseCh := make(chan bool, 1)

	timeout := time.Duration(m.config.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 300 * time.Second // Default 5 minutes
	}
	subCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// 3. Subscribe to the unique response topic for this execution
	unsubscribe := resBus.SubscribeOnce(subCtx, "hitl.responses."+executionID, func(res HITLApprovalResponse) {
		responseCh <- res.Approved
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
	case approved := <-responseCh:
		if !approved {
			return nil, fmt.Errorf("execution suspended for HITL approval: human denied request")
		}
		// Proceed if approved
		return next(ctx, req)
	}
}
