// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

type MockProvider struct {
	name     string
	services []*configv1.UpstreamServiceConfig
	err      error
}

func (m *MockProvider) Name() string {
	return m.name
}

func (m *MockProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return m.services, m.err
}

func TestManager_Run(t *testing.T) {
	manager := NewManager()

	// Provider 1: Success
	p1 := &MockProvider{
		name: "p1",
		services: []*configv1.UpstreamServiceConfig{
			{Name: pointer("s1")},
		},
	}
	manager.RegisterProvider(p1)

	// Provider 2: Failure
	p2 := &MockProvider{
		name: "p2",
		err:  errors.New("failed"),
	}
	manager.RegisterProvider(p2)

	// Run discovery
	services := manager.Run(context.Background())

	// Check results
	assert.Len(t, services, 1)
	assert.Equal(t, "s1", *services[0].Name)

	// Check statuses
	status1, ok := manager.GetProviderStatus("p1")
	assert.True(t, ok)
	assert.Equal(t, "OK", status1.Status)
	assert.Equal(t, 1, status1.DiscoveredCount)
	assert.Empty(t, status1.LastError)
	assert.WithinDuration(t, time.Now(), status1.LastRunAt, time.Second)

	status2, ok := manager.GetProviderStatus("p2")
	assert.True(t, ok)
	assert.Equal(t, "ERROR", status2.Status)
	assert.Equal(t, "failed", status2.LastError)
	assert.WithinDuration(t, time.Now(), status2.LastRunAt, time.Second)
}

func TestManager_GetStatuses(t *testing.T) {
	manager := NewManager()

	// Register multiple providers
	p1 := &MockProvider{name: "p1"}
	p2 := &MockProvider{name: "p2"}
	manager.RegisterProvider(p1)
	manager.RegisterProvider(p2)

	// Initial check - should be PENDING
	statuses := manager.GetStatuses()
	assert.Len(t, statuses, 2)

	statusMap := make(map[string]*ProviderStatus)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}

	assert.Equal(t, "PENDING", statusMap["p1"].Status)
	assert.Equal(t, "PENDING", statusMap["p2"].Status)

	// Run discovery
	manager.Run(context.Background())

	// Check updated statuses
	statuses = manager.GetStatuses()
	assert.Len(t, statuses, 2)

	statusMap = make(map[string]*ProviderStatus)
	for _, s := range statuses {
		statusMap[s.Name] = s
	}

	assert.Equal(t, "OK", statusMap["p1"].Status)
	assert.Equal(t, "OK", statusMap["p2"].Status)
}

func pointer(s string) *string {
	return &s
}
