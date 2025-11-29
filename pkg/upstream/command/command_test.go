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

package command

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// mockToolManager to simulate errors
type mockToolManager struct {
	tool.ToolManagerInterface
	tools    map[string]tool.Tool
	addError error
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		tools: make(map[string]tool.Tool),
	}
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	if m.addError != nil {
		return m.addError
	}
	sanitizedToolName, err := util.SanitizeToolName(t.Tool().GetName())
	if err != nil {
		return err
	}
	m.tools[t.Tool().GetServiceId()+"."+sanitizedToolName] = t
	return nil
}

func (m *mockToolManager) ExecuteTool(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}

func (m *mockToolManager) SetMCPServer(mcpServer tool.MCPServerProvider) {
}

func (m *mockToolManager) PruneToolsForService(serviceID string, keepToolNames []string) {
}

func (m *mockToolManager) AddMiddleware(middleware tool.ToolExecutionMiddleware) {
}

func (m *mockToolManager) GetTool(name string) (tool.Tool, bool) {
	t, ok := m.tools[name]
	return t, ok
}

func (m *mockToolManager) ListTools() []tool.Tool {
	tools := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		tools = append(tools, t)
	}
	return tools
}

func (m *mockToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {}

func TestNewCommandUpstream(t *testing.T) {
	u := NewCommandUpstream()
	assert.NotNil(t, u)
	_, ok := u.(*CommandUpstream)
	assert.True(t, ok)
}

func TestCommandUpstream_Register(t *testing.T) {
	tm := newMockToolManager()
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	u := NewCommandUpstream()

	t.Run("successful registration", func(t *testing.T) {
		tm := newMockToolManager()
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-command-service")
		cmdService := &configv1.CommandLineUpstreamService{}
		cmdService.SetCommand("/bin/echo")
		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("echo"),
			CallId: proto.String("echo-call"),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Id: proto.String("echo-call"),
		}.Build()
		calls := make(map[string]*configv1.CommandLineCallDefinition)
		calls["echo-call"] = callDef
		cmdService.SetCalls(calls)
		cmdService.SetTools([]*configv1.ToolDefinition{toolDef})
		serviceConfig.SetCommandLineService(cmdService)

		serviceID, discoveredTools, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)
		expectedKey, _ := util.SanitizeServiceName("test-command-service")
		assert.Equal(t, expectedKey, serviceID)
		assert.Len(t, discoveredTools, 1)
		assert.Equal(t, "echo", discoveredTools[0].GetName())
		assert.Len(t, tm.ListTools(), 1)

		// Execute the registered tool
		cmdTool := tm.ListTools()[0]
		assert.Equal(t, "echo", cmdTool.Tool().GetName())

		cmdTool.Tool().GetOutputSchema()
		outputSchema := cmdTool.Tool().GetOutputSchema()
		assert.NotNil(t, outputSchema)
		assert.Equal(t, "object", outputSchema.Fields["type"].GetStringValue())
		properties := outputSchema.Fields["properties"].GetStructValue().GetFields()
		assert.Contains(t, properties, "stdout")
		assert.Equal(t, "string", properties["stdout"].GetStructValue().GetFields()["type"].GetStringValue())

		inputData := map[string]interface{}{"args": []string{"-n", "hello from test"}}
		inputs, err := json.Marshal(inputData)
		require.NoError(t, err)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := cmdTool.Execute(context.Background(), req)
		require.NoError(t, err)

		resultMap, ok := result.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "hello from test", resultMap["stdout"])
		assert.Equal(t, "/bin/echo", resultMap["command"])
	})

	t.Run("successful prompt registration", func(t *testing.T) {
		tm := newMockToolManager()
		prm := prompt.NewPromptManager()
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-prompt-service")
		cmdService := &configv1.CommandLineUpstreamService{}
		promptDef := configv1.PromptDefinition_builder{
			Name: proto.String("test-prompt"),
		}.Build()
		cmdService.SetPrompts([]*configv1.PromptDefinition{promptDef})
		serviceConfig.SetCommandLineService(cmdService)

		serviceID, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)

		sanitizedName, _ := util.SanitizeServiceName("test-prompt-service")
		assert.Equal(t, sanitizedName, serviceID)

		p, ok := prm.GetPrompt(serviceID + ".test-prompt")
		require.True(t, ok)
		assert.Equal(t, serviceID+".test-prompt", p.Prompt().Name)
	})

	t.Run("successful dynamic resource registration", func(t *testing.T) {
		tm := newMockToolManager()
		rm := resource.NewResourceManager()
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-dynamic-resource-service")
		cmdService := &configv1.CommandLineUpstreamService{}

		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("list-files"),
			CallId: proto.String("list-files-call"),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Id: proto.String("list-files-call"),
		}.Build()
		calls := make(map[string]*configv1.CommandLineCallDefinition)
		calls["list-files-call"] = callDef
		cmdService.SetCalls(calls)
		cmdService.SetTools([]*configv1.ToolDefinition{toolDef})

		resourceDef := &configv1.ResourceDefinition{}
		resourceDef.SetName("files")
		resourceDef.SetUri("test-dynamic-resource-service.files")
		dynamicResource := &configv1.DynamicResource{}
		commandLineCall := &configv1.CommandLineCallDefinition{}
		commandLineCall.SetId("list-files-call")
		dynamicResource.SetCommandLineCall(commandLineCall)
		resourceDef.SetDynamic(dynamicResource)

		cmdService.SetResources([]*configv1.ResourceDefinition{resourceDef})
		serviceConfig.SetCommandLineService(cmdService)

		serviceID, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)
		assert.Len(t, rm.ListResources(), 1)
		dynResource, ok := rm.GetResource(serviceID + ".files")
		require.True(t, ok)
		assert.Equal(t, "files", dynResource.Resource().Name)
	})

	t.Run("missing call definition", func(t *testing.T) {
		tm := newMockToolManager()
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-missing-call-def")
		cmdService := &configv1.CommandLineUpstreamService{}
		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("echo"),
			CallId: proto.String("echo-call-missing"),
		}.Build()
		cmdService.SetTools([]*configv1.ToolDefinition{toolDef})
		serviceConfig.SetCommandLineService(cmdService)

		_, _, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tm,
			prm,
			rm,
			false,
		)
		require.NoError(t, err)
		assert.Len(t, tm.ListTools(), 0)
	})

	t.Run("nil command line service config", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-nil-config")
		_, _, _, err := u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "command line service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("") // empty name
		_, _, _, err := u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
		require.Error(t, err)
	})

	t.Run("add tool error", func(t *testing.T) {
		tmWithError := newMockToolManager()
		tmWithError.addError = errors.New("failed to add tool")

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-add-tool-error")
		cmdService := &configv1.CommandLineUpstreamService{}
		toolDef := configv1.ToolDefinition_builder{
			Name:   proto.String("ls"),
			CallId: proto.String("ls-call"),
		}.Build()
		callDef := configv1.CommandLineCallDefinition_builder{
			Id: proto.String("ls-call"),
		}.Build()
		calls := make(map[string]*configv1.CommandLineCallDefinition)
		calls["ls-call"] = callDef
		cmdService.SetCalls(calls)
		cmdService.SetTools([]*configv1.ToolDefinition{toolDef})
		serviceConfig.SetCommandLineService(cmdService)

		_, discoveredTools, _, err := u.Register(
			context.Background(),
			serviceConfig,
			tmWithError,
			prm,
			rm,
			false,
		)
		require.Error(t, err)
		assert.Nil(t, discoveredTools)
		assert.Empty(t, tmWithError.ListTools())
		assert.Contains(t, err.Error(), "failed to add tool")
	})
}
