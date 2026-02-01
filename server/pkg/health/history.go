// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"math"
	"math/rand"
	"sync"
	"time"
)

// HistoryPoint represents a single point in time for a service's health.
type HistoryPoint struct {
	Timestamp int64  `json:"timestamp"` // Unix millis
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms"`
}

// ServiceHealthHistory stores the history for a service.
type ServiceHealthHistory struct {
	Points []HistoryPoint
}

var (
	historyStore = make(map[string]*ServiceHealthHistory)
	historyMu    sync.RWMutex
)

// AddHistoryPoint adds a status point to the history.
func AddHistoryPoint(serviceName string, status string, latencyMs int64) {
	historyMu.Lock()
	defer historyMu.Unlock()

	hist, ok := historyStore[serviceName]
	if !ok {
		hist = &ServiceHealthHistory{Points: make([]HistoryPoint, 0, 20000)}
		historyStore[serviceName] = hist

		// Lazy seeding if empty (and we are creating it)
		// We call the seed function here.
		seedHistoryLocked(hist)
	}

	// Just append the point.
	now := time.Now().UnixMilli()
	hist.Points = append(hist.Points, HistoryPoint{
		Timestamp: now,
		Status:    status,
		LatencyMs: latencyMs,
	})

	// Prune
	if len(hist.Points) > 20000 {
		hist.Points = hist.Points[len(hist.Points)-20000:]
	}
}

// seedHistoryLocked populates the history with dummy data for the last 24 hours.
func seedHistoryLocked(hist *ServiceHealthHistory) {
	now := time.Now()
	start := now.Add(-24 * time.Hour)

	// Seed data every minute
	for t := start; t.Before(now); t = t.Add(1 * time.Minute) {
		// Simulate latency with a sine wave + noise
		// Period of 6 hours
		noise := rand.Float64() * 20
		baseLatency := 50.0
		latency := baseLatency + 20*math.Sin(float64(t.Unix())/10000.0) + noise

		// Simulate random errors (99.9% uptime)
		status := "OK"
		if rand.Float64() > 0.999 {
			status = "ERROR"
			latency = 0
		} else if rand.Float64() > 0.99 {
			status = "DEGRADED" // Use string to match typical status
			latency += 100 // Spike
		}

		hist.Points = append(hist.Points, HistoryPoint{
			Timestamp: t.UnixMilli(),
			Status:    status,
			LatencyMs: int64(latency),
		})
	}
}

// GetHealthHistory returns the history for all services.
func GetHealthHistory() map[string][]HistoryPoint {
	historyMu.RLock()
	defer historyMu.RUnlock()

	result := make(map[string][]HistoryPoint)
	for name, hist := range historyStore {
		points := make([]HistoryPoint, len(hist.Points))
		copy(points, hist.Points)
		result[name] = points
	}
	return result
}
