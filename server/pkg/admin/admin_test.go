package admin

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/tool"
	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockServiceRegistry is a mock for ServiceRegistryInterface
type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*configv1.UpstreamServiceConfig), args.Bool(1)
}

func TestServer_CreateService(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	s := NewServer(nil, nil, mockRegistry)

	ctx := context.Background()
	svcConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
	}

	mockRegistry.On("RegisterService", ctx, svcConfig).Return("test-service", []*configv1.ToolDefinition{}, []*configv1.ResourceDefinition{}, nil)

	req := &pb.CreateServiceRequest{
		Service: svcConfig,
	}

	resp, err := s.CreateService(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, "test-service", resp.GetServiceId())
	mockRegistry.AssertExpectations(t)
}

func TestServer_DeleteService(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	s := NewServer(nil, nil, mockRegistry)

	ctx := context.Background()
	serviceID := "test-service"

	mockRegistry.On("UnregisterService", ctx, serviceID).Return(nil)

	req := &pb.DeleteServiceRequest{
		ServiceId: &serviceID,
	}

	resp, err := s.DeleteService(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	mockRegistry.AssertExpectations(t)
}

func TestServer_CreateService_MissingConfig(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	s := NewServer(nil, nil, mockRegistry)
	req := &pb.CreateServiceRequest{}
	_, err := s.CreateService(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service config is required")
}

func TestServer_DeleteService_MissingID(t *testing.T) {
	s := NewServer(nil, nil, new(MockServiceRegistry))
	req := &pb.DeleteServiceRequest{}
	_, err := s.DeleteService(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "service ID is required")
}

func TestServer_CreateService_RegistryError(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	s := NewServer(nil, nil, mockRegistry)

	ctx := context.Background()
	svcConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
	}

	mockRegistry.On("RegisterService", ctx, svcConfig).Return("", ([]*configv1.ToolDefinition)(nil), ([]*configv1.ResourceDefinition)(nil), assert.AnError)

	req := &pb.CreateServiceRequest{
		Service: svcConfig,
	}

	_, err := s.CreateService(ctx, req)
	assert.Error(t, err)
	mockRegistry.AssertExpectations(t)
}
