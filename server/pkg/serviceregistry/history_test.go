// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRingBuffer(t *testing.T) {
	rb := NewRingBuffer(3)

	// Test Empty
	assert.Empty(t, rb.GetAll())

	// Test Add
	item1 := HealthStatus{Timestamp: time.Now(), Status: "ok"}
	rb.Add(item1)
	all := rb.GetAll()
	assert.Len(t, all, 1)
	assert.Equal(t, item1, all[0])

	// Test Fill
	item2 := HealthStatus{Timestamp: time.Now(), Status: "ok"}
	item3 := HealthStatus{Timestamp: time.Now(), Status: "error"}
	rb.Add(item2)
	rb.Add(item3)
	all = rb.GetAll()
	assert.Len(t, all, 3)
	assert.Equal(t, item1, all[0])
	assert.Equal(t, item2, all[1])
	assert.Equal(t, item3, all[2])

	// Test Overflow (Oldest item1 should be removed)
	item4 := HealthStatus{Timestamp: time.Now(), Status: "ok"}
	rb.Add(item4)
	all = rb.GetAll()
	assert.Len(t, all, 3)
	assert.Equal(t, item2, all[0])
	assert.Equal(t, item3, all[1])
	assert.Equal(t, item4, all[2])
}
