// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestMockManagerInterface_Coverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockManagerInterface(ctrl)

	// Test expectation setting for all methods
	m.EXPECT().AddMiddleware(gomock.Any()).AnyTimes()
	m.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any()).AnyTimes()
	m.EXPECT().AddTool(gomock.Any()).Return(nil).AnyTimes()
	m.EXPECT().ClearToolsForService(gomock.Any()).AnyTimes()
	m.EXPECT().ExecuteTool(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	m.EXPECT().GetAllowedServiceIDs(gomock.Any()).Return(nil, false).AnyTimes()
	m.EXPECT().GetServiceInfo(gomock.Any()).Return(nil, false).AnyTimes()
	m.EXPECT().GetTool(gomock.Any()).Return(nil, false).AnyTimes()
	m.EXPECT().IsServiceAllowed(gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	m.EXPECT().ListServices().Return(nil).AnyTimes()
	m.EXPECT().ListTools().Return(nil).AnyTimes()
	m.EXPECT().SetMCPServer(gomock.Any()).AnyTimes()
	m.EXPECT().SetProfiles(gomock.Any(), gomock.Any()).AnyTimes()
	m.EXPECT().ToolMatchesProfile(gomock.Any(), gomock.Any()).Return(true).AnyTimes()

	// Call methods
	m.AddMiddleware(nil)
	m.AddServiceInfo("id", nil)
	_ = m.AddTool(nil)
	m.ClearToolsForService("id")
	_, _ = m.ExecuteTool(nil, nil)
	_, _ = m.GetAllowedServiceIDs("profile")
	_, _ = m.GetServiceInfo("id")
	_, _ = m.GetTool("name")
	_ = m.IsServiceAllowed("s", "p")
	_ = m.ListServices()
	_ = m.ListTools()
	m.SetMCPServer(nil)
	m.SetProfiles(nil, nil)
	_ = m.ToolMatchesProfile(nil, "p")
}

func TestMockTool_Coverage(t *testing.T) {
	mt := &MockTool{}

	// Default behaviors
	assert.NotNil(t, mt.Tool())
	assert.Nil(t, mt.MCPTool())
	_, _ = mt.Execute(context.Background(), nil)
	assert.Nil(t, mt.GetCacheConfig())

	// Custom behaviors
	mt.ToolFunc = func() *routerv1.Tool { return &routerv1.Tool{} }
	assert.NotNil(t, mt.Tool())

	mt.MCPToolFunc = func() *mcp.Tool { return &mcp.Tool{} }
	assert.NotNil(t, mt.MCPTool())

	mt.ExecuteFunc = func(ctx context.Context, req *ExecutionRequest) (any, error) { return "ok", nil }
	res, _ := mt.Execute(context.Background(), nil)
	assert.Equal(t, "ok", res)

	mt.GetCacheConfigFunc = func() *configv1.CacheConfig { return &configv1.CacheConfig{} }
	assert.NotNil(t, mt.GetCacheConfig())
}
