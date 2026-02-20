// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_GetRecentServiceStats(t *testing.T) {
	m := &Manager{
		trafficHistory: make(map[int64]*MinuteStats),
	}

	now := time.Now()
	minute1 := now.Add(-5 * time.Minute).Truncate(time.Minute).Unix()
	minute2 := now.Add(-10 * time.Minute).Truncate(time.Minute).Unix()
	minuteOld := now.Add(-30 * time.Minute).Truncate(time.Minute).Unix()

	m.trafficHistory[minute1] = &MinuteStats{
		ServiceStats: map[string]*ServiceTrafficStats{
			"service-a": {Requests: 10, Latency: 1000, Errors: 0}, // 100ms avg
			"service-b": {Requests: 5, Latency: 500, Errors: 2},   // 100ms avg
		},
	}
	m.trafficHistory[minute2] = &MinuteStats{
		ServiceStats: map[string]*ServiceTrafficStats{
			"service-a": {Requests: 10, Latency: 2000, Errors: 2}, // 200ms avg
		},
	}
	m.trafficHistory[minuteOld] = &MinuteStats{
		ServiceStats: map[string]*ServiceTrafficStats{
			"service-a": {Requests: 10, Latency: 500, Errors: 0}, // Should be ignored
		},
	}

	// Test Service A (Last 15m)
	// Total Requests: 10 (min1) + 10 (min2) = 20
	// Total Latency: 1000 + 2000 = 3000
	// Avg Latency: 150ms
	// Total Errors: 0 + 2 = 2
	// Error Rate: 2/20 = 0.1
	lat, errRate := m.GetRecentServiceStats("service-a", 15*time.Minute)
	assert.Equal(t, 150*time.Millisecond, lat)
	assert.Equal(t, 0.1, errRate)

	// Test Service B (Last 15m)
	// Total Requests: 5
	// Total Latency: 500
	// Avg Latency: 100ms
	// Total Errors: 2
	// Error Rate: 2/5 = 0.4
	lat, errRate = m.GetRecentServiceStats("service-b", 15*time.Minute)
	assert.Equal(t, 100*time.Millisecond, lat)
	assert.Equal(t, 0.4, errRate)

	// Test Service A (Last 1m) - Should only pick up minute1 (maybe, depending on exact time)
	// Actually truncate ensures alignment.
	// If window is small, we might miss data if now is far from minute1.
	// 5 mins ago.
	lat, errRate = m.GetRecentServiceStats("service-a", 2*time.Minute)
	assert.Equal(t, time.Duration(0), lat)
	assert.Equal(t, 0.0, errRate)
}
