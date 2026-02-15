// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// ReconcileMockServiceRegistry is a mock for ServiceRegistryInterface
type ReconcileMockServiceRegistry struct {
	mock.Mock
}

func (m *ReconcileMockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	var tools []*configv1.ToolDefinition
	if args.Get(1) != nil {
		tools = args.Get(1).([]*configv1.ToolDefinition)
	}
	var resources []*configv1.ResourceDefinition
	if args.Get(2) != nil {
		resources = args.Get(2).([]*configv1.ResourceDefinition)
	}
	return args.String(0), tools, resources, args.Error(3)
}

func (m *ReconcileMockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *ReconcileMockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *ReconcileMockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*tool.ServiceInfo), args.Bool(1)
}

// Additional methods required by interface but not used in reconcileServices directly
func (m *ReconcileMockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	return nil, false
}
func (m *ReconcileMockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	return "", false
}

// MockBus is a mock implementation of the bus for testing Publish
type MockBus struct {
	PublishedMessages []*bus.ServiceRegistrationRequest
	PublishError      error
}

func (m *MockBus) Publish(ctx context.Context, topic string, msg *bus.ServiceRegistrationRequest) error {
	if m.PublishError != nil {
		return m.PublishError
	}
	m.PublishedMessages = append(m.PublishedMessages, msg)
	return nil
}

func (m *MockBus) Subscribe(ctx context.Context, topic string, handler func(*bus.ServiceRegistrationRequest)) func() {
	return func() {}
}

func (m *MockBus) SubscribeOnce(ctx context.Context, topic string, handler func(*bus.ServiceRegistrationRequest)) func() {
	return func() {}
}

func TestReconcileServices_Sync_AddRemoveUpdate(t *testing.T) {
	mockRegistry := new(ReconcileMockServiceRegistry)
	app := NewApplication()
	app.ServiceRegistry = mockRegistry
	app.busProvider = nil // Force sync path

	// Initial State: Service A (v1)
	serviceAv1 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-a"),
		Id:   proto.String("service-a"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://v1"),
		}.Build(),
	}.Build()

	// Service C (to be removed)
	serviceC := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-c"),
		Id:   proto.String("service-c"),
	}.Build()

	// New Config: Service A (v2), Service B (new)
	serviceAv2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-a"),
		Id:   proto.String("service-a"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://v2"),
		}.Build(),
	}.Build()

	serviceB := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-b"),
		Id:   proto.String("service-b"),
	}.Build()

	newConfig := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{serviceAv2, serviceB},
		GlobalSettings: configv1.GlobalSettings_builder{
			AutoDiscoverLocal: proto.Bool(false),
		}.Build(),
	}.Build()

	// Expectations
	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{serviceAv1, serviceC}, nil)

	// Service C should be removed
	mockRegistry.On("UnregisterService", mock.Anything, "service-c").Return(nil)

	// Service A should be updated (Unregister old, Register new)
	mockRegistry.On("UnregisterService", mock.Anything, "service-a").Return(nil)
	mockRegistry.On("RegisterService", mock.Anything, mock.MatchedBy(func(s *configv1.UpstreamServiceConfig) bool {
		return s.GetName() == "service-a" && s.GetHttpService().GetAddress() == "http://v2"
	})).Return("service-a", nil, nil, nil)

	// Service B should be added
	mockRegistry.On("RegisterService", mock.Anything, mock.MatchedBy(func(s *configv1.UpstreamServiceConfig) bool {
		return s.GetName() == "service-b"
	})).Return("service-b", nil, nil, nil)

	app.reconcileServices(context.Background(), newConfig)

	mockRegistry.AssertExpectations(t)
}

func TestReconcileServices_Sync_UnregisterFailure(t *testing.T) {
	mockRegistry := new(ReconcileMockServiceRegistry)
	app := NewApplication()
	app.ServiceRegistry = mockRegistry
	app.busProvider = nil

	serviceC := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-c"),
	}.Build()

	newConfig := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{}, // Empty config -> remove C
		GlobalSettings: configv1.GlobalSettings_builder{
			AutoDiscoverLocal: proto.Bool(false),
		}.Build(),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{serviceC}, nil)
	// Fail unregister
	mockRegistry.On("UnregisterService", mock.Anything, "service-c").Return(fmt.Errorf("unregister failed"))

	// Should not panic
	assert.NotPanics(t, func() {
		app.reconcileServices(context.Background(), newConfig)
	})

	mockRegistry.AssertExpectations(t)
}

