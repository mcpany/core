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
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWebsocketUpstream_createAndRegisterWebsocketTools_DisabledTool(t *testing.T) {
	toolManager := NewMockToolManager(nil)
	poolManager := pool.NewManager()
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	upstream := NewWebsocketUpstream(poolManager)

	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("echo"),
		Description: proto.String("Echoes a message"),
		CallId:      proto.String("echo-call"),
		Disable:     proto.Bool(true),
	}.Build()

	websocketService := &configv1.WebsocketUpstreamService{}
	websocketService.SetAddress("ws://localhost:8080/echo")
	websocketService.SetTools([]*configv1.ToolDefinition{toolDef})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-websocket-service")
	serviceConfig.SetWebsocketService(websocketService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	assert.Empty(t, discoveredTools)
}

func TestWebsocketUpstream_createAndRegisterWebsocketTools_MissingCallDefinition(t *testing.T) {
	toolManager := NewMockToolManager(nil)
	poolManager := pool.NewManager()
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	upstream := NewWebsocketUpstream(poolManager)

	toolDef := configv1.ToolDefinition_builder{
		Name:        proto.String("echo"),
		Description: proto.String("Echoes a message"),
		CallId:      proto.String("echo-call"),
	}.Build()

	websocketService := &configv1.WebsocketUpstreamService{}
	websocketService.SetAddress("ws://localhost:8080/echo")
	websocketService.SetTools([]*configv1.ToolDefinition{toolDef})

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-websocket-service")
	serviceConfig.SetWebsocketService(websocketService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	assert.Empty(t, discoveredTools)
}

func TestWebsocketUpstream_createAndRegisterWebsocketTools_MissingToolName(t *testing.T) {
	toolManager := NewMockToolManager(nil)
	poolManager := pool.NewManager()
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	upstream := NewWebsocketUpstream(poolManager)

	toolDef := configv1.ToolDefinition_builder{
		Description: proto.String("Echoes a message"),
		CallId:      proto.String("echo-call"),
	}.Build()

	websocketService := &configv1.WebsocketUpstreamService{}
	websocketService.SetAddress("ws://localhost:8080/echo")
	websocketService.SetTools([]*configv1.ToolDefinition{toolDef})
	calls := make(map[string]*configv1.WebsocketCallDefinition)
	calls["echo-call"] = configv1.WebsocketCallDefinition_builder{
		Id: proto.String("echo-call"),
	}.Build()
	websocketService.SetCalls(calls)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-websocket-service")
	serviceConfig.SetWebsocketService(websocketService)

	_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
	require.NoError(t, err)

	assert.NotEmpty(t, discoveredTools)
}

func TestWebsocketUpstream_createAndRegisterWebsocketTools_DynamicResourceMissingTool(t *testing.T) {
	toolManager := tool.NewToolManager(nil)
	resourceManager := resource.NewResourceManager()
	poolManager := pool.NewManager()
	upstream := NewWebsocketUpstream(poolManager)

	dynamicResource := configv1.ResourceDefinition_builder{
		Name: proto.String("test-resource"),
		Dynamic: configv1.DynamicResource_builder{
			WebsocketCall: configv1.WebsocketCallDefinition_builder{
				Id: proto.String("missing-tool"),
			}.Build(),
		}.Build(),
	}.Build()

	websocketService := configv1.WebsocketUpstreamService_builder{
		Address:   proto.String("ws://localhost:8080/test"),
		Resources: []*configv1.ResourceDefinition{dynamicResource},
	}.Build()

	serviceConfig := configv1.UpstreamServiceConfig_builder{
		Name:             proto.String("test-websocket-service"),
		WebsocketService: websocketService,
	}.Build()

	_, _, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, nil, resourceManager, false)
	require.NoError(t, err)

	assert.Empty(t, resourceManager.ListResources())
}
