// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ptrBool(b bool) *bool {
	return &b
}

func ptrString(s string) *string {
	return &s
}

func TestUpstream_Register_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)

	u := NewUpstream()
	defer u.Shutdown(context.Background())

	ctx := context.Background()

	// 1. Nil SQL Config
	config := &configv1.UpstreamServiceConfig{
		Id:            ptrString("test-service"),
		ServiceConfig: nil,
	}
	_, _, _, err := u.Register(ctx, config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sql service config is nil")

	// 2. Invalid Driver (Open fails or Ping fails)
	configInvalidDriver := &configv1.UpstreamServiceConfig{
		Id: ptrString("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
			SqlService: &configv1.SqlUpstreamService{
				Driver: ptrString("non-existent-driver"),
				Dsn:    ptrString("dsn"),
			},
		},
	}
	_, _, _, err = u.Register(ctx, configInvalidDriver, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open database")

	// 5. AddTool Failure
	configAddToolFail := &configv1.UpstreamServiceConfig{
		Id: ptrString("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
			SqlService: &configv1.SqlUpstreamService{
				Driver: ptrString("sqlite"),
				Dsn:    ptrString("file::memory:?cache=shared"),
				Calls: map[string]*configv1.SqlCallDefinition{
					"tool1": {
						Id:    ptrString("tool1"),
						Query: ptrString("SELECT 1"),
					},
				},
			},
		},
	}

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

	callDef := &configv1.SqlCallDefinition{
		Query: ptrString("SELECT * FROM users"),
	}

    // Create Tool
    toolProto := &v1.Tool{
        Name: ptrString("test_tool"),
    }
    toolInstance := NewTool(toolProto, db, callDef)

    ctx := context.Background()

    // 1. Invalid JSON Input
    req := &tool.ExecutionRequest{
        ToolName: "test_tool",
        ToolInputs: json.RawMessage(`{invalid-json}`),
    }
    _, err = toolInstance.Execute(ctx, req)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")

    // 2. Query Failure (Table does not exist)
    reqValid := &tool.ExecutionRequest{
        ToolName: "test_tool",
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

	callDef := &configv1.SqlCallDefinition{
		Query: ptrString("SELECT id, content, nullable FROM data WHERE id = ? OR nullable = ?"),
        ParameterOrder: []string{"id", "missing_param"},
	}

    toolProto := &v1.Tool{
        Name: ptrString("test_tool"),
    }
    toolInstance := NewTool(toolProto, db, callDef)

    ctx := context.Background()

    // 1. Missing Parameter (should pass nil)
    req := &tool.ExecutionRequest{
        ToolName: "test_tool",
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
    callDef := &configv1.SqlCallDefinition{
        Cache: &configv1.CacheConfig{
            IsEnabled: ptrBool(true),
            Ttl: &durationpb.Duration{Seconds: 60},
        },
    }
    tl := NewTool(nil, nil, callDef)
    cc := tl.GetCacheConfig()
    assert.NotNil(t, cc)
    assert.True(t, cc.GetIsEnabled())

    tl2 := NewTool(nil, nil, nil)
    assert.Nil(t, tl2.GetCacheConfig())
}

func TestTool_MCPTool(t *testing.T) {
     toolProto := &v1.Tool{
        Name: ptrString("test_tool"),
        Description: ptrString("desc"),
        ServiceId: ptrString("myservice"),
    }
    tl := NewTool(toolProto, nil, nil)

    mcpTool := tl.MCPTool()
    assert.NotNil(t, mcpTool)
    assert.Equal(t, "myservice.test_tool", mcpTool.Name)

    // Call again to hit Once
    mcpTool2 := tl.MCPTool()
    assert.Equal(t, mcpTool, mcpTool2)
}
