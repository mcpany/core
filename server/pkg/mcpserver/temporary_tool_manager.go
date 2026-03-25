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
//
// It is intended for use in ValidateService where we need to store service info
// and discovered tools for the duration of the validation request but discard them afterwards.
//
// Summary: Represents a TemporaryToolManager.
type TemporaryToolManager struct {
	NoOpToolManager
	mu          sync.RWMutex
	serviceInfo map[string]*tool.ServiceInfo
	tools       map[string]tool.Tool
}

// NewTemporaryToolManager creates a new temporary tool manager.
//
// Summary: Creates a new temporary tool manager.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - *TemporaryToolManager: The result.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
		tools:       make(map[string]tool.Tool),
	}
}

// AddServiceInfo addServiceInfo add service info.
//
// Summary: AddServiceInfo add service info.
//
// Parameters: - None.
//   - serviceID (string): The service id.
//   - info (*tool.ServiceInfo): The info.
//
// Returns: - None.
//   - None.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo retrieves the service info.
//
// Summary: Retrieves the service info.
//
// Parameters: - None.
//   - serviceID (string): The service id.
//
// Returns: - None.
//   - *tool.ServiceInfo: The result.
//   - bool: The result.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// AddTool addTool add tool.
//
// Summary: AddTool add tool.
//
// Parameters: - None.
//   - t (tool.Tool): The t.
//
// Returns: - None.
//   - error: An error if the operation fails.
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

// GetTool retrieves the tool.
//
// Summary: Retrieves the tool.
//
// Parameters: - None.
//   - toolName (string): The tool name.
//
// Returns: - None.
//   - tool.Tool: The result.
//   - bool: The result.
func (m *TemporaryToolManager) GetTool(toolName string) (tool.Tool, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.tools == nil {
		return nil, false
	}
	t, ok := m.tools[toolName]
	return t, ok
}

// ListTools retrieves a list of tools.
//
// Summary: Retrieves a list of tools.
//
// Parameters: - None.
//   - None.
//
// Returns: - None.
//   - []tool.Tool: The result.
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

// GetToolCountForService retrieves the tool count for service.
//
// Summary: Retrieves the tool count for service.
//
// Parameters: - None.
//   - serviceID (string): The service id.
//
// Returns: - None.
//   - int: The result.
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
