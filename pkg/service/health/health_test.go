// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockCheckable is a mock implementation of the Checkable interface for testing.
type mockCheckable struct {
	id         string
	interval   time.Duration
	checkFunc  func(ctx context.Context) error
	checkCount int
	mu         sync.Mutex
}

func newMockCheckable(id string, interval time.Duration, checkFunc func(ctx context.Context) error) *mockCheckable {
	return &mockCheckable{
		id:        id,
		interval:  interval,
		checkFunc: checkFunc,
	}
}

func (m *mockCheckable) ID() string {
	return m.id
}

func (m *mockCheckable) Interval() time.Duration {
	return m.interval
}

func (m *mockCheckable) HealthCheck(ctx context.Context) error {
	m.mu.Lock()
	m.checkCount++
	m.mu.Unlock()
	if m.checkFunc != nil {
		return m.checkFunc(ctx)
	}
	return nil
}

func (m *mockCheckable) getCheckCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.checkCount
}

func TestManager_StartStop(t *testing.T) {
	manager := NewManager()
	defer manager.Shutdown()

	healthyService := newMockCheckable("healthy-service", 100*time.Millisecond, nil)
	unhealthyService := newMockCheckable("unhealthy-service", 100*time.Millisecond, func(ctx context.Context) error {
		return fmt.Errorf("service is unhealthy")
	})

	manager.Start(healthyService)
	manager.Start(unhealthyService)

	// Allow time for a few checks
	time.Sleep(350 * time.Millisecond)

	assert.Equal(t, StatusHealthy, manager.Status("healthy-service"), "Healthy service should be reported as healthy")
	assert.Equal(t, StatusUnhealthy, manager.Status("unhealthy-service"), "Unhealthy service should be reported as unhealthy")

	require.GreaterOrEqual(t, healthyService.getCheckCount(), 3, "Healthy service should have been checked multiple times")
	require.GreaterOrEqual(t, unhealthyService.getCheckCount(), 3, "Unhealthy service should have been checked multiple times")

	manager.Stop("healthy-service")
	healthyServiceStartCount := healthyService.getCheckCount()

	// Allow more time for checks to run
	time.Sleep(250 * time.Millisecond)

	assert.Equal(t, StatusUnknown, manager.Status("healthy-service"), "Stopped service should have an unknown status")
	assert.Equal(t, healthyServiceStartCount, healthyService.getCheckCount(), "Check count for stopped service should not increase")
	require.GreaterOrEqual(t, unhealthyService.getCheckCount(), 5, "Unhealthy service should continue to be checked")
}

func TestManager_Status(t *testing.T) {
	manager := NewManager()
	defer manager.Shutdown()

	assert.Equal(t, StatusUnknown, manager.Status("non-existent-service"), "Status of a non-existent service should be Unknown")

	service := newMockCheckable("test-service", 1*time.Second, nil)
	manager.Start(service)

	// Status should be unknown initially, until the first check completes.
	assert.Equal(t, StatusUnknown, manager.Status("test-service"), "Status should be Unknown before the first check")

	time.Sleep(1100 * time.Millisecond) // Wait for the first check
	assert.Equal(t, StatusHealthy, manager.Status("test-service"), "Status should be Healthy after a successful check")
}

func TestManager_Shutdown(t *testing.T) {
	manager := NewManager()
	service1 := newMockCheckable("service1", 100*time.Millisecond, nil)
	service2 := newMockCheckable("service2", 100*time.Millisecond, nil)

	manager.Start(service1)
	manager.Start(service2)

	time.Sleep(150 * time.Millisecond) // Let checks run

	manager.Shutdown()

	count1 := service1.getCheckCount()
	count2 := service2.getCheckCount()

	time.Sleep(250 * time.Millisecond)

	assert.Equal(t, count1, service1.getCheckCount(), "Check count for service1 should not increase after shutdown")
	assert.Equal(t, count2, service2.getCheckCount(), "Check count for service2 should not increase after shutdown")
	assert.Equal(t, StatusUnknown, manager.Status("service1"), "Status should be Unknown after shutdown")
}

func TestManager_StatusTransition(t *testing.T) {
	manager := NewManager()
	defer manager.Shutdown()

	isHealthy := true
	var mu sync.Mutex
	service := newMockCheckable("transition-service", 100*time.Millisecond, func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()
		if isHealthy {
			return nil
		}
		return fmt.Errorf("now unhealthy")
	})

	manager.Start(service)

	// Initial healthy state
	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, StatusHealthy, manager.Status("transition-service"))

	// Transition to unhealthy
	mu.Lock()
	isHealthy = false
	mu.Unlock()

	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, StatusUnhealthy, manager.Status("transition-service"))

	// Transition back to healthy
	mu.Lock()
	isHealthy = true
	mu.Unlock()

	time.Sleep(150 * time.Millisecond)
	assert.Equal(t, StatusHealthy, manager.Status("transition-service"))
}
