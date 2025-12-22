// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package sql

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func ptrTest(s string) *string {
	return &s
}

func TestSQLUpstream_Register_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)

	// Setup SQLite DB in memory
	u := NewUpstream()
	defer u.Shutdown(context.Background())

	// Config
	config := &configv1.UpstreamServiceConfig{
		Id: ptrTest("test-sql-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
			SqlService: &configv1.SqlUpstreamService{
				Driver: ptrTest("sqlite"),
				Dsn:    ptrTest("file::memory:?cache=shared"),
				Calls: map[string]*configv1.SqlCallDefinition{
					"get_users": {
						Id:             ptrTest("get_users"),
						Query:          ptrTest("SELECT id, name FROM users WHERE active = ?"),
						ParameterOrder: []string{"active"},
					},
				},
			},
		},
	}

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

	// Coverage for getters
	assert.NotNil(t, capturedTool.Tool())
	assert.NotNil(t, capturedTool.MCPTool())
	assert.Nil(t, capturedTool.GetCacheConfig())

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