func TestReconcileServices_Sync_RegisterFailure(t *testing.T) {
	mockRegistry := new(ReconcileMockServiceRegistry)
	app := NewApplication()
	app.ServiceRegistry = mockRegistry
	app.busProvider = nil

	serviceB := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-b"),
	}.Build()

	newConfig := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{serviceB},
		GlobalSettings: configv1.GlobalSettings_builder{
			AutoDiscoverLocal: proto.Bool(false),
		}.Build(),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{}, nil)
	// Fail register
	mockRegistry.On("RegisterService", mock.Anything, mock.MatchedBy(func(s *configv1.UpstreamServiceConfig) bool {
		return s.GetName() == "service-b"
	})).Return("", nil, nil, fmt.Errorf("register failed"))

	// Should not panic
	assert.NotPanics(t, func() {
		app.reconcileServices(context.Background(), newConfig)
	})

	mockRegistry.AssertExpectations(t)
}

func TestReconcileServices_Bus_Publish(t *testing.T) {
	mockRegistry := new(ReconcileMockServiceRegistry)
	busProvider, _ := bus.NewProvider(nil)
	app := NewApplication()
	app.ServiceRegistry = mockRegistry
	app.busProvider = busProvider

	// Hook into GetBus to return our MockBus
	mockBus := &MockBus{}
	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == bus.ServiceRegistrationRequestTopic {
			return mockBus, nil
		}
		return nil, nil
	}
	defer func() { bus.GetBusHook = nil }()

	serviceB := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-b"),
	}.Build()

	newConfig := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{serviceB},
		GlobalSettings: configv1.GlobalSettings_builder{
			AutoDiscoverLocal: proto.Bool(false),
		}.Build(),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{}, nil)

	app.reconcileServices(context.Background(), newConfig)

	assert.Len(t, mockBus.PublishedMessages, 1)
	assert.Equal(t, "service-b", mockBus.PublishedMessages[0].Config.GetName())
}

func TestReconcileServices_Bus_PublishFailure(t *testing.T) {
	mockRegistry := new(ReconcileMockServiceRegistry)
	busProvider, _ := bus.NewProvider(nil)
	app := NewApplication()
	app.ServiceRegistry = mockRegistry
	app.busProvider = busProvider

	mockBus := &MockBus{PublishError: fmt.Errorf("publish failed")}
	bus.GetBusHook = func(p *bus.Provider, topic string) (any, error) {
		if topic == bus.ServiceRegistrationRequestTopic {
			return mockBus, nil
		}
		return nil, nil
	}
	defer func() { bus.GetBusHook = nil }()

	serviceB := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("service-b"),
	}.Build()

	newConfig := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{serviceB},
		GlobalSettings: configv1.GlobalSettings_builder{
			AutoDiscoverLocal: proto.Bool(false),
		}.Build(),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{}, nil)

	// Should not panic and should log error (verified by coverage/logs inspection)
	assert.NotPanics(t, func() {
		app.reconcileServices(context.Background(), newConfig)
	})
}

func TestReconcileServices_NoChange(t *testing.T) {
	mockRegistry := new(ReconcileMockServiceRegistry)
	app := NewApplication()
	app.ServiceRegistry = mockRegistry
	app.busProvider = nil

	// We must ensure that the service objects are actually equal according to proto.Equal
	// AND that the logic inside reconcileServices (which clones and sets ID/SanitizedName)
	// results in an equal object.
	serviceA := configv1.UpstreamServiceConfig_builder{
		Name:          proto.String("service-a"),
		Id:            proto.String("service-a"),
		SanitizedName: proto.String("service-a-sanitized"),
	}.Build()

	// Clone to ensure we have distinct objects but same content
	serviceAClone := proto.Clone(serviceA).(*configv1.UpstreamServiceConfig)

	newConfig := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{serviceAClone},
		GlobalSettings: configv1.GlobalSettings_builder{
			AutoDiscoverLocal: proto.Bool(false),
		}.Build(),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{serviceA}, nil)

	// Verify equality pre-condition
	assert.True(t, proto.Equal(serviceA, serviceAClone))

	// No Register or Unregister should be called
	app.reconcileServices(context.Background(), newConfig)

	mockRegistry.AssertExpectations(t)
}
