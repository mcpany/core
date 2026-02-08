// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestTool_Execute_Coverage(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	t.Run("init_error", func(t *testing.T) {
		// Create a tool with an invalid regex in policy to trigger initError
		invalidPolicy := configv1.CallPolicy_builder{
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{
					NameRegex: proto.String("[invalid-regex"), // Missing closing bracket
				}.Build(),
			},
		}.Build()

		callDef := configv1.SqlCallDefinition_builder{
			Query: proto.String("SELECT 1"),
		}.Build()

		toolInstance := NewTool(
			v1.Tool_builder{Name: proto.String("broken_tool")}.Build(),
			db,
			callDef,
			[]*configv1.CallPolicy{invalidPolicy},
			"broken_tool_call",
		)

		req := &tool.ExecutionRequest{
			ToolName:   "broken_tool",
			ToolInputs: []byte("{}"),
		}

		_, err := toolInstance.Execute(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to compile call policies")
	})

	t.Run("missing_param_nil", func(t *testing.T) {
		callDef := configv1.SqlCallDefinition_builder{
			Query:          proto.String("SELECT * FROM users WHERE age = ? AND name = ?"),
			ParameterOrder: []string{"age", "name"},
		}.Build()

		toolInstance := NewTool(
			v1.Tool_builder{Name: proto.String("users_tool")}.Build(),
			db,
			callDef,
			nil,
			"users_call",
		)

		// Expect query with arguments: age=25, name=nil (missing in input)
		mock.ExpectQuery("SELECT \\* FROM users WHERE age = \\? AND name = \\?").
			WithArgs(float64(25), nil).
			WillReturnRows(sqlmock.NewRows([]string{"id"}))

		inputs := map[string]interface{}{
			"age": 25,
			// "name" is missing
		}
		inputsBytes, _ := json.Marshal(inputs)

		req := &tool.ExecutionRequest{
			ToolName:   "users_tool",
			ToolInputs: inputsBytes,
		}

		_, err := toolInstance.Execute(ctx, req)
		require.NoError(t, err)
	})

	t.Run("empty_result", func(t *testing.T) {
		callDef := configv1.SqlCallDefinition_builder{
			Query: proto.String("SELECT * FROM users WHERE 1=0"),
		}.Build()

		toolInstance := NewTool(
			v1.Tool_builder{Name: proto.String("empty_tool")}.Build(),
			db,
			callDef,
			nil,
			"empty_call",
		)

		mock.ExpectQuery("SELECT \\* FROM users WHERE 1=0").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"})) // No rows added

		req := &tool.ExecutionRequest{
			ToolName:   "empty_tool",
			ToolInputs: []byte("{}"),
		}

		result, err := toolInstance.Execute(ctx, req)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		assert.NotNil(t, results)
		assert.Empty(t, results)
	})

	t.Run("null_values", func(t *testing.T) {
		callDef := configv1.SqlCallDefinition_builder{
			Query: proto.String("SELECT id, name FROM users"),
		}.Build()

		toolInstance := NewTool(
			v1.Tool_builder{Name: proto.String("null_tool")}.Build(),
			db,
			callDef,
			nil,
			"null_call",
		)

		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, nil) // name is NULL

		mock.ExpectQuery("SELECT id, name FROM users").
			WillReturnRows(rows)

		req := &tool.ExecutionRequest{
			ToolName:   "null_tool",
			ToolInputs: []byte("{}"),
		}

		result, err := toolInstance.Execute(ctx, req)
		require.NoError(t, err)

		results, ok := result.([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, results, 1)
		assert.EqualValues(t, 1, results[0]["id"])
		assert.Nil(t, results[0]["name"])
	})

	t.Run("rows_iteration_error", func(t *testing.T) {
		callDef := configv1.SqlCallDefinition_builder{
			Query: proto.String("SELECT * FROM large_table"),
		}.Build()

		toolInstance := NewTool(
			v1.Tool_builder{Name: proto.String("iter_error_tool")}.Build(),
			db,
			callDef,
			nil,
			"iter_error_call",
		)

		// Simulate error immediately (on first row)
		rows := sqlmock.NewRows([]string{"id"}).
			AddRow(1).
			RowError(0, errors.New("connection lost"))

		mock.ExpectQuery("SELECT \\* FROM large_table").
			WillReturnRows(rows)

		req := &tool.ExecutionRequest{
			ToolName:   "iter_error_tool",
			ToolInputs: []byte("{}"),
		}

		_, err := toolInstance.Execute(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error during row iteration")
		assert.Contains(t, err.Error(), "connection lost")
	})

	t.Run("close_error", func(t *testing.T) {
		callDef := configv1.SqlCallDefinition_builder{
			Query: proto.String("SELECT 1"),
		}.Build()

		toolInstance := NewTool(
			v1.Tool_builder{Name: proto.String("close_error_tool")}.Build(),
			db,
			callDef,
			nil,
			"close_error_call",
		)

		// CloseError is propagated to rows.Err() because rows.Next() calls rows.Close() on completion.
		// Therefore, Execute catches it in rows.Err() check.
		rows := sqlmock.NewRows([]string{"id"}).
			AddRow(1).
			CloseError(errors.New("close failed"))

		mock.ExpectQuery("SELECT 1").
			WillReturnRows(rows)

		req := &tool.ExecutionRequest{
			ToolName:   "close_error_tool",
			ToolInputs: []byte("{}"),
		}

		_, err := toolInstance.Execute(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error during row iteration")
		assert.Contains(t, err.Error(), "close failed")
	})
}
