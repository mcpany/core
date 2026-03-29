// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// HACAConfig defines the configuration for Hardware-Attested Cost Attribution.
//
// Summary: Configuration for HACA Middleware.
type HACAConfig struct {
	// Enabled determines if the HACA middleware is active.
	Enabled bool `json:"enabled"`
	// MaxTokens sets the maximum allowed token count per lineage branch.
	MaxTokens int64 `json:"max_tokens"`
	// MaxReasoningTime sets the maximum allowed reasoning time in milliseconds per lineage branch.
	MaxReasoningTime int64 `json:"max_reasoning_time"`
}

// HACAMiddleware implements the Hardware-Attested Cost Attribution middleware.
// It attributes token usage and reasoning time to specific sub-process lineages.
//
// Summary: Represents the HACA Middleware.
type HACAMiddleware struct {
	config HACAConfig

	// Internal budget tracking (in-memory for this implementation, typically a registry).
	// Maps lineage hashes to consumed costs.
	mu             sync.RWMutex
	budgetRegistry map[string]*attributedCost
}

type attributedCost struct {
	tokensConsumed    int64
	reasoningTimeMs   int64
}

// NewHACAMiddleware creates a new HACAMiddleware instance.
//
// Summary: Creates a new Hardware-Attested Cost Attribution instance.
//
// Parameters:
//   - config (HACAConfig): The configuration settings.
//
// Returns:
//   - *HACAMiddleware: The resulting HACA middleware instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewHACAMiddleware(config HACAConfig) *HACAMiddleware {
	return &HACAMiddleware{
		config: config,
		budgetRegistry: make(map[string]*attributedCost),
	}
}

// Execute enforces cost attribution before proceeding to the next handler.
//
// Summary: Executes the cost attribution check on the request.
//
// Parameters:
//   - ctx (context.Context): The execution context.
//   - req (*tool.ExecutionRequest): The tool execution request.
//   - next (tool.ExecutionFunc): The next handler in the chain.
//
// Returns:
//   - any: The execution result if allowed.
//   - error: An error if the budget is exceeded or validation fails.
//
// Errors:
//   - Returns error if the budget is exhausted.
//   - Returns error if required headers are missing.
//
// Side Effects:
//   - Updates internal budget registry.
func (m *HACAMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "haca_middleware")

	args := req.Arguments
	if args == nil {
		return next(ctx, req)
	}

	lineageRaw, hasLineage := args["_x_mission_lineage"]
	effortRaw, hasEffort := args["_x_gemini_reasoning_effort"]
	tokensRaw, hasTokens := args["_x_tokens_consumed"]

	if !hasLineage {
		logger.Warn("HACA rejected request: missing mission lineage", "tool", req.ToolName)
		return nil, fmt.Errorf("Cost Attribution Failed: missing mission lineage")
	}

	lineage, ok := lineageRaw.(string)
	if !ok || lineage == "" {
		return nil, fmt.Errorf("Cost Attribution Failed: invalid mission lineage format")
	}

	// Calculate a cryptographic hash for the lineage to simulate hardware attestation verification
	hash := sha256.Sum256([]byte(lineage))
	lineageHash := hex.EncodeToString(hash[:])

	// Parse effort (reasoning time)
	var effort int64
	if hasEffort {
		switch e := effortRaw.(type) {
		case string:
			parsed, err := strconv.ParseInt(e, 10, 64)
			if err == nil {
				effort = parsed
			}
		case int:
			effort = int64(e)
		case float64:
			effort = int64(e)
		}
	}

	// Parse tokens
	var tokens int64
	if hasTokens {
		switch t := tokensRaw.(type) {
		case string:
			parsed, err := strconv.ParseInt(t, 10, 64)
			if err == nil {
				tokens = parsed
			}
		case int:
			tokens = int64(t)
		case float64:
			tokens = int64(t)
		}
	}

	// Lock for budget check and update
	m.mu.Lock()
	cost, exists := m.budgetRegistry[lineageHash]
	if !exists {
		cost = &attributedCost{}
		m.budgetRegistry[lineageHash] = cost
	}

	// Pre-check budget limits
	if m.config.MaxTokens > 0 && (cost.tokensConsumed+tokens) > m.config.MaxTokens {
		m.mu.Unlock()
		logger.Warn("HACA interdicted request: Token budget exhausted", "lineage", lineageHash, "requested", tokens, "consumed", cost.tokensConsumed)
		return nil, fmt.Errorf("Economic Interdiction: Token budget exhausted for mission branch")
	}

	if m.config.MaxReasoningTime > 0 && (cost.reasoningTimeMs+effort) > m.config.MaxReasoningTime {
		m.mu.Unlock()
		logger.Warn("HACA interdicted request: Reasoning time budget exhausted", "lineage", lineageHash, "requested", effort, "consumed", cost.reasoningTimeMs)
		return nil, fmt.Errorf("Economic Interdiction: Reasoning time budget exhausted for mission branch")
	}

	// Attribute costs
	cost.tokensConsumed += tokens
	cost.reasoningTimeMs += effort

	totalTokens := cost.tokensConsumed
	totalTime := cost.reasoningTimeMs
	m.mu.Unlock()

	logger.Debug("HACA cost attributed", "lineage", lineageHash, "total_tokens", totalTokens, "total_time", totalTime)

	// Remove internal headers before passing downstream
	cleanedArgs := make(map[string]interface{})
	for k, v := range args {
		if k != "_x_mission_lineage" && k != "_x_gemini_reasoning_effort" && k != "_x_tokens_consumed" {
			cleanedArgs[k] = v
		}
	}
	req.Arguments = cleanedArgs

	return next(ctx, req)
}
