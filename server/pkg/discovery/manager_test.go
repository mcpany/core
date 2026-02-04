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

type SlowMockProvider struct {
	MockProvider
	delay time.Duration
}

func (m *SlowMockProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	time.Sleep(m.delay)
	return m.MockProvider.Discover(ctx)
}

func TestManager_Run(t *testing.T) {
	manager := NewManager()

	// Provider 1: Success
	p1 := &MockProvider{
		name: "p1",
		services: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{Name: pointer("s1")}.Build(),
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
	assert.Equal(t, "s1", services[0].GetName())

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

func TestManager_Run_Parallel(t *testing.T) {
	manager := NewManager()

	// Provider 1: Slow (50ms)
	p1 := &SlowMockProvider{
		MockProvider: MockProvider{name: "p1"},
		delay:        50 * time.Millisecond,
	}
	manager.RegisterProvider(p1)

	// Provider 2: Slow (50ms)
	p2 := &SlowMockProvider{
		MockProvider: MockProvider{name: "p2"},
		delay:        50 * time.Millisecond,
	}
	manager.RegisterProvider(p2)

	start := time.Now()
	manager.Run(context.Background())
	duration := time.Since(start)

	// If sequential, it would take > 100ms. If parallel, it should take ~50ms (plus overhead).
	// We allow some buffer, so if it's < 90ms it's definitely parallel.
	assert.Less(t, duration, 90*time.Millisecond, "Discovery took too long, expected parallel execution")
}

func pointer(s string) *string {
	return &s
}

func TestManager_GetStatuses(t *testing.T) {
	manager := NewManager()

	p1 := &MockProvider{name: "p1"}
	manager.RegisterProvider(p1)

	statuses := manager.GetStatuses()
	assert.Len(t, statuses, 1)
	assert.Equal(t, "p1", statuses[0].Name)
	assert.Equal(t, "PENDING", statuses[0].Status)
}
