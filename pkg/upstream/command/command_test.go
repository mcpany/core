/*
 * Copyright 2025 Author(s) of MCP-XY
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
	"errors"
	"testing"

	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/util"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	m.tools[t.Tool().GetName()] = t
	return nil
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
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-command-service")
		cmdService := &configv1.CommandLineUpstreamService{}
		callDef := &configv1.StdioCallDefinition{}
		callDef.SetMethod("echo")
		cmdService.SetCalls([]*configv1.StdioCallDefinition{callDef})
		serviceConfig.SetCommandLineService(cmdService)

		serviceKey, discoveredTools, err := u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
		require.NoError(t, err)
		expectedKey, _ := util.GenerateID("test-command-service")
		assert.Equal(t, expectedKey, serviceKey)
		assert.Len(t, discoveredTools, 1)
		assert.Equal(t, "echo", discoveredTools[0].GetName())
		assert.Len(t, tm.ListTools(), 1)
	})

	t.Run("nil command line service config", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-nil-config")
		_, _, err := u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "command line service config is nil")
	})

	t.Run("invalid service name", func(t *testing.T) {
		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("") // empty name
		_, _, err := u.Register(context.Background(), serviceConfig, tm, prm, rm, false)
		require.Error(t, err)
	})

	t.Run("add tool error", func(t *testing.T) {
		tmWithError := newMockToolManager()
		tmWithError.addError = errors.New("failed to add tool")

		serviceConfig := &configv1.UpstreamServiceConfig{}
		serviceConfig.SetName("test-add-tool-error")
		cmdService := &configv1.CommandLineUpstreamService{}
		callDef := &configv1.StdioCallDefinition{}
		callDef.SetMethod("ls")
		cmdService.SetCalls([]*configv1.StdioCallDefinition{callDef})
		serviceConfig.SetCommandLineService(cmdService)

		_, discoveredTools, err := u.Register(context.Background(), serviceConfig, tmWithError, prm, rm, false)
		require.NoError(t, err)
		assert.Empty(t, discoveredTools)
		assert.Empty(t, tmWithError.ListTools())
	})
}
