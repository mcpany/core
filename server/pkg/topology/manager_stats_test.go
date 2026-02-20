// Copyright 2025 Author(s) of MCP Any
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

	now := time.Now()
	minuteKey := now.Truncate(time.Minute).Unix()
	prevMinuteKey := now.Add(-1 * time.Minute).Truncate(time.Minute).Unix()

	// Seed traffic history
	m.mu.Lock()
	m.trafficHistory[minuteKey] = &MinuteStats{
		Requests: 100,
		Errors:   5,
		Latency:  5000, // 50ms avg
		ServiceStats: map[string]*ServiceTrafficStats{
			"service-a": {
				Requests: 80,
				Errors:   4,
				Latency:  4000, // 50ms avg
			},
			"service-b": {
				Requests: 20,
				Errors:   1,
				Latency:  1000, // 50ms avg
			},
		},
	}
	m.trafficHistory[prevMinuteKey] = &MinuteStats{
		Requests: 50,
		Errors:   0,
		Latency:  2500, // 50ms avg
		ServiceStats: map[string]*ServiceTrafficStats{
			"service-a": {
				Requests: 40,
				Errors:   0,
				Latency:  2000, // 50ms avg
			},
			"service-b": {
				Requests: 10,
				Errors:   0,
				Latency:  500, // 50ms avg
			},
		},
	}
	m.mu.Unlock()

	// Test case 1: Get stats for service-a over last 2 minutes
	avgLat, errRate := m.GetRecentServiceStats("service-a", 2*time.Minute)
	// Total Requests: 80 + 40 = 120
	// Total Errors: 4 + 0 = 4
	// Total Latency: 4000 + 2000 = 6000
	// Avg Latency: 6000 / 120 = 50ms
	// Error Rate: (4 / 120) * 100 = 3.333...%

	assert.Equal(t, 50*time.Millisecond, avgLat)
	assert.InDelta(t, 3.333333, errRate, 0.0001)

	// Test case 2: Get stats for service-b over last 1 minute
	avgLat, errRate = m.GetRecentServiceStats("service-b", 1*time.Minute)
	// Total Requests: 20
	// Total Errors: 1
	// Total Latency: 1000
	// Avg Latency: 50ms
	// Error Rate: (1 / 20) * 100 = 5.0%

	assert.Equal(t, 50*time.Millisecond, avgLat)
	assert.Equal(t, 5.0, errRate)

	// Test case 3: Global stats (empty service ID)
	avgLat, errRate = m.GetRecentServiceStats("", 2*time.Minute)
	// Total Requests: 100 + 50 = 150
	// Total Errors: 5 + 0 = 5
	// Total Latency: 5000 + 2500 = 7500
	// Avg Latency: 7500 / 150 = 50ms
	// Error Rate: (5 / 150) * 100 = 3.333...%

	assert.Equal(t, 50*time.Millisecond, avgLat)
	assert.InDelta(t, 3.333333, errRate, 0.0001)

	// Test case 4: Non-existent service
	avgLat, errRate = m.GetRecentServiceStats("non-existent", 2*time.Minute)
	assert.Equal(t, 0*time.Millisecond, avgLat)
	assert.Equal(t, 0.0, errRate)
}
