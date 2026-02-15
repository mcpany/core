package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	_ "modernc.org/sqlite"
)

func TestSQLUpstream_Register_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)

	// Setup SQLite DB in memory
	u := NewUpstream()
	defer u.Shutdown(context.Background())

	// Config
	config := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-sql-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("sqlite"),
			Dsn:    proto.String("file::memory:?cache=shared"),
			Calls: map[string]*configv1.SqlCallDefinition{
				"get_users": configv1.SqlCallDefinition_builder{
					Id:             proto.String("get_users"),
					Query:          proto.String("SELECT id, name FROM users WHERE active = ?"),
					ParameterOrder: []string{"active"},
				}.Build(),
			},
		}.Build(),
	}.Build()

	// Expect AddTool to be called
	var capturedTool tool.Tool
	mockToolManager.EXPECT().AddTool(gomock.Any()).DoAndReturn(func(t tool.Tool) error {
		capturedTool = t
		return nil
	})

	// Register
	id, toolDefs, _, err := u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	require.NoError(t, err)
	assert.Equal(t, "test-sql-service", id)
	assert.Len(t, toolDefs, 1)

	// Setup DB schema and data
	// Access DB directly to setup
	// Since Register opens a new DB connection, we can assume u.db is set.
	require.NotNil(t, u.db)
	_, err = u.db.Exec("CREATE TABLE users (id INTEGER, name TEXT, active BOOLEAN)")
	require.NoError(t, err)
	_, err = u.db.Exec("INSERT INTO users (id, name, active) VALUES (1, 'Alice', true), (2, 'Bob', false)")
	require.NoError(t, err)

	// Test Tool Execution
	require.NotNil(t, capturedTool)

	// Test Case 1: Active users
	req := &tool.ExecutionRequest{
		ToolName:   "get_users",
		ToolInputs: json.RawMessage(`{"active": true}`),
	}
	result, err := capturedTool.Execute(context.Background(), req)
	require.NoError(t, err)

	// Result should be []map[string]any
	resSlice, ok := result.([]map[string]any)
	require.True(t, ok)
	require.Len(t, resSlice, 1)
	assert.EqualValues(t, 1, resSlice[0]["id"])
	assert.Equal(t, "Alice", resSlice[0]["name"])

	// Test Case 2: Inactive users
	req = &tool.ExecutionRequest{
		ToolName:   "get_users",
		ToolInputs: json.RawMessage(`{"active": false}`),
	}
	result, err = capturedTool.Execute(context.Background(), req)
	require.NoError(t, err)
	resSlice, ok = result.([]map[string]any)
	require.True(t, ok)
	require.Len(t, resSlice, 1)
	assert.EqualValues(t, 2, resSlice[0]["id"])
	assert.Equal(t, "Bob", resSlice[0]["name"])
}

func TestUpstream_Register_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)

	u := NewUpstream()
	defer u.Shutdown(context.Background())

	ctx := context.Background()

	// 1. Nil SQL Config
	config := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
	}.Build()
	_, _, _, err := u.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql service config is nil")

	// 2. Invalid Driver (Open fails or Ping fails)
	configInvalidDriver := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("non-existent-driver"),
			Dsn:    proto.String("dsn"),
		}.Build(),
	}.Build()
	_, _, _, err = u.Register(ctx, configInvalidDriver, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open database")

	// 5. AddTool Failure
	configAddToolFail := configv1.UpstreamServiceConfig_builder{
		Id: proto.String("test-service"),
		SqlService: configv1.SqlUpstreamService_builder{
			Driver: proto.String("sqlite"),
			Dsn:    proto.String("file::memory:?cache=shared"),
			Calls: map[string]*configv1.SqlCallDefinition{
				"tool1": configv1.SqlCallDefinition_builder{
					Id:    proto.String("tool1"),
					Query: proto.String("SELECT 1"),
				}.Build(),
			},
		}.Build(),
	}.Build()

	mockToolManager.EXPECT().AddTool(gomock.Any()).Return(errors.New("add tool failed"))

	_, _, _, err = u.Register(ctx, configAddToolFail, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add tool")
}

func TestTool_Execute_Errors(t *testing.T) {
	// Setup DB
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	callDef := configv1.SqlCallDefinition_builder{
		Query: proto.String("SELECT * FROM users"),
	}.Build()

	// Create Tool
	toolProto := v1.Tool_builder{
		Name: proto.String("test_tool"),
	}.Build()
	toolInstance := NewTool(toolProto, db, callDef, nil, "test_tool_call")

	ctx := context.Background()

	// 1. Invalid JSON Input
	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: json.RawMessage(`{invalid-json}`),
	}
	_, err = toolInstance.Execute(ctx, req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")

	// 2. Query Failure (Table does not exist)
	reqValid := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: json.RawMessage(`{}`),
	}
	_, err = toolInstance.Execute(ctx, reqValid)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query")
}

