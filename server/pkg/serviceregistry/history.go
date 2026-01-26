// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"sync"
	"time"
)

// HealthStatus represents the result of a single health check.
type HealthStatus struct {
	Timestamp time.Time `json:"timestamp"`
	Status    string    `json:"status"` // "ok", "error"
	Error     string    `json:"error,omitempty"`
	Latency   string    `json:"latency"` // duration string
}

// RingBuffer is a fixed-size buffer that overwrites the oldest elements when full.
// It is thread-safe.
type RingBuffer struct {
	mu       sync.RWMutex
	data     []HealthStatus
	head     int
	capacity int
	size     int
}

// NewRingBuffer creates a new RingBuffer with the specified capacity.
func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{
		data:     make([]HealthStatus, capacity),
		capacity: capacity,
	}
}

// Add adds a new element to the buffer.
func (rb *RingBuffer) Add(item HealthStatus) {
	rb.mu.Lock()
	defer rb.mu.Unlock()

	rb.data[rb.head] = item
	rb.head = (rb.head + 1) % rb.capacity
	if rb.size < rb.capacity {
		rb.size++
	}
}

// GetAll returns all elements in the buffer, ordered from oldest to newest.
func (rb *RingBuffer) GetAll() []HealthStatus {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	result := make([]HealthStatus, 0, rb.size)

	if rb.size < rb.capacity {
		for i := 0; i < rb.size; i++ {
			result = append(result, rb.data[i])
		}
	} else {
		// Buffer is full, start from head (which is the oldest overwritten index, so head is now the oldest)
		// Wait, if head is 0, oldest is 0?
		// Example: Cap 3.
		// Add A: [A, _, _], head=1, size=1. All: [A]
		// Add B: [A, B, _], head=2, size=2. All: [A, B]
		// Add C: [A, B, C], head=0, size=3. All: [A, B, C]
		// Add D: [D, B, C], head=1, size=3. Oldest is B (at 1). head points to 1.
		// So iterate from head to capacity-1, then 0 to head-1.

		for i := rb.head; i < rb.capacity; i++ {
			result = append(result, rb.data[i])
		}
		for i := 0; i < rb.head; i++ {
			result = append(result, rb.data[i])
		}
	}

	return result
}
