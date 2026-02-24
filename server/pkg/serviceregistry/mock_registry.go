// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/mock"
)

// MockServiceRegistry is a mock implementation of ServiceRegistryInterface.
//
// Summary: Mock service registry for testing.
type MockServiceRegistry struct {
	mock.Mock
}

// RegisterService registers a new upstream service based on the provided configuration.
//
// Summary: Mock implementation of RegisterService.
//
// Parameters:
//   - ctx: context.Context. The registration context.
//   - serviceConfig: *configv1.UpstreamServiceConfig. The configuration for the service.
//
// Returns:
//   - string: The unique service ID.
//   - []*configv1.ToolDefinition: A list of discovered tools.
//   - []*configv1.ResourceDefinition: A list of discovered resources.
//   - error: An error if registration fails.
func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

// UnregisterService removes a service from the registry.
//
// Summary: Mock implementation of UnregisterService.
//
// Parameters:
//   - ctx: context.Context. The context for the unregistration.
//   - serviceName: string. The name of the service to remove.
//
// Returns:
//   - error: An error if the service is not found or shutdown fails.
func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

// GetAllServices returns a list of all currently registered services.
//
// Side Effects:
//   - None.
func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

// GetServiceInfo retrieves the metadata for a service by its ID.
//
// Summary: Mock implementation of GetServiceInfo.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *tool.ServiceInfo: The service metadata.
//   - bool: True if the service was found, false otherwise.
func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if info, ok := args.Get(0).(*tool.ServiceInfo); ok {
		return info, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceConfig returns the configuration for a given service ID.
//
// Summary: Mock implementation of GetServiceConfig.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The service configuration.
//   - bool: True if the service was found, false otherwise.
func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if config, ok := args.Get(0).(*configv1.UpstreamServiceConfig); ok {
		return config, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceError returns the last known registration or health error for a service.
//
// Summary: Mock implementation of GetServiceError.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - string: The error message.
//   - bool: True if an error is present, false otherwise.
func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}
