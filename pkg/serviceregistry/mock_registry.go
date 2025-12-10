// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"

	"github.com/mcpany/core/pkg/healthstatus"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	config "github.com/mcpany/core/proto/config/v1"
)

// MockServiceRegistry is a mock implementation of the ServiceRegistryInterface.
type MockServiceRegistry struct {
	RegisteredServices map[string]*config.UpstreamServiceConfig
	HealthStatus       map[string]healthstatus.HealthStatus
}

// NewMockServiceRegistry creates a new mock service registry.
func NewMockServiceRegistry() *MockServiceRegistry {
	return &MockServiceRegistry{
		RegisteredServices: make(map[string]*config.UpstreamServiceConfig),
		HealthStatus:       make(map[string]healthstatus.HealthStatus),
	}
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *config.UpstreamServiceConfig) (string, []*config.ToolDefinition, []*config.ResourceDefinition, error) {
	id := serviceConfig.GetName()
	m.RegisteredServices[id] = serviceConfig
	m.HealthStatus[id] = healthstatus.UNKNOWN
	return id, nil, nil, nil
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	delete(m.RegisteredServices, serviceName)
	delete(m.HealthStatus, serviceName)
	return nil
}

func (m *MockServiceRegistry) GetAllServices() ([]*config.UpstreamServiceConfig, error) {
	services := make([]*config.UpstreamServiceConfig, 0, len(m.RegisteredServices))
	for _, service := range m.RegisteredServices {
		services = append(services, service)
	}
	return services, nil
}

func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *MockServiceRegistry) SetHealthStatus(serviceID string, status healthstatus.HealthStatus) {
	m.HealthStatus[serviceID] = status
}

func (m *MockServiceRegistry) GetService(serviceID string) (upstream.Upstream, error) {
	return nil, nil
}
