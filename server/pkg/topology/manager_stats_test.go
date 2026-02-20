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

	// Seed some data in trafficHistory
	now := time.Now()
	minute1 := now.Truncate(time.Minute).Unix()
	minute2 := now.Add(-5 * time.Minute).Truncate(time.Minute).Unix()
	minuteOld := now.Add(-30 * time.Minute).Truncate(time.Minute).Unix()

	m.mu.Lock()
	m.trafficHistory[minute1] = &MinuteStats{
		Requests: 10,
		Errors:   1,
		Latency:  1000, // 100ms avg
		ServiceStats: map[string]*ServiceTrafficStats{
			"svc-a": {Requests: 10, Errors: 1, Latency: 1000},
		},
	}
	m.trafficHistory[minute2] = &MinuteStats{
		Requests: 20,
		Errors:   0,
		Latency:  4000, // 200ms avg
		ServiceStats: map[string]*ServiceTrafficStats{
			"svc-a": {Requests: 20, Errors: 0, Latency: 4000},
		},
	}
	// Old data, should be ignored for 15m window
	m.trafficHistory[minuteOld] = &MinuteStats{
		Requests: 100,
		Errors:   50,
		Latency:  10000,
		ServiceStats: map[string]*ServiceTrafficStats{
			"svc-a": {Requests: 100, Errors: 50, Latency: 10000},
		},
	}
	m.mu.Unlock()

	// Test with 15 minute window
	avgLat, errRate := m.GetRecentServiceStats("svc-a", 15*time.Minute)

	// Total requests in window: 10 + 20 = 30
	// Total latency: 1000 + 4000 = 5000
	// Avg latency: 5000 / 30 = 166.66... ms
	// Total errors: 1 + 0 = 1
	// Error rate: 1 / 30 = 0.0333...

	expectedLat := time.Duration(166) * time.Millisecond
	assert.InDelta(t, float64(expectedLat), float64(avgLat), float64(time.Millisecond))
	assert.InDelta(t, 0.0333, errRate, 0.001)

	// Test with global stats (empty serviceID)
	// Same data since only svc-a exists
	avgLatGlobal, errRateGlobal := m.GetRecentServiceStats("", 15*time.Minute)
	assert.Equal(t, avgLat, avgLatGlobal)
	assert.Equal(t, errRate, errRateGlobal)

	// Test with unknown service
	avgLatUnknown, errRateUnknown := m.GetRecentServiceStats("unknown", 15*time.Minute)
	assert.Equal(t, time.Duration(0), avgLatUnknown)
	assert.Equal(t, 0.0, errRateUnknown)
}
