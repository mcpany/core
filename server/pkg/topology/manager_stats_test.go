// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_GetRecentServiceStats(t *testing.T) {
	mockRegistry := new(MockServiceRegistry)
	mockTM := new(MockToolManager)
	m := NewManager(mockRegistry, mockTM)
	defer m.Close()

	// Helper to manually inject stats
	injectStats := func(offsetMinutes int, reqs, errors, latency int64, serviceID string) {
		tTime := time.Now().Add(time.Duration(-offsetMinutes) * time.Minute).Truncate(time.Minute)
		key := tTime.Unix()

		m.mu.Lock()
		defer m.mu.Unlock()

		if _, ok := m.trafficHistory[key]; !ok {
			m.trafficHistory[key] = &MinuteStats{
				ServiceStats: make(map[string]*ServiceTrafficStats),
			}
		}
		stats := m.trafficHistory[key]
		stats.Requests += reqs
		stats.Errors += errors
		stats.Latency += latency

		if serviceID != "" {
			if _, ok := stats.ServiceStats[serviceID]; !ok {
				stats.ServiceStats[serviceID] = &ServiceTrafficStats{}
			}
			sStats := stats.ServiceStats[serviceID]
			sStats.Requests += reqs
			sStats.Errors += errors
			sStats.Latency += latency
		}
	}

	// Case 1: No data
	latency, errRate := m.GetRecentServiceStats("service-a", 15*time.Minute)
	assert.Equal(t, time.Duration(0), latency)
	assert.Equal(t, 0.0, errRate)

	// Case 2: Data within window
	// Minute 0 (Now): 10 reqs, 1 error, 100ms total latency (10ms avg)
	injectStats(0, 10, 1, 100, "service-a")
	// Minute 5: 10 reqs, 0 errors, 200ms total latency (20ms avg)
	injectStats(5, 10, 0, 200, "service-a")

	// Total: 20 reqs, 1 error, 300ms total -> 15ms avg, 0.05 error rate
	latency, errRate = m.GetRecentServiceStats("service-a", 15*time.Minute)
	assert.Equal(t, 15*time.Millisecond, latency)
	assert.Equal(t, 0.05, errRate)

	// Case 3: Data outside window
	// Minute 20: 100 reqs, 0 errors, 1000ms latency. Should be ignored.
	injectStats(20, 100, 0, 1000, "service-a")

	latency, errRate = m.GetRecentServiceStats("service-a", 15*time.Minute)
	assert.Equal(t, 15*time.Millisecond, latency) // Unchanged
	assert.Equal(t, 0.05, errRate)                 // Unchanged

	// Case 4: Other service data mixed in
	// Minute 1: Service B: 10 reqs, 5 errors, 500ms
	injectStats(1, 10, 5, 500, "service-b")

	latency, errRate = m.GetRecentServiceStats("service-a", 15*time.Minute)
	assert.Equal(t, 15*time.Millisecond, latency) // Unchanged for Service A
	assert.Equal(t, 0.05, errRate)                 // Unchanged for Service A

	// Verify Service B
	latency, errRate = m.GetRecentServiceStats("service-b", 15*time.Minute)
	assert.Equal(t, 50*time.Millisecond, latency)
	assert.Equal(t, 0.5, errRate)

	// Verify Global (no service ID)
	// Service A: 20 reqs, 1 error, 300ms
	// Service B: 10 reqs, 5 errors, 500ms
	// Total: 30 reqs, 6 errors, 800ms -> 26.66ms avg, 0.2 error rate
	latency, errRate = m.GetRecentServiceStats("", 15*time.Minute)
	assert.Equal(t, 26*time.Millisecond, latency.Truncate(time.Millisecond))
	assert.Equal(t, 0.2, errRate)
}
