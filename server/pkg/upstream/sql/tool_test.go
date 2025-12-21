// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestTool_Execute_Errors(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	// Create table for valid query testing
	_, err = db.Exec("CREATE TABLE test (id INT)")
	require.NoError(t, err)

	callDef := &configv1.SqlCallDefinition{
		Query:          ptr("SELECT * FROM test WHERE id = ?"),
		ParameterOrder: []string{"id"},
	}

	// Create tool
	mcpTool := &v1.Tool{Name: ptr("test-tool")}
	sqlTool := NewTool(mcpTool, db, callDef)

	t.Run("Malformed JSON Inputs", func(t *testing.T) {
		req := &tool.ExecutionRequest{
			ToolName:   "test-tool",
			ToolInputs: json.RawMessage(`{invalid-json}`),
		}
		_, err := sqlTool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})

	t.Run("Missing Parameter", func(t *testing.T) {
		// "id" is in ParameterOrder, but we provide empty input.
		// The code executes query with nil.
		req := &tool.ExecutionRequest{
			ToolName:   "test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		res, err := sqlTool.Execute(context.Background(), req)
		require.NoError(t, err)
		// Should return empty list as id=NULL matches nothing (usually)
		resSlice, ok := res.([]map[string]any)
		require.True(t, ok)
		assert.Empty(t, resSlice)
	})

	t.Run("Query Error", func(t *testing.T) {
		badCallDef := &configv1.SqlCallDefinition{
			Query: ptr("SELECT * FROM non_existent_table"),
		}
		badTool := NewTool(mcpTool, db, badCallDef)
		req := &tool.ExecutionRequest{
			ToolName:   "test-tool",
			ToolInputs: json.RawMessage(`{}`),
		}
		_, err := badTool.Execute(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute query")
	})

	t.Run("Closed DB", func(t *testing.T) {
		dbClosed, err := sql.Open("sqlite", ":memory:")
		require.NoError(t, err)
		_ = dbClosed.Close()

		toolClosed := NewTool(mcpTool, dbClosed, callDef)
		req := &tool.ExecutionRequest{
			ToolName:   "test-tool",
			ToolInputs: json.RawMessage(`{"id": 1}`),
		}
		_, err = toolClosed.Execute(context.Background(), req)
		assert.Error(t, err)
	})
}
