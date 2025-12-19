// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/middleware"
	"github.com/mcpany/core/pkg/tool"
	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager)
	assert.NotNil(t, server)
}

func TestClearCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)

	// We use a real CachingMiddleware here since it is a struct and cannot be mocked easily.
	// The default implementation uses an in-memory cache which should succeed.
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager)

	req := &pb.ClearCacheRequest{}
	resp, err := server.ClearCache(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestListServices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager)

	mockServices := []*tool.ServiceInfo{
		{
			Name: "service1",
			Config: &configv1.UpstreamServiceConfig{
				Id:            proto.String("hash1"),
				SanitizedName: proto.String("id1"),
			},
		},
		{
			Name: "service2",
			Config: &configv1.UpstreamServiceConfig{
				Id:            proto.String("hash2"),
				SanitizedName: proto.String("id2"),
			},
		},
	}

	mockManager.EXPECT().ListServices().Return(mockServices)

	req := &pb.ListServicesRequest{}
	resp, err := server.ListServices(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Len(t, resp.Services, 2)
	assert.Equal(t, "service1", resp.Services[0].GetName())
	assert.Equal(t, "id1", resp.Services[0].GetId())
}

func TestGetService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager)

	serviceID := "test-id"
	mockService := &tool.ServiceInfo{
		Name: "test-service",
		Config: &configv1.UpstreamServiceConfig{
			Id:   proto.String(serviceID),
			Name: proto.String("test-service"),
		},
	}

	mockManager.EXPECT().GetServiceInfo(serviceID).Return(mockService, true)

	req := &pb.GetServiceRequest{Id: proto.String(serviceID)}
	resp, err := server.GetService(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-service", resp.Config.GetName())

	// Test Not Found
	mockManager.EXPECT().GetServiceInfo("non-existent").Return(nil, false)
	_, err = server.GetService(context.Background(), &pb.GetServiceRequest{Id: proto.String("non-existent")})
	assert.Error(t, err)
}

func TestListTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager)

	mockTool1 := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{Name: proto.String("tool1"), ServiceId: proto.String("svc1")}
		},
	}

	mockTool2 := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{Name: proto.String("tool2"), ServiceId: proto.String("svc2")}
		},
	}

	mockManager.EXPECT().ListTools().Return([]tool.Tool{mockTool1, mockTool2})

	req := &pb.ListToolsRequest{}
	resp, err := server.ListTools(context.Background(), req)

	assert.NoError(t, err)
	assert.Len(t, resp.Tools, 2)

	// Test Filtering
	mockManager.EXPECT().ListTools().Return([]tool.Tool{mockTool1, mockTool2})
	reqFilter := &pb.ListToolsRequest{ServiceId: proto.String("svc1")}
	respFilter, err := server.ListTools(context.Background(), reqFilter)
	assert.NoError(t, err)
	assert.Len(t, respFilter.Tools, 1)
	assert.Equal(t, "tool1", respFilter.Tools[0].GetName())
}

func TestGetTool(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager)

	toolName := "test-tool"
	mockTool := &tool.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{Name: proto.String(toolName)}
		},
	}

	mockManager.EXPECT().GetTool(toolName).Return(mockTool, true)

	req := &pb.GetToolRequest{Name: proto.String(toolName)}
	resp, err := server.GetTool(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, toolName, resp.Tool.GetName())

	// Test Not Found
	mockManager.EXPECT().GetTool("non-existent").Return(nil, false)
	_, err = server.GetTool(context.Background(), &pb.GetToolRequest{Name: proto.String("non-existent")})
	assert.Error(t, err)
}
