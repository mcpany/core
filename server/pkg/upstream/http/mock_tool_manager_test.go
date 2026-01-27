// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"errors"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// mockToolManager to simulate errors
type mockToolManager struct {
	tool.ManagerInterface
	addError    error
	addedTools  []tool.Tool
	failOnClear bool
}

func newMockToolManager() *mockToolManager {
	return &mockToolManager{
		addedTools: make([]tool.Tool, 0),
	}
}

func (m *mockToolManager) AddTool(t tool.Tool) error {
	if m.addError != nil {
		return m.addError
	}
	m.addedTools = append(m.addedTools, t)
	return nil
}

func (m *mockToolManager) GetTool(name string) (tool.Tool, bool) {
	for _, t := range m.addedTools {
		sanitizedToolName, _ := util.SanitizeToolName(t.Tool().GetName())
		toolID := t.Tool().GetServiceId() + "." + sanitizedToolName
		if toolID == name {
			return t, true
		}
	}
	return nil, false
}

func (m *mockToolManager) ListTools() []tool.Tool {
	return m.addedTools
}

func (m *mockToolManager) ListServices() []*tool.ServiceInfo {
	return nil
}

func (m *mockToolManager) ClearToolsForService(serviceID string) {
	if m.failOnClear {
		return
	}
	var remainingTools []tool.Tool
	for _, t := range m.addedTools {
		if t.Tool().GetServiceId() != serviceID {
			remainingTools = append(remainingTools, t)
		}
	}
	m.addedTools = remainingTools
}

func (m *mockToolManager) AddServiceInfo(_ string, _ *tool.ServiceInfo) {}
func (m *mockToolManager) GetServiceInfo(_ string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *mockToolManager) SetMCPServer(_ tool.MCPServerProvider) {}
func (m *mockToolManager) CallTool(_ context.Context, _ *tool.ExecutionRequest) (any, error) {
	return nil, errors.New("not implemented")
}

func (m *mockToolManager) SetProfiles(_ []string, _ []*configv1.ProfileDefinition) {}
