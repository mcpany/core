// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestSQLUpstream_Register_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	u := NewUpstream()
	defer u.Shutdown(context.Background())

	// Case 1: Nil SQL config
	config := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		// Missing SqlService
	}.Build()
	_, _, _, err := u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql service config is nil")

	// Case 2: Invalid Driver
	config = configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("unknown-driver"),
			Dsn:    proto.String("dsn"),
		}.Build(),
	}.Build()
	_, _, _, err = u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	// sql.Open itself might succeed, but Ping should fail or Open might fail for unknown driver
	assert.Contains(t, err.Error(), "failed to open database")

	// Case 3: Invalid Tool Name (Empty)
	config = configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("sqlite"),
			Dsn:    proto.String("file::memory:?cache=shared"),
			Calls: map[string]*configv1.SqlCallDefinition{
				"": configv1.SqlCallDefinition_builder{ // Empty name triggers validation error
					Query: proto.String("SELECT 1"),
				}.Build(),
			},
		}.Build(),
	}.Build()
	// Note: Register iterates map, so order is random if multiple. Here only one.
	_, _, _, err = u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tool name")

	// Case 4: ToolManager AddTool Error
	// We need a valid config that passes until AddTool
	mockToolManager.EXPECT().AddTool(gomock.Any()).Return(assert.AnError)

	config = configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("sqlite"),
			Dsn:    proto.String("file::memory:?cache=shared"),
			Calls: map[string]*configv1.SqlCallDefinition{
				"tool1": configv1.SqlCallDefinition_builder{
					Query: proto.String("SELECT 1"),
				}.Build(),
			},
		}.Build(),
	}.Build()
	_, _, _, err = u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add tool")
}

func TestSQLUpstream_Register_Twice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockToolManager := tool.NewMockManagerInterface(ctrl)

	u := NewUpstream()
	defer u.Shutdown(context.Background())

	config := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("sqlite"),
			Dsn:    proto.String("file::memory:?cache=shared"),
		}.Build(),
	}.Build()

	// First Register
	_, _, _, err := u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	require.NoError(t, err)

	// Second Register (should close previous DB)
	_, _, _, err = u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	require.NoError(t, err)
}

func TestSQLUpstream_Register_PingFailure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockToolManager := tool.NewMockManagerInterface(ctrl)

	u := NewUpstream()
	defer u.Shutdown(context.Background())

	// Use postgres driver with invalid DSN to force Ping failure
	config := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("postgres"),
			// Use short timeout
			Dsn: proto.String("postgres://invalid_host:5432/db?sslmode=disable&connect_timeout=1"),
		}.Build(),
	}.Build()
	_, _, _, err := u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to ping database")
}

func TestSQLUpstream_Execute_Errors(t *testing.T) {
	// Setup SQLite DB manually
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	defer db.Close()

	callDef := configv1.SqlCallDefinition_builder{
		Query: proto.String("SELECT * FROM non_existent_table"),
	}.Build()

	toolInstance := NewTool(v1.Tool_builder{Name: proto.String("test")}.Build(), db, callDef, nil, "test_call")

	// Case 1: Invalid Input JSON
	req := &tool.ExecutionRequest{
		ToolName:   "test",
		ToolInputs: json.RawMessage(`{invalid-json`),
	}
	_, err = toolInstance.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")

	// Case 2: Query Error
	req = &tool.ExecutionRequest{
		ToolName:   "test",
		ToolInputs: json.RawMessage(`{}`),
	}
	_, err = toolInstance.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query") // table not found
}

func TestSQLTool_Execute_EdgeCases(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	defer db.Close()

	// Setup table for edge cases
	_, err = db.Exec("CREATE TABLE edge_cases (id INTEGER, data BLOB)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO edge_cases (id, data) VALUES (1, X'48656c6c6f')") // 'Hello' in hex
	require.NoError(t, err)

	// Case 1: Missing parameter passed as NULL
	// Query: SELECT ? IS NULL as is_null
	callDef := configv1.SqlCallDefinition_builder{
		Query:          proto.String("SELECT ? IS NULL as is_null"),
		ParameterOrder: []string{"missing_param"},
	}.Build()
	toolInstance := NewTool(v1.Tool_builder{Name: proto.String("test_null")}.Build(), db, callDef, nil, "test_null_call")

	req := &tool.ExecutionRequest{
		ToolName:   "test_null",
		ToolInputs: json.RawMessage(`{}`), // missing_param is missing
	}
	result, err := toolInstance.Execute(context.Background(), req)
	require.NoError(t, err)
	resSlice := result.([]map[string]any)
	require.Len(t, resSlice, 1)
	// in sqlite boolean is 0 or 1.
	assert.EqualValues(t, 1, resSlice[0]["is_null"])

	// Case 2: BLOB handling
	callDefBlob := configv1.SqlCallDefinition_builder{
		Query: proto.String("SELECT data FROM edge_cases WHERE id = 1"),
	}.Build()
	toolInstanceBlob := NewTool(v1.Tool_builder{Name: proto.String("test_blob")}.Build(), db, callDefBlob, nil, "test_blob_call")
	reqBlob := &tool.ExecutionRequest{
		ToolName:   "test_blob",
		ToolInputs: json.RawMessage(`{}`),
	}
	resultBlob, err := toolInstanceBlob.Execute(context.Background(), reqBlob)
	require.NoError(t, err)
	resSliceBlob := resultBlob.([]map[string]any)
	require.Len(t, resSliceBlob, 1)
	// Should be string "Hello"
	assert.Equal(t, "Hello", resSliceBlob[0]["data"])
}

func TestSQLTool_Methods(t *testing.T) {
	db, err := sql.Open("sqlite", "file::memory:?cache=shared")
	require.NoError(t, err)
	defer db.Close()

	callDef := configv1.SqlCallDefinition_builder{Query: proto.String("SELECT 1")}.Build()
	// Provide ServiceId to ensure correct naming in MCPTool
	toolInstance := NewTool(v1.Tool_builder{
		Name:      proto.String("test_tool"),
		ServiceId: proto.String("service"),
	}.Build(), db, callDef, nil, "test_tool_call")

	assert.NotNil(t, toolInstance.Tool())
	assert.Equal(t, "test_tool", toolInstance.Tool().GetName())

	assert.NotNil(t, toolInstance.MCPTool())
	assert.Equal(t, "service.test_tool", toolInstance.MCPTool().Name)

	// Call again to test Once
	assert.NotNil(t, toolInstance.MCPTool())
}

func TestSQLTool_GetCacheConfig(t *testing.T) {
	t.Run("Returns Config", func(t *testing.T) {
		callDef := configv1.SqlCallDefinition_builder{
			Cache: configv1.CacheConfig_builder{
				Ttl: durationpb.New(time.Second * 60),
			}.Build(),
		}.Build()
		toolInstance := NewTool(nil, nil, callDef, nil, "")
		assert.NotNil(t, toolInstance.GetCacheConfig())
		assert.Equal(t, int64(60), toolInstance.GetCacheConfig().GetTtl().GetSeconds())
	})

	t.Run("Returns Nil when CallDef is Nil", func(t *testing.T) {
		toolInstance := NewTool(nil, nil, nil, nil, "")
		assert.Nil(t, toolInstance.GetCacheConfig())
	})
}
