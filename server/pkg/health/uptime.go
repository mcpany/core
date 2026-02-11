// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"strings"
	"time"
)

// CalculateUptime calculates the uptime percentage for a service over a given window.
// It uses the in-memory history of status changes.
func CalculateUptime(serviceName string, window time.Duration) float64 {
	historyMu.RLock()
	hist, ok := historyStore[serviceName]
	historyMu.RUnlock()

	if !ok || len(hist.Points) == 0 {
		// No history means we don't know.
		// If we return 0, it looks bad. If 100, it looks good.
		// Since we assume services are healthy by default until checked, 100 is a safe bet for "no data yet".
		return 100.0
	}

	now := time.Now().UnixMilli()
	windowMillis := window.Milliseconds()
	windowStart := now - windowMillis

	points := hist.Points

	// Find the first point within the window
	firstInWindowIdx := -1
	for i, p := range points {
		if p.Timestamp >= windowStart {
			firstInWindowIdx = i
			break
		}
	}

	var effectiveStart int64
	var currentStatus string

	if firstInWindowIdx == -1 {
		// All points are before the window.
		// The state at windowStart is the state of the last point.
		lastPoint := points[len(points)-1]
		effectiveStart = windowStart
		currentStatus = lastPoint.Status
	} else if firstInWindowIdx == 0 {
		// All points are inside the window.
		// We can only calculate from the first known point.
		effectiveStart = points[0].Timestamp
		currentStatus = points[0].Status
	} else {
		// Some points before, some inside.
		// The state at windowStart is determined by the point just before firstInWindowIdx.
		effectiveStart = windowStart
		currentStatus = points[firstInWindowIdx-1].Status
	}

	// If effectiveStart is effectively now (no duration), return 100 or current status?
	if now <= effectiveStart {
		if isStatusUp(currentStatus) {
			return 100.0
		}
		return 0.0
	}

	var upTime int64 = 0
	cursor := effectiveStart

	// Iterate through points strictly inside the window (from firstInWindowIdx onwards)
	startIterIdx := firstInWindowIdx
	if firstInWindowIdx == -1 {
		startIterIdx = len(points) // Loop won't run
	}

	for i := startIterIdx; i < len(points); i++ {
		p := points[i]

		// Add duration of previous segment
		segmentDuration := p.Timestamp - cursor
		if isStatusUp(currentStatus) {
			upTime += segmentDuration
		}

		cursor = p.Timestamp
		currentStatus = p.Status
	}

	// Add final segment from last point to now
	finalSegment := now - cursor
	if isStatusUp(currentStatus) {
		upTime += finalSegment
	}

	totalTime := now - effectiveStart
	if totalTime == 0 {
		return 100.0
	}

	return (float64(upTime) / float64(totalTime)) * 100.0
}

func isStatusUp(status string) bool {
	s := strings.ToLower(status)
	return s == "up" || s == "healthy" || s == "serving"
}
