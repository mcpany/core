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
// Summary: Represents a MockServiceRegistry.
type MockServiceRegistry struct {
	mock.Mock
}

// RegisterService registerService register service.
//
// Summary: RegisterService register service.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - serviceConfig (*configv1.UpstreamServiceConfig): The service config.
//
// Returns:
//   - string: The result.
//   - []*configv1.ToolDefinition: The result.
//   - []*configv1.ResourceDefinition: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

// UnregisterService unregisterService unregister service.
//
// Summary: UnregisterService unregister service.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - serviceName (string): The service name.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
//
// Side Effects:
//   - None.
func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

// GetAllServices retrieves the all services.
//
// Summary: Retrieves the all services.
//
// Parameters:
//   None.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The result.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails.
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

// GetServiceInfo retrieves the service info.
//
// Summary: Retrieves the service info.
//
// Parameters:
//   - serviceID (string): The service id.
//
// Returns:
//   - *tool.ServiceInfo: The result.
//   - bool: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if info, ok := args.Get(0).(*tool.ServiceInfo); ok {
		return info, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceConfig retrieves the service config.
//
// Summary: Retrieves the service config.
//
// Parameters:
//   - serviceID (string): The service id.
//
// Returns:
//   - *configv1.UpstreamServiceConfig: The result.
//   - bool: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if config, ok := args.Get(0).(*configv1.UpstreamServiceConfig); ok {
		return config, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceError retrieves the service error.
//
// Summary: Retrieves the service error.
//
// Parameters:
//   - serviceID (string): The service id.
//
// Returns:
//   - string: The result.
//   - bool: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}