func TestTool_Execute_EdgeCases(t *testing.T) {
	// Setup DB
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Exec("CREATE TABLE data (id INTEGER, content BLOB, nullable TEXT)")
	require.NoError(t, err)
	_, err = db.Exec("INSERT INTO data (id, content, nullable) VALUES (1, 'blobdata', NULL)")
	require.NoError(t, err)

	callDef := configv1.SqlCallDefinition_builder{
		Query:          proto.String("SELECT id, content, nullable FROM data WHERE id = ? OR nullable = ?"),
		ParameterOrder: []string{"id", "missing_param"},
	}.Build()

	toolProto := v1.Tool_builder{
		Name: proto.String("test_tool"),
	}.Build()
	toolInstance := NewTool(toolProto, db, callDef, nil, "test_tool_call")

	ctx := context.Background()

	// 1. Missing Parameter (should pass nil)
	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: json.RawMessage(`{"id": 1}`),
	}

	res, err := toolInstance.Execute(ctx, req)
	require.NoError(t, err)

	resSlice, ok := res.([]map[string]any)
	require.True(t, ok)
	require.Len(t, resSlice, 1)

	// Check Blob handling (should be string)
	assert.Equal(t, "blobdata", resSlice[0]["content"])

	// Check NULL handling
	assert.Nil(t, resSlice[0]["nullable"])
}

func TestUpstream_Shutdown(t *testing.T) {
	u := NewUpstream()
	// db is nil initially
	err := u.Shutdown(context.Background())
	assert.NoError(t, err)

	// Set a db (mock or real)
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	u.db = db

	err = u.Shutdown(context.Background())
	assert.NoError(t, err)

	// Verify db is closed
	err = u.db.Ping()
	assert.Error(t, err) // closed
}

func TestTool_GetCacheConfig(t *testing.T) {
	callDef := configv1.SqlCallDefinition_builder{
		Cache: configv1.CacheConfig_builder{
			IsEnabled: proto.Bool(true),
			Ttl:       durationpb.New(60 * 1000 * 1000 * 1000), // 60s
		}.Build(),
	}.Build()
	tl := NewTool(nil, nil, callDef, nil, "")
	cc := tl.GetCacheConfig()
	assert.NotNil(t, cc)
	assert.True(t, cc.GetIsEnabled())

	tl2 := NewTool(nil, nil, nil, nil, "")
	assert.Nil(t, tl2.GetCacheConfig())
}

func TestTool_MCPTool(t *testing.T) {
	toolProto := v1.Tool_builder{
		Name:        proto.String("test_tool"),
		Description: proto.String("desc"),
		ServiceId:   proto.String("myservice"),
	}.Build()
	tl := NewTool(toolProto, nil, nil, nil, "")

	mcpTool := tl.MCPTool()
	assert.NotNil(t, mcpTool)
	assert.Equal(t, "myservice.test_tool", mcpTool.Name)

	// Call again to hit Once
	mcpTool2 := tl.MCPTool()
	assert.Equal(t, mcpTool, mcpTool2)
}
