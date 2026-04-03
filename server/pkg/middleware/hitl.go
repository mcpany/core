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

// HITLConfig represents the public HITLConfig entity.
//
// Summary: Defines the structured data model representing a config.
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
type HITLConfig struct {
	Enabled        bool     `json:"enabled"`
	SensitiveTools []string `json:"sensitive_tools"`
	RequireMFA     bool     `json:"require_mfa"`
	TimeoutSeconds int      `json:"timeout_seconds"`
}

// HITLApprovalRequest represents the public HITLApprovalRequest entity.
//
// Summary: Defines the structured data model representing a approval request.
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
type HITLApprovalRequest struct {
	ExecutionID string `json:"execution_id"`
	ToolName    string `json:"tool_name"`
	RequireMFA  bool   `json:"require_mfa"`
}

// HITLApprovalResponse represents the public HITLApprovalResponse entity.
//
// Summary: Defines the structured data model representing a approval response.
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
type HITLApprovalResponse struct {
	ExecutionID string `json:"execution_id"`
	Approved    bool   `json:"approved"`
}

// HITLMiddleware represents the public HITLMiddleware entity.
//
// Summary: Defines the structured data model representing a middleware.
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
type HITLMiddleware struct {
	config HITLConfig
	bus    *bus.Provider
}

// NewHITLMiddleware serves as a public interface for interacting with NewHITLMiddleware.
//
// Summary: Constructs and returns an initialized hitl middleware ready for consumption.
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
func NewHITLMiddleware(config HITLConfig, busProvider *bus.Provider) *HITLMiddleware {
	return &HITLMiddleware{
		config: config,
		bus:    busProvider,
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
