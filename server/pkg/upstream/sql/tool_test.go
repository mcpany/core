// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestTool_Execute(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	callDef := configv1.SqlCallDefinition_builder{
		Query:          proto.String("SELECT id, name FROM users WHERE age > ?"),
		ParameterOrder: []string{"age"},
	}.Build()

	toolInstance := NewTool(v1.Tool_builder{Name: proto.String("get_users")}.Build(), db, callDef, nil, "get_users_call")

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "Alice").
			AddRow(2, "Bob")

		mock.ExpectQuery("SELECT id, name FROM users WHERE age > ?").
			WithArgs(float64(20)). // JSON numbers are floats
			WillReturnRows(rows)

		inputs := map[string]interface{}{
			"age": 20,
		}
		inputsBytes, _ := json.Marshal(inputs)

		req := &tool.ExecutionRequest{
			ToolName:   "get_users",
			ToolInputs: inputsBytes,
		}

		result, err := toolInstance.Execute(context.Background(), req)
		require.NoError(t, err)

		results := result.([]map[string]interface{})
		assert.Len(t, results, 2)
		// Note: sqlmock returns int64 for integers when using NewRows with AddRow(int)
		// But here we are asserting on what Execute returns.
		// Execute uses rows.Scan. The driver (sqlmock) returns int64.
		// So result should contain int64.
		assert.Equal(t, int64(1), results[0]["id"])
		assert.Equal(t, "Alice", results[0]["name"])
		assert.Equal(t, int64(2), results[1]["id"])
		assert.Equal(t, "Bob", results[1]["name"])
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name FROM users WHERE age > ?").
			WithArgs(float64(20)).
			WillReturnError(assert.AnError)

		inputs := map[string]interface{}{
			"age": 20,
		}
		inputsBytes, _ := json.Marshal(inputs)

		req := &tool.ExecutionRequest{
			ToolName:   "get_users",
			ToolInputs: inputsBytes,
		}

		_, err := toolInstance.Execute(context.Background(), req)
		require.Error(t, err)
	})

	t.Run("input unmarshal error", func(t *testing.T) {
		req := &tool.ExecutionRequest{
			ToolName:   "get_users",
			ToolInputs: []byte("invalid json"),
		}
		_, err := toolInstance.Execute(context.Background(), req)
		require.Error(t, err)
	})

	t.Run("metadata methods", func(t *testing.T) {
		assert.Equal(t, "get_users", toolInstance.Tool().GetName())
		assert.NotNil(t, toolInstance.MCPTool())
		assert.Nil(t, toolInstance.GetCacheConfig())

		// Test with cache config
		cachedCallDef := configv1.SqlCallDefinition_builder{
			Cache: configv1.CacheConfig_builder{
				Ttl: durationpb.New(60 * time.Second),
			}.Build(),
		}.Build()
		cachedTool := NewTool(v1.Tool_builder{Name: proto.String("cached_tool")}.Build(), db, cachedCallDef, nil, "cached_tool_call")
		assert.NotNil(t, cachedTool.GetCacheConfig())
		assert.Equal(t, int64(60), cachedTool.GetCacheConfig().GetTtl().GetSeconds())
	})

	t.Run("policy blocked", func(t *testing.T) {
		policy := configv1.CallPolicy_builder{
			Rules: []*configv1.CallPolicyRule{
				configv1.CallPolicyRule_builder{
					Action: configv1.CallPolicy_DENY.Enum(),
				}.Build(),
			},
			DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
		}.Build()

		blockedTool := NewTool(v1.Tool_builder{Name: proto.String("blocked_tool")}.Build(), db, callDef, []*configv1.CallPolicy{policy}, "blocked_tool_call")

		inputs := map[string]interface{}{
			"age": 20,
		}
		inputsBytes, _ := json.Marshal(inputs)

		req := &tool.ExecutionRequest{
			ToolName:   "blocked_tool",
			ToolInputs: inputsBytes,
		}

		_, err := blockedTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "blocked by policy")
	})
}
