// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestServiceHistory(t *testing.T) {
	h := newServiceHistory()

	// Test adding records
	for i := 0; i < HistorySize+10; i++ {
		h.add(Record{
			Timestamp: time.Now(),
			Status:    "UP",
			LatencyMs: int64(i),
		})
	}

	records := h.get()

	// Check size cap
	assert.Equal(t, HistorySize, len(records))

	// Check that the oldest records were removed (FIFO)
	// The inserted latencies are 0, 1, ..., HistorySize+9
	// We expect records to contain 10, ..., HistorySize+9
	assert.Equal(t, int64(10), records[0].LatencyMs)
	assert.Equal(t, int64(HistorySize+9), records[len(records)-1].LatencyMs)
}

func TestGlobalHistory(t *testing.T) {
	svc := fmt.Sprintf("test-service-%d", time.Now().UnixNano())
	AddRecord(svc, "UP", 100, "")
	AddRecord(svc, "DOWN", 0, "timeout")

	history := GetHistory(svc)
	assert.Len(t, history, 2)
	assert.Equal(t, "UP", history[0].Status)
	assert.Equal(t, "DOWN", history[1].Status)
	assert.Equal(t, "timeout", history[1].Error)

	empty := GetHistory("non-existent")
	assert.Empty(t, empty)
}

func TestConcurrency(t *testing.T) {
	svc := "concurrent-service"
	go func() {
		for i := 0; i < 1000; i++ {
			AddRecord(svc, "UP", int64(i), "")
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			GetHistory(svc)
		}
	}()

	// Just ensure no panic/race detection (run with -race)
	time.Sleep(100 * time.Millisecond)
}
