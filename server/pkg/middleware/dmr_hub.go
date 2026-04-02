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

// DMRHubConfig defines the configuration for the Dynamic Mesh Resilience Hub.
//
// Summary: Configuration for Dynamic Mesh Resilience Hub.
type DMRHubConfig struct {
	// Enabled determines if the DMR Hub is active.
	Enabled bool `json:"enabled"`
	// StatefulTools is a list of tools that require state migration verification.
	// If empty and Enabled is true, it applies to all tools.
	StatefulTools []string `json:"stateful_tools"`
}

// DMRHub implements the Dynamic Mesh Resilience Hub middleware.
// It verifies Zero-Knowledge State Attestation (ZKSA) migrations between
// physical nodes upon subagent failure.
//
// Summary: Represents the DMR Hub middleware.
// NewDMRHub creates a new DMRHub middleware instance.
// Summary: Creates a new Dynamic Mesh Resilience Hub instance.
// Parameters:
//   - config (DMRHubConfig): The configuration settings.
//
// Returns:
//   - *DMRHub: The resulting DMR Hub instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
//
// Execute enforces state migration proofs before proceeding to the next handler.
// Summary: Executes the node status and ZKSA proof checks.
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*tool.ExecutionRequest): The tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - any: The execution result if allowed.
//   - error: An error if the migration proof is invalid or missing during failure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type DMRHub struct {
	config DMRHubConfig
}

func NewDMRHub(config DMRHubConfig) *DMRHub {
	return &DMRHub{
		config: config,
	}
}

func (h *DMRHub) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !h.config.Enabled {
		return next(ctx, req)
	}

	// Check if this tool requires state validation
	requiresValidation := false
	if len(h.config.StatefulTools) == 0 {
		requiresValidation = true
	} else {
		for _, toolName := range h.config.StatefulTools {
			if toolName == req.ToolName {
				requiresValidation = true
				break
			}
		}
	}

	if !requiresValidation {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "dmr_hub")

	args := req.Arguments
	if args == nil {
		// Proceed if no arguments, assuming node is healthy by default if not specified
		// In a real system, we might require the status header for all stateful calls.
		return next(ctx, req)
	}

	statusRaw, hasStatus := args["_x_dmr_node_status"]
	if !hasStatus {
		// Proceed if no status is reported, assuming healthy
		return next(ctx, req)
	}

	status, ok := statusRaw.(string)
	if !ok {
		logger.Warn("DMR Hub rejected request: invalid node status format", "tool", req.ToolName)
		return nil, fmt.Errorf("DMR Error: invalid node status format")
	}

	if status == "healthy" {
		// Node is healthy, clean up arg and proceed
		return h.proceed(ctx, req, next)
	}

	if status != "failed" && status != "migrating" {
		logger.Warn("DMR Hub rejected request: unknown node status", "tool", req.ToolName, "status", status)
		return nil, fmt.Errorf("DMR Error: unknown node status '%s'", status)
	}

	// Node is failed or migrating, require ZKSA proof
	proofRaw, hasProof := args["_x_zksa_migration_proof"]
	if !hasProof {
		logger.Warn("DMR Hub rejected request: missing migration proof for failed node", "tool", req.ToolName)
		return nil, fmt.Errorf("DMR Error: migration required, missing ZKSA proof")
	}

	proof, okProof := proofRaw.(string)
	if !okProof || proof == "" {
		logger.Warn("DMR Hub rejected request: invalid migration proof format", "tool", req.ToolName)
		return nil, fmt.Errorf("DMR Error: invalid ZKSA migration proof format")
	}

	// Validate ZKSA proof (simulated with a prefix check)
	expectedPrefix := "zksa-proof-"
	if !strings.HasPrefix(proof, expectedPrefix) {
		logger.Warn("DMR Hub rejected request: invalid ZKSA migration proof", "tool", req.ToolName)
		return nil, fmt.Errorf("DMR Error: invalid ZKSA migration proof")
	}

	// Attestation successful
	logger.Debug("DMR Hub validated state migration", "tool", req.ToolName)

	return h.proceed(ctx, req, next)
}

func (h *DMRHub) proceed(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	// Clean up internal arguments before passing upstream
	cleanedArgs := make(map[string]interface{})
	for k, v := range req.Arguments {
		if k != "_x_dmr_node_status" && k != "_x_zksa_migration_proof" {
			cleanedArgs[k] = v
		}
	}
	req.Arguments = cleanedArgs
	return next(ctx, req)
}
