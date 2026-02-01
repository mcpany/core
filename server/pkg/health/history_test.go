// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHealthHistory(t *testing.T) {
	// Reset store
	historyMu.Lock()
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu.Unlock()

	// First call triggers seeding (~1440 points)
	AddHistoryPoint("svc1", "healthy", 10)
	AddHistoryPoint("svc1", "unhealthy", 20)

	hist := GetHealthHistory()
	// Seeding (24h * 60m = 1440 points) + 2 new points = 1442
	// But `now` advances slightly, so let's just check length is reasonable.
	assert.Greater(t, len(hist["svc1"]), 1000)

	l := len(hist["svc1"])
	assert.Equal(t, "healthy", hist["svc1"][l-2].Status)
	assert.Equal(t, int64(10), hist["svc1"][l-2].LatencyMs)
	assert.Equal(t, "unhealthy", hist["svc1"][l-1].Status)
	assert.Equal(t, int64(20), hist["svc1"][l-1].LatencyMs)

	// Test pruning
	for i := 0; i < 21000; i++ {
		AddHistoryPoint("svc2", "ok", 5)
	}
	hist = GetHealthHistory()
	assert.Len(t, hist["svc2"], 20000)
}
