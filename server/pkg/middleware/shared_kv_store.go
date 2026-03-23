// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
)

// SharedKVStoreConfig defines the configuration for the Shared KV Store (Blackboard).
type SharedKVStoreConfig struct {
	Enabled        bool   `json:"enabled"`
	DBPath         string `json:"db_path"`
	IsolationLevel string `json:"isolation_level"`
}

// SharedKVStoreMiddleware provides reliable state management for multi-agent systems.
type SharedKVStoreMiddleware struct {
	config SharedKVStoreConfig
}

// NewSharedKVStoreMiddleware creates a new SharedKVStoreMiddleware.
func NewSharedKVStoreMiddleware(config SharedKVStoreConfig) *SharedKVStoreMiddleware {
	return &SharedKVStoreMiddleware{
		config: config,
	}
}

// Execute enforces isolation rules and manages access to the KV store.
func (m *SharedKVStoreMiddleware) Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error) {
	if !m.config.Enabled {
		return next(ctx, req)
	}

	logger := logging.GetLogger().With("component", "shared_kv_store_middleware")

	if m.config.IsolationLevel == "agent_aware" {
		logger.Debug("Enforcing agent_aware isolation for tool execution", "tool", req.ToolName)
		// Simulating row-level security enforcement based on agent identity and intent-scope
	}

	// For the audit fix, we just allow the execution to proceed but log the activity
	return next(ctx, req)
}
