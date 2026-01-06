// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcprouterv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

// Mock Storage interface
type MockStorage struct {
	SaveServiceFunc   func(ctx context.Context, service *configv1.UpstreamServiceConfig) error
	DeleteServiceFunc func(ctx context.Context, name string) error
}

func (m *MockStorage) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) { return nil, nil }
func (m *MockStorage) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	if m.SaveServiceFunc != nil {
		return m.SaveServiceFunc(ctx, service)
	}
	return nil
}
func (m *MockStorage) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	return nil, nil
}
func (m *MockStorage) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return nil, nil
}
func (m *MockStorage) DeleteService(ctx context.Context, name string) error {
	if m.DeleteServiceFunc != nil {
		return m.DeleteServiceFunc(ctx, name)
	}
	return nil
}
func (m *MockStorage) GetGlobalSettings() (*configv1.GlobalSettings, error)            { return nil, nil }
func (m *MockStorage) SaveGlobalSettings(settings *configv1.GlobalSettings) error      { return nil }
func (m *MockStorage) ListSecrets() ([]*configv1.Secret, error)                        { return nil, nil }
func (m *MockStorage) GetSecret(id string) (*configv1.Secret, error)                   { return nil, nil }
func (m *MockStorage) SaveSecret(secret *configv1.Secret) error                        { return nil }
func (m *MockStorage) DeleteSecret(id string) error                                    { return nil }
func (m *MockStorage) Close() error                                                    { return nil }

// Mock ServiceRegistry interface (since we don't have gomock generated for it yet, or avoiding import cycles)
type MockServiceRegistry struct {
	RegisterServiceFunc   func(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error)
	UnregisterServiceFunc func(ctx context.Context, serviceName string) error
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	if m.RegisterServiceFunc != nil {
		return m.RegisterServiceFunc(ctx, serviceConfig)
	}
	return "test-id", nil, nil, nil
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	if m.UnregisterServiceFunc != nil {
		return m.UnregisterServiceFunc(ctx, serviceName)
	}
	return nil
}

func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	return nil, nil
}
func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	return nil, false
}
func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	return "", false
}

func TestNewServer(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager, nil, nil)
	assert.NotNil(t, server)
}

func TestClearCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)

	// We use a real CachingMiddleware here since it is a struct and cannot be mocked easily.
	// The default implementation uses an in-memory cache which should succeed.
	cache := middleware.NewCachingMiddleware(mockManager)
	server := NewServer(cache, mockManager, nil, nil)

	req := &pb.ClearCacheRequest{}
	resp, err := server.ClearCache(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestClearCache_NilCache(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	server := NewServer(nil, mockManager, nil, nil)

	req := &pb.ClearCacheRequest{}
	_, err := server.ClearCache(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "caching is not enabled")
}

func TestListServices(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	server := NewServer(nil, mockManager, nil, nil)

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
	server := NewServer(nil, mockManager, nil, nil)

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
	server := NewServer(nil, mockManager, nil, nil)

	mockManager.EXPECT().GetServiceInfo("unknown").Return(nil, false)

	_, err := server.GetService(context.Background(), &pb.GetServiceRequest{ServiceId: proto.String("unknown")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestListTools(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	server := NewServer(nil, mockManager, nil, nil)

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
	server := NewServer(nil, mockManager, nil, nil)

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
	server := NewServer(nil, mockManager, nil, nil)

	mockManager.EXPECT().GetTool("unknown").Return(nil, false)

	_, err := server.GetTool(context.Background(), &pb.GetToolRequest{ToolName: proto.String("unknown")})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRegisterService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	mockRegistry := &MockServiceRegistry{
		RegisterServiceFunc: func(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
			assert.Equal(t, "test-service", serviceConfig.GetName())
			return "service-id-123", nil, nil, nil
		},
	}
	mockStorage := &MockStorage{
		SaveServiceFunc: func(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
			assert.Equal(t, "service-id-123", service.GetId())
			return nil
		},
	}

	server := NewServer(nil, mockManager, mockRegistry, mockStorage)

	req := &pb.RegisterServiceRequest{
		ServiceConfig: &configv1.UpstreamServiceConfig{
			Name: proto.String("test-service"),
		},
	}

	resp, err := server.RegisterService(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "service-id-123", resp.GetServiceId())
}

func TestUnregisterService(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockManager := tool.NewMockManagerInterface(ctrl)
	mockRegistry := &MockServiceRegistry{
		UnregisterServiceFunc: func(ctx context.Context, serviceName string) error {
			assert.Equal(t, "test-service", serviceName)
			return nil
		},
	}
	mockStorage := &MockStorage{
		DeleteServiceFunc: func(ctx context.Context, name string) error {
			assert.Equal(t, "test-service", name)
			return nil
		},
	}

	server := NewServer(nil, mockManager, mockRegistry, mockStorage)

	req := &pb.UnregisterServiceRequest{
		ServiceName: proto.String("test-service"),
	}

	resp, err := server.UnregisterService(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
}
