// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// It is intended for use in ValidateService where we need to store service info
// for the duration of the validation request but discard it afterwards.
type TemporaryToolManager struct {
	NoOpToolManager
	serviceInfo map[string]*tool.ServiceInfo
}

// NewTemporaryToolManager creates a new TemporaryToolManager.
//
// Returns:
//   - *TemporaryToolManager: A new instance of TemporaryToolManager.
func NewTemporaryToolManager() *TemporaryToolManager {
	return &TemporaryToolManager{
		serviceInfo: make(map[string]*tool.ServiceInfo),
	}
}

// AddServiceInfo adds service information to the manager.
//
// Parameters:
//   - serviceID: string. The ID of the service.
//   - info: *tool.ServiceInfo. The service information.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo retrieves service information by ID.
//
// Parameters:
//   - serviceID: string. The ID of the service.
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
