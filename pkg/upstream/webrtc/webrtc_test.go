/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package webrtc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockToolManager is a mock implementation of the ToolManagerInterface.
type MockToolManager struct {
	mu      sync.Mutex
	tools   map[string]tool.Tool
	lastErr error
}

func NewMockToolManager() *MockToolManager {
	return &MockToolManager{
		tools: make(map[string]tool.Tool),
	}
}

func (m *MockToolManager) AddTool(t tool.Tool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lastErr != nil {
		return m.lastErr
	}
	sanitizedToolName, _ := util.SanitizeToolName(t.Tool().GetName())
	toolID := t.Tool().GetServiceId() + "." + sanitizedToolName
	m.tools[toolID] = t
	return nil
}

func (m *MockToolManager) GetTool(name string) (tool.Tool, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.tools[name]
	return t, ok
}

func (m *MockToolManager) ListTools() []tool.Tool {
	m.mu.Lock()
	defer m.mu.Unlock()
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

func (m *MockToolManager) ClearToolsForService(serviceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
			delete(m.tools, name)
		}
	}
}

func (m *MockToolManager) SetMCPServer(provider tool.MCPServerProvider) {}

func (m *MockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}

func (m *MockToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}

func (m *MockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func TestNewWebrtcUpstream(t *testing.T) {
	poolManager := pool.NewManager()
	upstream := NewWebrtcUpstream(poolManager)
	require.NotNil(t, upstream)
	assert.IsType(t, &WebrtcUpstream{}, upstream)
}

func TestWebrtcUpstream_Register(t *testing.T) {
	t.Run("successful registration", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface

		upstream := NewWebrtcUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String("echo"),
			Description: proto.String("Echoes a message"),
			CallId:      proto.String("echo-call"),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["echo-call"] = configv1.WebrtcCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		webrtcService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		serviceConfig.SetWebrtcService(webrtcService)

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		tools := toolManager.ListTools()
		assert.Len(t, tools, 1)

		sanitizedToolName, _ := util.SanitizeToolName("echo")
		toolID := serviceID + "." + sanitizedToolName
		_, ok := toolManager.GetTool(toolID)
		assert.True(t, ok, "tool should be registered")
	})

	t.Run("nil service config", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		_, _, _, err := upstream.Register(context.Background(), nil, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "service config is nil")
	})

	t.Run("nil webrtc service config", func(t *testing.T) {
		toolManager := NewMockToolManager()
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		serviceConfig.SetWebrtcService(nil)

		_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Equal(t, "webrtc service config is nil", err.Error())
	})

	t.Run("add tool error", func(t *testing.T) {
		toolManager := NewMockToolManager()
		toolManager.lastErr = errors.New("failed to add tool")
		poolManager := pool.NewManager()
		var promptManager prompt.PromptManagerInterface
		var resourceManager resource.ResourceManagerInterface
		upstream := NewWebrtcUpstream(poolManager)

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("echo"),
			CallId: proto.String("echo-call"),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["echo-call"] = configv1.WebrtcCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		webrtcService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-webrtc-service")
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})
}

func TestWebrtcUpstream_Register_ToolNameGeneration(t *testing.T) {
	toolManager := NewMockToolManager()
	poolManager := pool.NewManager()
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface
	upstream := NewWebrtcUpstream(poolManager)

	toolDef := configv1.ToolDefinition_builder{
		Description: proto.String("A test description"),
		CallId:      proto.String("test-call"),
	}.Build()

	webrtcService := &configv1.WebrtcUpstreamService{}
	webrtcService.SetAddress("http://localhost:8080/signal")
	webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
	calls := make(map[string]*configv1.WebrtcCallDefinition)
	calls["test-call"] = configv1.WebrtcCallDefinition_builder{
		Id: proto.String("test-call"),
	}.Build()
	webrtcService.SetCalls(calls)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-webrtc-service-tool-name-generation")
	serviceConfig.SetWebrtcService(webrtcService)

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	tools := toolManager.ListTools()
	assert.Len(t, tools, 1)
	assert.Equal(t, util.SanitizeOperationID("A test description"), tools[0].Tool().GetName())
}
