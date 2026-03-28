// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestMockToolCoverage(t *testing.T) {
	mt := &MockTool{}

	// Tool()
	assert.NotNil(t, mt.Tool())

	mt.ToolFunc = func() *v1.Tool { return v1.Tool_builder{Name: proto.String("mock")}.Build() }
	assert.Equal(t, "mock", mt.Tool().GetName())

	// MCPTool()
	assert.Nil(t, mt.MCPTool())

	// Execute()
	res, err := mt.Execute(context.Background(), nil)
	assert.Nil(t, res)
	assert.NoError(t, err)

	mt.ExecuteFunc = func(ctx context.Context, req *ExecutionRequest) (any, error) {
		return "ok", nil
	}
	res, err = mt.Execute(context.Background(), nil)
	assert.Equal(t, "ok", res)
	assert.NoError(t, err)

	// GetCacheConfig
	assert.Nil(t, mt.GetCacheConfig())
}

func TestMockManagerInterfaceCoverage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockManagerInterface(ctrl)

	m.EXPECT().AddTool(gomock.Any()).Return(nil)
	_ = m.AddTool(nil)

	m.EXPECT().GetTool(gomock.Any()).Return(nil, false)
	_, _ = m.GetTool("t")

	m.EXPECT().ListTools().Return(nil)
	_ = m.ListTools()

	m.EXPECT().ClearToolsForService(gomock.Any())
	m.ClearToolsForService("s")

	m.EXPECT().ExecuteTool(gomock.Any(), gomock.Any()).Return(nil, nil)
	_, _ = m.ExecuteTool(context.Background(), nil)

	m.EXPECT().SetMCPServer(gomock.Any())
	m.SetMCPServer(nil)

	m.EXPECT().AddMiddleware(gomock.Any())
	m.AddMiddleware(nil)

	m.EXPECT().AddServiceInfo(gomock.Any(), gomock.Any())
	m.AddServiceInfo("s", nil)

	m.EXPECT().GetServiceInfo(gomock.Any()).Return(nil, false)
	_, _ = m.GetServiceInfo("s")

	m.EXPECT().ListServices().Return(nil)
	_ = m.ListServices()

	m.EXPECT().SetProfiles(gomock.Any(), gomock.Any())
	m.SetProfiles(nil, nil)

	m.EXPECT().IsServiceAllowed(gomock.Any(), gomock.Any()).Return(false)
	m.IsServiceAllowed("s", "p")

	m.EXPECT().ToolMatchesProfile(gomock.Any(), gomock.Any()).Return(false)
	m.ToolMatchesProfile(nil, "p")

	m.EXPECT().GetAllowedServiceIDs(gomock.Any()).Return(nil, false)
	_, _ = m.GetAllowedServiceIDs("p")
}
