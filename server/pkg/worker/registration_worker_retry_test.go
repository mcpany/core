package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockServiceRegistry is a mock implementation of ServiceRegistryInterface
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
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	return args.Get(0).(*configv1.UpstreamServiceConfig), args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

func TestServiceRegistrationWorker_RetryLogic_Backoff(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	mockRegistry := new(MockServiceRegistry)
	worker := NewServiceRegistrationWorker(busProvider, mockRegistry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("retry-service"),
	}

	// Simulate persistent failure
	mockRegistry.On("RegisterService", mock.Anything, serviceConfig).Return("", ([]*configv1.ToolDefinition)(nil), ([]*configv1.ResourceDefinition)(nil), errors.New("persistent failure"))

	// Start worker
	worker.Start(ctx)

	// Publish request
	req := &bus.ServiceRegistrationRequest{Config: serviceConfig}
	reqBus, _ := bus.GetBus[*bus.ServiceRegistrationRequest](busProvider, bus.ServiceRegistrationRequestTopic)
	err := reqBus.Publish(ctx, "request", req)
	assert.NoError(t, err)

	// Wait enough time for 2 retries.
    // 1st retry: 1s
    // 2nd retry: 2s
    // Total wait > 3s
	time.Sleep(3500 * time.Millisecond)

	// Verify calls
	// Expect initial + 2 retries (since 2nd retry is after 3s total)
    // Actually, check if it's AT LEAST 2 retries (3 calls total).
    // Initial call (t=0)
    // Retry 1 (t=1s)
    // Retry 2 (t=1s + 2s = 3s)
	mockRegistry.AssertNumberOfCalls(t, "RegisterService", 3)
}

func TestServiceRegistrationWorker_RetryLogic_MaxRetries(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	mockRegistry := new(MockServiceRegistry)
	worker := NewServiceRegistrationWorker(busProvider, mockRegistry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("max-retry-service"),
	}

	// Simulate persistent failure
	mockRegistry.On("RegisterService", mock.Anything, serviceConfig).Return("", ([]*configv1.ToolDefinition)(nil), ([]*configv1.ResourceDefinition)(nil), errors.New("persistent failure"))

	// Start worker
	worker.Start(ctx)

	// Publish request with RetryCount = 4 (MaxRetries = 5)
    // So it should try once (count=4), fail, retry (count=5), fail, stop.
    // Total calls = 2.

	req := &bus.ServiceRegistrationRequest{
        Config: serviceConfig,
        RetryCount: 4,
    }
	reqBus, _ := bus.GetBus[*bus.ServiceRegistrationRequest](busProvider, bus.ServiceRegistrationRequestTopic)
	err := reqBus.Publish(ctx, "request", req)
	assert.NoError(t, err)

    // Wait enough for the 1 retry (delay = 1 * 2^4 = 16s).
    // But wait, the first call happens immediately.
    // The retry is scheduled for 16s later.
    // We can't wait 16s in unit test.

    // So we verify that it was called ONCE immediately.
    time.Sleep(100 * time.Millisecond)
	mockRegistry.AssertNumberOfCalls(t, "RegisterService", 1)

    // We can't verify the retry happens without waiting 16s or mocking time.
}

func TestServiceRegistrationWorker_RetryLogic_StopAfterMax(t *testing.T) {
	// Setup
	busProvider, _ := bus.NewProvider(nil)
	mockRegistry := new(MockServiceRegistry)
	worker := NewServiceRegistrationWorker(busProvider, mockRegistry)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serviceConfig := &configv1.UpstreamServiceConfig{
		Name: proto.String("stop-retry-service"),
	}

	// Simulate persistent failure
	mockRegistry.On("RegisterService", mock.Anything, serviceConfig).Return("", ([]*configv1.ToolDefinition)(nil), ([]*configv1.ResourceDefinition)(nil), errors.New("persistent failure"))

	// Start worker
	worker.Start(ctx)

    // RetryCount = 5. MaxRetries = 5.
    // Should run once, fail. Check 5 < 5 -> False. Stop.
    // Total calls = 1.

    req5 := &bus.ServiceRegistrationRequest{
        Config: serviceConfig,
        RetryCount: 5,
    }

	reqBus, _ := bus.GetBus[*bus.ServiceRegistrationRequest](busProvider, bus.ServiceRegistrationRequestTopic)
    err := reqBus.Publish(ctx, "request", req5)
    assert.NoError(t, err)

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Verify calls
    // Should be called once.
	mockRegistry.AssertNumberOfCalls(t, "RegisterService", 1)
}
