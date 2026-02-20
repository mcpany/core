// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/health"
	"github.com/stretchr/testify/assert"
)

func TestCalculateUptime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		history  []health.HistoryPoint
		window   time.Duration
		expected string
	}{
		{
			name:     "Empty history",
			history:  []health.HistoryPoint{},
			window:   24 * time.Hour,
			expected: "0.0%",
		},
		{
			name: "All UP within window",
			history: []health.HistoryPoint{
				{Timestamp: now.Add(-1 * time.Hour).UnixMilli(), Status: "up"},
			},
			window:   2 * time.Hour,
			expected: "100.0%",
		},
		{
			name: "Mixed status",
			history: []health.HistoryPoint{
				{Timestamp: now.Add(-1 * time.Hour).UnixMilli(), Status: "up"},
				{Timestamp: now.Add(-30 * time.Minute).UnixMilli(), Status: "down"},
			},
			window:   1 * time.Hour,
			expected: "50.0%",
		},
		{
			name: "History older than window (All UP)",
			history: []health.HistoryPoint{
				{Timestamp: now.Add(-25 * time.Hour).UnixMilli(), Status: "up"},
			},
			window:   24 * time.Hour,
			expected: "100.0%",
		},
		{
			name: "History older than window (All DOWN)",
			history: []health.HistoryPoint{
				{Timestamp: now.Add(-25 * time.Hour).UnixMilli(), Status: "down"},
			},
			window:   24 * time.Hour,
			expected: "0.0%",
		},
		{
			name: "Complex scenario",
			history: []health.HistoryPoint{
				{Timestamp: now.Add(-2 * time.Hour).UnixMilli(), Status: "up"},
				{Timestamp: now.Add(-1 * time.Hour).UnixMilli(), Status: "down"},
			},
			window:   3 * time.Hour,
			expected: "50.0%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateUptime(tt.history, tt.window, now)
			assert.Equal(t, tt.expected, result)
		})
	}
}
