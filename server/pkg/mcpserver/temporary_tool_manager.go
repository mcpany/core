// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"fmt"
	"sync"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/util"
)

// TemporaryToolManager is a tool manager that stores service info and tools temporarily.
// It is intended for use in ValidateService where we need to store service info
// and discovered tools for the duration of the validation request but discard them afterwards.
//
// Summary: A temporary storage for tool and service information.
type TemporaryToolManager struct {
	NoOpToolManager
	mu          sync.RWMutex
	serviceInfo map[string]*tool.ServiceInfo
	tools       map[string]tool.Tool
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Summary: Initializes a new TemporaryToolManager.
//
// Returns:
//   - *TemporaryToolManager: A new instance of TemporaryToolManager.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
		tools:       make(map[string]tool.Tool),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: Adds service information to the manager.
//
// Parameters:
//   - serviceID: string. The ID of the service.
//   - info: *tool.ServiceInfo. The service information.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: Retrieves service information by ID.
//
// Parameters:
//   - serviceID: string. The ID of the service.
//
// Returns:
//   - *tool.ServiceInfo: The service information if found.
//   - bool: True if the service information exists.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// AddTool implements tool.ManagerInterface.
//
// Summary: Adds a tool to the manager.
//
// Parameters:
//   - t: tool.Tool. The tool to add.
//
// Returns:
//   - error: An error if the tool service ID is empty or name sanitization fails.
func (m *TemporaryToolManager) AddTool(t tool.Tool) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.tools == nil {
		m.tools = make(map[string]tool.Tool)
	}

	if t.Tool().GetServiceId() == "" {
		return fmt.Errorf("tool service ID cannot be empty")
	}

	sanitizedToolName, err := util.SanitizeToolName(t.Tool().GetName())
	if err != nil {
		return fmt.Errorf("failed to sanitize tool name: %w", err)
	}

	toolID := t.Tool().GetServiceId() + "." + sanitizedToolName
	m.tools[toolID] = t
	return nil
}

// GetTool implements tool.ManagerInterface.
//
// Summary: Retrieves a tool by name.
//
// Parameters:
//   - toolName: string. The name of the tool.
//
// Returns:
//   - tool.Tool: The tool if found.
//   - bool: True if the tool exists.
func (m *TemporaryToolManager) GetTool(toolName string) (tool.Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return nil, false
	}
	t, ok := m.tools[toolName]
	return t, ok
}

// ListTools implements tool.ManagerInterface.
//
// Summary: Lists all tools in the manager.
//
// Returns:
//   - []tool.Tool: A slice of all tools.
func (m *TemporaryToolManager) ListTools() []tool.Tool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return nil
	}
	list := make([]tool.Tool, 0, len(m.tools))
	for _, t := range m.tools {
		list = append(list, t)
	}
	return list
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: Counts the number of tools for a given service.
//
// Parameters:
//   - serviceID: string. The ID of the service.
//
// Returns:
//   - int: The number of tools for the service.
func (m *TemporaryToolManager) GetToolCountForService(serviceID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return 0
	}
	count := 0
	for _, t := range m.tools {
		if t.Tool().GetServiceId() == serviceID {
			count++
		}
	}
	return count
}
