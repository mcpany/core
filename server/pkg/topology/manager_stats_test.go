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
	// Seed data
	// 5 minutes ago: 10 reqs, 1000ms total latency (100ms avg), 1 error
	t5 := now.Add(-5 * time.Minute).Unix()
	m.trafficHistory[t5] = &MinuteStats{
		Requests: 10,
		Errors:   1,
		Latency:  1000,
		ServiceStats: map[string]*ServiceTrafficStats{
			"svc-1": {Requests: 10, Errors: 1, Latency: 1000},
		},
	}

	// 20 minutes ago (should be excluded): 10 reqs, 2000ms total latency (200ms avg), 0 errors
	t20 := now.Add(-20 * time.Minute).Unix()
	m.trafficHistory[t20] = &MinuteStats{
		Requests: 10,
		Errors:   0,
		Latency:  2000,
		ServiceStats: map[string]*ServiceTrafficStats{
			"svc-1": {Requests: 10, Errors: 0, Latency: 2000},
		},
	}

	// Test "svc-1" for last 15 minutes
	avgLatency, errorRate := m.GetRecentServiceStats("svc-1", 15*time.Minute)
	assert.Equal(t, 100*time.Millisecond, avgLatency)
	assert.Equal(t, 0.1, errorRate)

	// Test "svc-1" for last 30 minutes
	avgLatency, errorRate = m.GetRecentServiceStats("svc-1", 30*time.Minute)
	// Total: 20 reqs, 3000ms latency, 1 error
	// Avg: 150ms
	// Error rate: 1/20 = 0.05
	assert.Equal(t, 150*time.Millisecond, avgLatency)
	assert.Equal(t, 0.05, errorRate)

	// Test non-existent service
	avgLatency, errorRate = m.GetRecentServiceStats("svc-2", 15*time.Minute)
	assert.Equal(t, time.Duration(0), avgLatency)
	assert.Equal(t, 0.0, errorRate)
}
