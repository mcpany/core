// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package worker_test

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/worker"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockToolManager is a simple mock for tool.ManagerInterface
type MockToolManager struct{}

func (m *MockToolManager) AddTool(_ tool.Tool) error { return nil }
func (m *MockToolManager) GetTool(_ string) (tool.Tool, bool) { return nil, false }
func (m *MockToolManager) ListTools() []tool.Tool { return nil }
func (m *MockToolManager) ClearToolsForService(_ string) {}
func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *MockToolManager) SetMCPServer(_ tool.MCPServerProvider) {}
func (m *MockToolManager) AddMiddleware(_ tool.ExecutionMiddleware) {}
func (m *MockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *MockToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) { return nil, false }
func (m *MockToolManager) ListServices() []*tool.ServiceInfo { return nil }
func (m *MockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}

func TestUpstreamWorker_Stop(t *testing.T) {
	// Setup bus
	b, err := bus.NewProvider(nil)
	require.NoError(t, err)

	// Setup worker
	toolManager := &MockToolManager{}
	w := worker.NewUpstreamWorker(b, toolManager)

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
