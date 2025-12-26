// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"

	"github.com/mcpany/core/pkg/tool"
	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/mock"
)

// MockServiceRegistry is a mock implementation of serviceregistry.ServiceRegistryInterface
type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *config_v1.UpstreamServiceConfig) (string, []*config_v1.ToolDefinition, []*config_v1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	var tools []*config_v1.ToolDefinition
	if args.Get(1) != nil {
		tools = args.Get(1).([]*config_v1.ToolDefinition)
	}
	var resources []*config_v1.ResourceDefinition
	if args.Get(2) != nil {
		resources = args.Get(2).([]*config_v1.ResourceDefinition)
	}
	return args.String(0), tools, resources, args.Error(3)
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *MockServiceRegistry) GetAllServices() ([]*config_v1.UpstreamServiceConfig, error) {
	args := m.Called()
	return args.Get(0).([]*config_v1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if s := args.Get(0); s != nil {
		return s.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*config_v1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if c := args.Get(0); c != nil {
		return c.(*config_v1.UpstreamServiceConfig), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}
