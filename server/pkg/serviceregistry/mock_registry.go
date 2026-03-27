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
// Summary. Represents a MockServiceRegistry.
type MockServiceRegistry struct {
	mock.Mock
}

// RegisterService provides registerservice functionality.
//
// Summary: RegisterService.
//
// Parameters.
//   - ctx: The parameter.
//   - serviceConfig: The parameter.
//   - []*configv1.ToolDefinition: The parameter.
//   - []*configv1.ResourceDefinition: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

// UnregisterService provides unregisterservice functionality.
//
// Summary: UnregisterService.
//
// Parameters.
//   - ctx: The parameter.
//   - serviceName: The parameter.
//
// Returns.
//   - result: The result.
func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

// GetAllServices provides getallservices functionality.
//
// Summary: GetAllServices.
//
// Parameters.
//   - ): The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

// GetServiceInfo provides getserviceinfo functionality.
//
// Summary: GetServiceInfo.
//
// Parameters.
//   - serviceID: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if info, ok := args.Get(0).(*tool.ServiceInfo); ok {
		return info, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceConfig provides getserviceconfig functionality.
//
// Summary: GetServiceConfig.
//
// Parameters.
//   - serviceID: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if config, ok := args.Get(0).(*configv1.UpstreamServiceConfig); ok {
		return config, args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceError provides getserviceerror functionality.
//
// Summary: GetServiceError.
//
// Parameters.
//   - serviceID: The parameter.
//   - bool: The parameter.
//
// Returns.
//   - None.
func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}
