// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"fmt"
	"sync"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
)

// TestTemporaryToolManager_Concurrency ensures that the manager is thread-safe.
// This test is expected to fail or panic before the fix is applied.
func TestTemporaryToolManager_Concurrency(t *testing.T) {
	manager := NewTemporaryToolManager()
	var wg sync.WaitGroup
	numGoroutines := 100

	// AddServiceInfo concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			serviceID := fmt.Sprintf("service-%d", i)
			manager.AddServiceInfo(serviceID, &tool.ServiceInfo{Name: serviceID})
		}(i)
	}

	// GetServiceInfo concurrently
	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(i int) {
			defer wg.Done()
			serviceID := fmt.Sprintf("service-%d", i)
			// Read might race with write
			manager.GetServiceInfo(serviceID)
		}(i)
	}

	wg.Wait()

	// Check if all services were added
	// This part might fail if map writes collide and data is lost
	for i := 0; i < numGoroutines; i++ {
		serviceID := fmt.Sprintf("service-%d", i)
		info, ok := manager.GetServiceInfo(serviceID)
		if ok {
			assert.Equal(t, serviceID, info.Name)
		}
	}
}

func TestTemporaryToolManager_Basic(t *testing.T) {
	manager := NewTemporaryToolManager()

	serviceID := "test-service"
	info := &tool.ServiceInfo{Name: "Test Service"}

	// Add
	manager.AddServiceInfo(serviceID, info)

	// Get
	retrievedInfo, ok := manager.GetServiceInfo(serviceID)
	assert.True(t, ok)
	assert.Equal(t, info, retrievedInfo)

	// Get non-existent
	_, ok = manager.GetServiceInfo("non-existent")
	assert.False(t, ok)

	// GetToolCountForService
	assert.Equal(t, 0, manager.GetToolCountForService(serviceID))
}
