// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateUptime(t *testing.T) {
	now := time.Now().UnixMilli()
	window := 1 * time.Hour

	tests := []struct {
		name     string
		points   []HistoryPoint
		expected float64 // approx
	}{
		{
			name:     "Empty",
			points:   []HistoryPoint{},
			expected: 1.0,
		},
		{
			name: "Always Up",
			points: []HistoryPoint{
				{Timestamp: now - 3600*1000, Status: "up"},
			},
			expected: 1.0,
		},
		{
			name: "Always Down",
			points: []HistoryPoint{
				{Timestamp: now - 3600*1000, Status: "down"},
			},
			expected: 0.0,
		},
		{
			name: "Mixed",
			points: []HistoryPoint{
				{Timestamp: now - 3600*1000, Status: "down"},
				{Timestamp: now - 2700*1000, Status: "up"},
				{Timestamp: now - 1800*1000, Status: "down"},
				{Timestamp: now - 900*1000, Status: "up"},
			},
			// Intervals:
			// -3600 -> -2700: Down (900s)
			// -2700 -> -1800: Up (900s)
			// -1800 -> -900: Down (900s)
			// -900 -> 0: Up (900s)
			// Total Up: 1800s. Total Window: 3600s.
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateUptime(tt.points, window)
			assert.InDelta(t, tt.expected, got, 0.01)
		})
	}
}
