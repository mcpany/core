// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/types/known/durationpb"
)

func ptrTo[T any](v T) *T {
	return &v
}

func TestTool_Methods(t *testing.T) {
	// Setup
	callDef := &configv1.SqlCallDefinition{
		Id:    ptrTo("test-call"),
		Query: ptrTo("SELECT 1"),
		Cache: &configv1.CacheConfig{
			Ttl: durationpb.New(time.Second * 60),
		},
	}
	toolDef := &v1.Tool{
		Name:        ptrTo("test-tool"),
		Description: ptrTo("test description"),
		ServiceId:   ptrTo("test-service"),
	}

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	sqlTool := NewTool(toolDef, db, callDef)

	// Test Tool()
	assert.Equal(t, toolDef, sqlTool.Tool())

	// Test MCPTool()
	mcpTool := sqlTool.MCPTool()
	require.NotNil(t, mcpTool)
	assert.Equal(t, "test-service.test-tool", mcpTool.Name)
	// Second call should return cached value
	assert.Equal(t, mcpTool, sqlTool.MCPTool())

	// Test GetCacheConfig()
	cacheConfig := sqlTool.GetCacheConfig()
	require.NotNil(t, cacheConfig)
	assert.Equal(t, int64(60), cacheConfig.Ttl.Seconds)

	// Test GetCacheConfig with nil callDef
	sqlToolNil := NewTool(toolDef, db, nil)
	assert.Nil(t, sqlToolNil.GetCacheConfig())
}

func TestTool_Execute_Errors(t *testing.T) {
	// Setup
	callDef := &configv1.SqlCallDefinition{
		Id:    ptrTo("test-call"),
		Query: ptrTo("SELECT * FROM non_existent_table"), // Will cause query error
	}
	toolDef := &v1.Tool{
		Name: ptrTo("test-tool"),
	}

	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	sqlTool := NewTool(toolDef, db, callDef)

	// Test Execute - Invalid JSON input
	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(`invalid-json`),
	}
	_, err = sqlTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")

	// Test Execute - Query Error
	req = &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(`{}`),
	}
	_, err = sqlTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute query")
}

func TestUpstream_Register_Errors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	u := NewUpstream()
	defer u.Shutdown(context.Background())

	// Test missing SqlService config
	config := &configv1.UpstreamServiceConfig{
		Id: ptrTo("test-service"),
		// Missing SqlService
	}
	_, _, _, err := u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Equal(t, "sql service config is nil", err.Error())

	// Test Open Error (invalid driver)
	config = &configv1.UpstreamServiceConfig{
		Id: ptrTo("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
			SqlService: &configv1.SqlUpstreamService{
				Driver: ptrTo("invalid-driver"),
				Dsn:    ptrTo(""),
			},
		},
	}
	_, _, _, err = u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open database")

	// Test Tool Manager Error
	config = &configv1.UpstreamServiceConfig{
		Id: ptrTo("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
			SqlService: &configv1.SqlUpstreamService{
				Driver: ptrTo("sqlite"),
				Dsn:    ptrTo(":memory:"),
				Calls: map[string]*configv1.SqlCallDefinition{
					"call1": {
						Id:    ptrTo("call1"),
						Query: ptrTo("SELECT 1"),
					},
				},
			},
		},
	}

	mockToolManager.EXPECT().AddTool(gomock.Any()).Return(assert.AnError)

	_, _, _, err = u.Register(context.Background(), config, mockToolManager, nil, nil, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to add tool")
}
