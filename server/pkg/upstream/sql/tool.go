// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package sql provides a SQL upstream implementation.
package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool implements the Tool interface for a tool that executes a SQL query.
//
// Summary: Tool implements the Tool interface for a tool that executes a SQL query.
//
// Summary: Tool implements the Tool interface for a tool that executes a SQL query.
type Tool struct {
	tool        *v1.Tool
	mcpTool     *mcp.Tool
	mcpToolOnce sync.Once
	db          *sql.DB
	callDef     *configv1.SqlCallDefinition
	policies    []*tool.CompiledCallPolicy
	callID      string
	initError   error
// NewTool creates a new SQL Tool.
//
// Summary: NewTool creates a new SQL Tool.
//
// Parameters:
//   - t (*v1.Tool): The provided t data.
//   - db (*sql.DB): The provided db data.
//   - callDef (*configv1.SqlCallDefinition): The provided calldef data.
//   - policies ([]*configv1.CallPolicy): The provided policies data.
//   - callID (string): The textual representation of callid.
//
// Returns:
//   - *Tool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - *Tool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func NewTool(t *v1.Tool, db *sql.DB, callDef *configv1.SqlCallDefinition, policies []*configv1.CallPolicy, callID string) *Tool {
	compiled, err := tool.CompileCallPolicies(policies)
	to := &Tool{
		tool:     t,
		db:       db,
		callDef:  callDef,
		policies: compiled,
		callID:   callID,
// Tool returns the protobuf definition of the tool.
//
// Summary: Tool returns the protobuf definition of the tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *v1.Tool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Tool returns the protobuf definition of the tool.
//
// Summary: Tool returns the protobuf definition of the tool.
//
// MCPTool returns the MCP tool definition.
//
// Summary: MCPTool returns the MCP tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - *mcp.Tool: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *Tool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool returns the MCP tool definition.
//
// Summary: MCPTool returns the MCP tool definition.
// GetCacheConfig returns the cache configuration for the tool.
//
// Summary: GetCacheConfig returns the cache configuration for the tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - *configv1.CacheConfig: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *Tool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
// Execute runs the SQL query with the provided inputs.
//
// Summary: Execute runs the SQL query with the provided inputs.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*tool.ExecutionRequest): The incoming request payload.
//
// Returns:
//   - any: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
//
// Returns:
//   - *configv1.CacheConfig: The resulting object or data structure.
//
// Errors:
//   - None.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *Tool) GetCacheConfig() *configv1.CacheConfig {
	if t.callDef == nil {
		return nil
	}
	return t.callDef.GetCache()
}

// Execute runs the SQL query with the provided inputs.
//
// Summary: Execute runs the SQL query with the provided inputs.
//
// Parameters:
//   - ctx (context.Context): The cancellation and deadline context.
//   - req (*tool.ExecutionRequest): The incoming request payload.
//
// Returns:
//   - any: The resulting object or data structure.
//   - error: An error if the execution fails, otherwise nil.
//
// Errors:
//   - Returns an error if the operation fails, invalid input is provided, or a downstream dependency fails.
//
// Side Effects:
//   - May modify internal state or perform external network calls.
func (t *Tool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if t.initError != nil {
		return nil, t.initError
	}

	if logging.GetLogger().Enabled(ctx, slog.LevelDebug) {
		// Use util.RedactJSON directly as prettyPrint is not available in this package
		// and we know it is JSON.
		logging.GetLogger().Debug("executing tool", "tool", req.ToolName, "inputs", string(util.RedactJSON(req.ToolInputs)))
	}
	defer metrics.MeasureSince([]string{"sql", "request", "latency"}, time.Now())

	if allowed, err := tool.EvaluateCompiledCallPolicy(t.policies, t.tool.GetName(), t.callID, req.ToolInputs); err != nil {
		return nil, fmt.Errorf("failed to evaluate call policy: %w", err)
	} else if !allowed {
		return nil, fmt.Errorf("tool execution blocked by policy")
	}

	var inputs map[string]any
	if err := json.Unmarshal(req.ToolInputs, &inputs); err != nil {
		metrics.IncrCounter([]string{"sql", "request", "error"}, 1)
		return nil, fmt.Errorf("failed to unmarshal tool inputs: %w", err)
	}

	// Prepare arguments based on parameter_order
	args := make([]any, 0, len(t.callDef.GetParameterOrder()))
	for _, paramName := range t.callDef.GetParameterOrder() {
		val, ok := inputs[paramName]
		if !ok {
			// If missing, pass nil.
			args = append(args, nil)
		} else {
			args = append(args, val)
		}
	}

	rows, err := t.db.QueryContext(ctx, t.callDef.GetQuery(), args...)
	if err != nil {
		metrics.IncrCounter([]string{"sql", "request", "error"}, 1)
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logging.GetLogger().Warn("Failed to close rows", "error", err)
		}
	}()

	// Get column names
	columns, err := rows.Columns()
	if err != nil {
		metrics.IncrCounter([]string{"sql", "request", "error"}, 1)
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	results := []map[string]any{}

	for rows.Next() {
		// Create a slice of interface{} to hold values
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			metrics.IncrCounter([]string{"sql", "request", "error"}, 1)
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		rowMap := make(map[string]any)
		for i, col := range columns {
			val := values[i]

			// Handle []byte as string for better JSON output
			if b, ok := val.([]byte); ok {
				rowMap[col] = string(b)
			} else {
				rowMap[col] = val
			}
		}
		results = append(results, rowMap)
	}

	if err := rows.Err(); err != nil {
		metrics.IncrCounter([]string{"sql", "request", "error"}, 1)
		return nil, fmt.Errorf("error during row iteration: %w", err)
	}

	metrics.IncrCounter([]string{"sql", "request", "success"}, 1)
	return results, nil
}
