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

package websocket

import (
	"context"
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type mockPromptManager struct {
	prompt.PromptManagerInterface
	prompts map[string]prompt.Prompt
}

func newMockPromptManager() *mockPromptManager {
	return &mockPromptManager{
		prompts: make(map[string]prompt.Prompt),
	}
}

func (m *mockPromptManager) AddPrompt(p prompt.Prompt) {
	m.prompts[p.Prompt().Name] = p
}

func (m *mockPromptManager) GetPrompt(name string) (prompt.Prompt, bool) {
	p, ok := m.prompts[name]
	return p, ok
}

func TestWebsocketUpstream_Coverage(t *testing.T) {
	toolManager := NewMockToolManager(nil)
	poolManager := pool.NewManager()
	promptManager := newMockPromptManager()
	var resourceManager resource.ResourceManagerInterface

	upstream := NewWebsocketUpstream(poolManager)

	t.Run("Disabled Tool", func(t *testing.T) {
		toolDef := configv1.ToolDefinition_builder{
			Name:    proto.String("disabled-tool"),
			CallId:  proto.String("call1"),
			Disable: proto.Bool(true),
		}.Build()

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")
		websocketService.SetTools([]*configv1.ToolDefinition{toolDef})
		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["call1"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call1"),
		}.Build()
		websocketService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-disabled-tool")
		serviceConfig.SetWebsocketService(websocketService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)

		// Check TM
		sanitizedName, _ := util.SanitizeServiceName("test-disabled-tool")
		// Note: ListTools() returns all tools from all services in mock
		// But MockToolManager doesn't clear between runs unless we tell it
		// Actually we reuse toolManager here.

		// Better to check by ID
		_, ok := toolManager.GetTool(sanitizedName + ".disabled-tool")
		assert.False(t, ok)
	})

	t.Run("Prompts Registration", func(t *testing.T) {
		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")

		promptDef1 := configv1.PromptDefinition_builder{
			Name: proto.String("prompt1"),
		}.Build()
		promptDef2 := configv1.PromptDefinition_builder{
			Name:    proto.String("prompt2"),
			Disable: proto.Bool(true),
		}.Build()

		websocketService.SetPrompts([]*configv1.PromptDefinition{promptDef1, promptDef2})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-prompts")
		serviceConfig.SetWebsocketService(websocketService)

		serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)

		// Check prompt1 exists
		_, ok := promptManager.GetPrompt(serviceID + ".prompt1")
		assert.True(t, ok)

		// Check prompt2 does not exist
		_, ok = promptManager.GetPrompt(serviceID + ".prompt2")
		assert.False(t, ok)
	})

	t.Run("Missing Call Definition", func(t *testing.T) {
		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("missing-call-tool"),
			CallId: proto.String("missing-call"),
		}.Build()

		websocketService.SetTools([]*configv1.ToolDefinition{toolDef})
		// No calls map set

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-missing-call")
		serviceConfig.SetWebsocketService(websocketService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)

		sanitizedName, _ := util.SanitizeServiceName("test-missing-call")
		_, ok := toolManager.GetTool(sanitizedName + ".missing-call-tool")
		assert.False(t, ok)
	})

	t.Run("Add Tool Error", func(t *testing.T) {
		// Create a new tool manager with error set
		tm := NewMockToolManager(nil)
		tm.lastErr = assert.AnError

		websocketService := &configv1.WebsocketUpstreamService{}
		websocketService.SetAddress("ws://localhost:8080/echo")

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("error-tool"),
			CallId: proto.String("call1"),
		}.Build()

		calls := make(map[string]*configv1.WebsocketCallDefinition)
		calls["call1"] = configv1.WebsocketCallDefinition_builder{
			Id: proto.String("call1"),
		}.Build()
		websocketService.SetCalls(calls)
		websocketService.SetTools([]*configv1.ToolDefinition{toolDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-add-error")
		serviceConfig.SetWebsocketService(websocketService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.NoError(t, err) // It logs error and continues
		assert.Empty(t, discoveredTools)
	})
}
