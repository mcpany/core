// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	store := memory.NewStore()
	server := NewServer(cache, mockManager, store)
	assert.NotNil(t, server)
}

func TestClearCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)

	// We use a real CachingMiddleware here since it is a struct and cannot be mocked easily.
	// The default implementation uses an in-memory cache which should succeed.
	cache := middleware.NewCachingMiddleware(mockManager)
	store := memory.NewStore()
	server := NewServer(cache, mockManager, store)

	req := &pb.ClearCacheRequest{}
	resp, err := server.ClearCache(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClearCache_NilCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	server := NewServer(nil, mockManager, store)

	req := &pb.ClearCacheRequest{}
	_, err := server.ClearCache(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "caching is not enabled")
}

func TestListServices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	server := NewServer(nil, mockManager, store)

	expectedServices := []*tool.ServiceInfo{
		{
			Config: &configv1.UpstreamServiceConfig{
				Name: proto.String("test-service"),
			},
		},
	}
	mockManager.EXPECT().ListServices().Return(expectedServices)

	resp, err := server.ListServices(context.Background(), &pb.ListServicesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Services, 1)
	assert.Equal(t, "test-service", resp.Services[0].GetName())
}

func TestGetService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	server := NewServer(nil, mockManager, store)

	serviceInfo := &tool.ServiceInfo{
		Config: &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
		},
	}
	mockManager.EXPECT().GetServiceInfo("test-service-id").Return(serviceInfo, true)

	resp, err := server.GetService(context.Background(), &pb.GetServiceRequest{ServiceId: proto.String("test-service-id")})
	assert.NoError(t, err)
	assert.Equal(t, "test-service", resp.Service.GetName())
}

func TestGetService_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	server := NewServer(nil, mockManager, store)

	mockManager.EXPECT().GetServiceInfo("unknown").Return(nil, false)

	_, err := server.GetService(context.Background(), &pb.GetServiceRequest{ServiceId: proto.String("unknown")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	server := NewServer(nil, mockManager, store)

	mockTool := &tool.MockTool{
		ToolFunc: func() *mcprouterv1.Tool {
			return &mcprouterv1.Tool{Name: proto.String("test-tool")}
		},
	}

	mockManager.EXPECT().ListTools().Return([]tool.Tool{mockTool})

	resp, err := server.ListTools(context.Background(), &pb.ListToolsRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.Tools, 1)
	assert.Equal(t, "test-tool", resp.Tools[0].GetName())
}

func TestGetTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	server := NewServer(nil, mockManager, store)

	mockTool := &tool.MockTool{
		ToolFunc: func() *mcprouterv1.Tool {
			return &mcprouterv1.Tool{Name: proto.String("test-tool")}
		},
	}

	mockManager.EXPECT().GetTool("test-tool").Return(mockTool, true)

	resp, err := server.GetTool(context.Background(), &pb.GetToolRequest{ToolName: proto.String("test-tool")})
	assert.NoError(t, err)
	assert.Equal(t, "test-tool", resp.Tool.GetName())
}

func TestGetTool_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	store := memory.NewStore()
	server := NewServer(nil, mockManager, store)

	mockManager.EXPECT().GetTool("unknown").Return(nil, false)

	_, err := server.GetTool(context.Background(), &pb.GetToolRequest{ToolName: proto.String("unknown")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
