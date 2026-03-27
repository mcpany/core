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

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Tool implements the Tool interface for a tool that executes a SQL query.
//
// Summary. Represents a Tool.
type Tool struct {
	tool        *v1.Tool
	mcpTool     *mcp.Tool
	mcpToolOnce sync.Once
	db          *sql.DB
	callDef     *configv1.SqlCallDefinition
	policies    []*tool.CompiledCallPolicy
	callID      string
	initError   error
}

// NewTool provides newtool functionality.
//
// Summary: NewTool.
//
// Parameters.
//   - t: The parameter.
//   - db: The parameter.
//   - callDef: The parameter.
//   - policies: The parameter.
//   - callID: The parameter.
//
// Returns.
//   - result: The result.
func NewTool(t *v1.Tool, db *sql.DB, callDef *configv1.SqlCallDefinition, policies []*configv1.CallPolicy, callID string) *Tool {
	compiled, err := tool.CompileCallPolicies(policies)
	to := &Tool{
		tool:     t,
		db:       db,
		callDef:  callDef,
		policies: compiled,
		callID:   callID,
	}
	if err != nil {
		to.initError = fmt.Errorf("failed to compile call policies: %w", err)
	}
	return to
}

// Tool provides tool functionality.
//
// Summary: Tool.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (t *Tool) Tool() *v1.Tool {
	return t.tool
}

// MCPTool provides mcptool functionality.
//
// Summary: MCPTool.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (t *Tool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = tool.ConvertProtoToMCPTool(t.tool)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.tool.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

// GetCacheConfig provides getcacheconfig functionality.
//
// Summary: GetCacheConfig.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func (t *Tool) GetCacheConfig() *configv1.CacheConfig {
	if t.callDef == nil {
		return nil
	}
	return t.callDef.GetCache()
}

// Execute provides execute functionality.
//
// Summary: Execute.
//
// Parameters.
//   - ctx: The parameter.
//   - req: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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
