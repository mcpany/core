// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// Summary: is a tool manager that stores service info temporarily.
type TemporaryToolManager struct {
	NoOpToolManager
	serviceInfo map[string]*tool.ServiceInfo
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Summary: creates a new TemporaryToolManager.
//
// Parameters:
//   None.
//
// Returns:
//   - *TemporaryToolManager: The *TemporaryToolManager.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

// AddServiceInfo implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The serviceID.
//   - info: *tool.ServiceInfo. The info.
//
// Returns:
//   None.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - serviceID: string. The serviceID.
//
// Returns:
//   - *tool.ServiceInfo: The *tool.ServiceInfo.
//   - bool: The bool.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}

// GetToolCountForService implements tool.ManagerInterface.
//
// Summary: implements tool.ManagerInterface.
//
// Parameters:
//   - _: string. The _.
//
// Returns:
//   - int: The int.
func (m *TemporaryToolManager) GetToolCountForService(_ string) int {
	return 0
}
