// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHistoryManager(t *testing.T) {
	// Create a dedicated instance for testing to avoid singleton issues
	hm := &HistoryManager{
		history: make(map[string][]Record),
		maxSize: 10, // Small size for testing
	}

	// Test adding records
	serviceName := "test-service"
	hm.AddRecord(serviceName, true)
	hm.AddRecord(serviceName, false)

	history := hm.GetHistory(serviceName)
	assert.Len(t, history, 2)
	assert.True(t, history[0].Status)
	assert.False(t, history[1].Status)

	// Test Pruning
	for i := 0; i < 20; i++ {
		hm.AddRecord(serviceName, true)
	}

	history = hm.GetHistory(serviceName)
	assert.Len(t, history, 10) // Should be capped at maxSize (10)

	// Verify timestamps are roughly correct
	assert.WithinDuration(t, time.Now(), history[9].Timestamp, 1*time.Second)
}
