// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// AMTBrokerConfig configuration for the Attested Mesh Tunneling Broker.
type AMTBrokerConfig struct {
	Enabled bool
}

// AMTBroker represents the Attested Mesh Tunneling Broker.
// It intercepts remote tool requests, enforces "Mesh Tickets" validation for fast-path resumption,
// and ensures hardware-attested, origin-locked remote executions.
//
// Summary: Attested Mesh Tunneling (AMT) Broker middleware.
type AMTBroker struct {
	config AMTBrokerConfig
}

// NewAMTBroker creates a new AMTBroker middleware.
//
// Summary: Creates a new AMT Broker.
//
// Parameters:
//   - config: AMTBrokerConfig. The configuration for the broker.
//
// Returns:
//   - *AMTBroker: The newly created middleware.
func NewAMTBroker(config AMTBrokerConfig) *AMTBroker {
	return &AMTBroker{
		config: config,
	}
}

// Execute performs the middleware logic, enforcing Mesh Ticket validation.
//
// Summary: Executes the AMT broker logic.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*tool.ExecutionRequest): The tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - any: The execution result if allowed.
//   - error: An error if validation fails or the mesh ticket is missing/invalid.
func (b *AMTBroker) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !b.config.Enabled {
		return next(ctx, req)
	}

	// This is a placeholder for real cryptographic validation of mesh tickets
	// during P2P remote tool execution handshakes.
	// We only enforce ticket validation if the request is destined for a remote mesh node.
	// For simulation/demonstration, we assume any tool name prefixed with "remote_" is a mesh tool.
	isRemote := strings.HasPrefix(req.ToolName, "remote_")

	if isRemote {
		// Fast-path resumption validation
		ticketRaw, hasTicket := req.Arguments["meshTicket"]
		if !hasTicket {
			logging.GetLogger().WarnContext(ctx, "AMT Broker: Missing mesh ticket for remote tool execution", "tool", req.ToolName)
			return nil, errors.New("unauthorized: missing mesh ticket for attested mesh tunneling")
		}

		ticket, ok := ticketRaw.(string)
		if !ok || ticket == "" {
			logging.GetLogger().WarnContext(ctx, "AMT Broker: Invalid mesh ticket format", "tool", req.ToolName)
			return nil, errors.New("unauthorized: invalid mesh ticket format")
		}

		// In a production system, this would verify the cryptographic signature
		// of the mesh ticket against the node's TPM identity.
		if ticket != "valid-mission-bound-ticket" {
			logging.GetLogger().WarnContext(ctx, "AMT Broker: Invalid mesh ticket provided", "tool", req.ToolName, "ticket", ticket)
			return nil, fmt.Errorf("unauthorized: invalid mesh ticket: %s", ticket)
		}

		logging.GetLogger().DebugContext(ctx, "AMT Broker: Mesh ticket validated for fast-path resumption", "tool", req.ToolName)
	}

	return next(ctx, req)
}
