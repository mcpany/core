// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/mcpany/core/server/pkg/worker"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockServiceRegistry is a simple mock for serviceregistry.ServiceRegistryInterface
type MockServiceRegistry struct {
	serviceregistry.ServiceRegistryInterface
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, config *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return "service1", nil, nil, nil
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, name string) error {
	return nil
}

func (m *MockServiceRegistry) GetServiceStatus(serviceID string) string {
	return "OK"
}

func TestServiceRegistrationWorker_Stop(t *testing.T) {
	// Setup bus
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	// Setup worker
	registry := &MockServiceRegistry{}
	w := worker.NewServiceRegistrationWorker(b, registry)

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// Ensure it started (async)
	time.Sleep(10 * time.Millisecond)

	// Test Stop (graceful shutdown)
	cancel()
	w.Stop()

	// If we reached here, it didn't deadlock
	assert.True(t, true)
}
