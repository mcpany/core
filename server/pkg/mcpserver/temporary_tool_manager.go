// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"github.com/mcpany/core/server/pkg/tool"
)

// TemporaryToolManager is a tool manager that stores service info temporarily.
//
// Summary: Implements a transient tool manager for validation purposes.
//
// Description:
// It is intended for use in ValidateService where we need to store service info
// for the duration of the validation request but discard it afterwards.
type TemporaryToolManager struct {
	NoOpToolManager
	serviceInfo map[string]*tool.ServiceInfo
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
	}
}

// AddServiceInfo adds service information to the temporary storage.
//
// Summary: Stores service information in memory.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - info: *tool.ServiceInfo. The service information to store.
//
// Side Effects:
//   - Updates the internal map of service information.
func (m *TemporaryToolManager) AddServiceInfo(serviceID string, info *tool.ServiceInfo) {
	if m.serviceInfo == nil {
		m.serviceInfo = make(map[string]*tool.ServiceInfo)
	}
	m.serviceInfo[serviceID] = info
}

// GetServiceInfo retrieves service information from the temporary storage.
//
// Summary: Retrieves stored service information by ID.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//
// Returns:
//   - *tool.ServiceInfo: The stored service information if found.
//   - bool: A boolean indicating whether the service information was found.
func (m *TemporaryToolManager) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	if m.serviceInfo == nil {
		return nil, false
	}
	info, ok := m.serviceInfo[serviceID]
	return info, ok
}
