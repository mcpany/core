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
	"testing"

	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestWebrtcUpstream_Coverage(t *testing.T) {
	toolManager := NewMockToolManager()
	poolManager := pool.NewManager()
	promptManager := NewMockPromptManager()
	resourceManager := NewMockResourceManager()

	upstream := NewWebrtcUpstream(poolManager)

	t.Run("Disabled Items", func(t *testing.T) {
		toolDef := configv1.ToolDefinition_builder{
			Name:    proto.String("disabled-tool"),
			CallId:  proto.String("call1"),
			Disable: proto.Bool(true),
		}.Build()

		promptDef := configv1.PromptDefinition_builder{
			Name:    proto.String("disabled-prompt"),
			Disable: proto.Bool(true),
		}.Build()

		resourceDef := configv1.ResourceDefinition_builder{
			Name:    proto.String("disabled-resource"),
			Disable: proto.Bool(true),
		}.Build()

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		webrtcService.SetPrompts([]*configv1.PromptDefinition{promptDef})
		webrtcService.SetResources([]*configv1.ResourceDefinition{resourceDef})

		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["call1"] = configv1.WebrtcCallDefinition_builder{
			Id: proto.String("call1"),
		}.Build()
		webrtcService.SetCalls(calls)

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-disabled-items")
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, discoveredResources, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
		assert.Empty(t, discoveredResources)

		sanitizedName, _ := util.SanitizeServiceName("test-disabled-items")
		_, ok := toolManager.GetTool(sanitizedName + ".disabled-tool")
		assert.False(t, ok)

		_, ok = promptManager.GetPrompt(sanitizedName + ".disabled-prompt")
		assert.False(t, ok)

		_, ok = resourceManager.GetResource("disabled-resource")
		assert.False(t, ok)
	})

	t.Run("Missing Call Definition", func(t *testing.T) {
		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("missing-call-tool"),
			CallId: proto.String("missing-call"),
		}.Build()

		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})
		// No calls map set

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-missing-call")
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, toolManager, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)

		sanitizedName, _ := util.SanitizeServiceName("test-missing-call")
		_, ok := toolManager.GetTool(sanitizedName + ".missing-call-tool")
		assert.False(t, ok)
	})

	t.Run("Add Tool Error", func(t *testing.T) {
		tm := NewMockToolManager()
		tm.lastErr = assert.AnError

		webrtcService := &configv1.WebrtcUpstreamService{}
		webrtcService.SetAddress("http://localhost:8080/signal")

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("error-tool"),
			CallId: proto.String("call1"),
		}.Build()

		calls := make(map[string]*configv1.WebrtcCallDefinition)
		calls["call1"] = configv1.WebrtcCallDefinition_builder{
			Id: proto.String("call1"),
		}.Build()
		webrtcService.SetCalls(calls)
		webrtcService.SetTools([]*configv1.ToolDefinition{toolDef})

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-add-error")
		serviceConfig.SetWebrtcService(webrtcService)

		_, discoveredTools, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
	})
}
