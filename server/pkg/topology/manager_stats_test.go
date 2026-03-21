// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_GetRecentServiceStats(t *testing.T) {
	m := NewManager(nil, nil)
	defer m.Close()

	// Seed some data
	now := time.Now()
	minuteKey := now.Truncate(time.Minute).Unix()

	m.mu.Lock()
	m.trafficHistory[minuteKey] = &MinuteStats{
		Requests: 10,
		Errors:   1,
		Latency:  1000, // 100ms avg
		ServiceStats: map[string]*ServiceTrafficStats{
			"svc-1": {
				Requests: 5,
				Errors:   0,
				Latency:  250, // 50ms avg
			},
			"svc-2": {
				Requests: 5,
				Errors:   1,
				Latency:  750, // 150ms avg
			},
		},
	}
	m.mu.Unlock()

	// Test specific service
	avgLatency, errorRate := m.GetRecentServiceStats("svc-1", 15*time.Minute)
	assert.Equal(t, 50*time.Millisecond, avgLatency)
	assert.Equal(t, 0.0, errorRate)

	avgLatency, errorRate = m.GetRecentServiceStats("svc-2", 15*time.Minute)
	assert.Equal(t, 150*time.Millisecond, avgLatency)
	assert.Equal(t, 0.2, errorRate)

	// Test global aggregation
	avgLatency, errorRate = m.GetRecentServiceStats("", 15*time.Minute)
	assert.Equal(t, 100*time.Millisecond, avgLatency)
	assert.Equal(t, 0.1, errorRate)
}
