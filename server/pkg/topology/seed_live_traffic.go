// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"github.com/mcpany/core/server/pkg/logging"
	"time"
)

// LiveTrafficSeedPoint represents a point for live traffic seeding.
type LiveTrafficSeedPoint struct {
	FromNode string `json:"from_node"`
	ToNode   string `json:"to_node"`
	Status   string `json:"status"`
	Bytes    int64  `json:"bytes"`
	Latency  int64  `json:"latency"`
}

// SeedLiveTraffic allows seeding real-time data flow points.
func (m *Manager) SeedLiveTraffic(points []LiveTrafficSeedPoint) {
	m.mu.Lock()
	defer m.mu.Unlock()

	log := logging.GetLogger()
	log.Info("Seeding live traffic history", "points", len(points))
	now := time.Now()

	// Add dummy sessions for the seeded points.
	for _, p := range points {
		sessionID := "seed-session-" + p.FromNode + "-" + p.ToNode
		if _, exists := m.sessions[sessionID]; !exists {
			m.sessions[sessionID] = &SessionStats{
				ID:             sessionID,
				LastActive:     now,
				RequestCount:   1,
				TotalLatency:   time.Duration(p.Latency) * time.Millisecond,
				ErrorCount:     0,
				TotalBytes:     p.Bytes,
				ServiceCounts:  make(map[string]int64),
				ServiceErrors:  make(map[string]int64),
				ServiceLatency: make(map[string]time.Duration),
			}
		}

		session := m.sessions[sessionID]
		session.RequestCount++
		session.TotalBytes += p.Bytes
		session.TotalLatency += time.Duration(p.Latency) * time.Millisecond
		session.ServiceCounts[p.ToNode]++
		session.ServiceLatency[p.ToNode] += time.Duration(p.Latency) * time.Millisecond

		if p.Status != "ok" {
			session.ErrorCount++
			session.ServiceErrors[p.ToNode]++
		}
	}
}
