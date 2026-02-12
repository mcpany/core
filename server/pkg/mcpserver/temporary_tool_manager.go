// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// Summary: Stores service info temporarily for validation requests.
type TemporaryToolManager struct {
	NoOpToolManager
	serviceInfo map[string]*tool.ServiceInfo
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Summary: Initializes a new TemporaryToolManager.
//
// Returns:
//   - *TemporaryToolManager: The initialized manager.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: Adds service information to the temporary storage.
//
// Parameters:
//   - serviceID: string. The service identifier.
//   - info: *tool.ServiceInfo. The service information.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: Retrieves service information from temporary storage.
//
// Parameters:
//   - serviceID: string. The service identifier.
//
// Returns:
//   - *tool.ServiceInfo: The service information if found.
//   - bool: True if found, false otherwise.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: Returns the tool count for a service (Always 0).
//
// Parameters:
//   - _: string. Unused.
//
// Returns:
//   - int: Always 0.
func (m *TemporaryToolManager) GetToolCountForService(_ string) int {
	return 0
}
