// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHistory(t *testing.T) {
	// Reset singleton for testing (not thread safe but okay for unit test if sequential)
	// Or just use the singleton.
	h := GetHistory()
	// Clear existing entries? We can't access entries directly easily due to mutex.
	// But since it's a singleton, previous tests might have polluted it.
	// However, we can just check if our added entries are present at the end.

	initialCount := len(h.Get())

	h.Add("test-service-1", "up")
	time.Sleep(10 * time.Millisecond)
	h.Add("test-service-1", "down")

	entries := h.Get()
	assert.Equal(t, initialCount+2, len(entries))

	// Check the last two entries
	last := entries[len(entries)-1]
	secondLast := entries[len(entries)-2]

	assert.Equal(t, "down", last.Status)
	assert.Equal(t, "test-service-1", last.ServiceName)

	assert.Equal(t, "up", secondLast.Status)
	assert.Equal(t, "test-service-1", secondLast.ServiceName)
}
