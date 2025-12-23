// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vectorstore_test

import (
	"testing"
	"time"

	"github.com/mcpany/core/pkg/vectorstore"
	"github.com/stretchr/testify/assert"
)

func TestSimpleStore(t *testing.T) {
	s := vectorstore.NewSimpleStore(100)

	v1 := []float32{1, 0, 0}
	v2 := []float32{0, 1, 0}
	v3 := []float32{0.707, 0.707, 0} // ~45 deg between v1 and v2

	s.Add(v1, "v1", time.Hour)
	s.Add(v2, "v2", time.Hour)

	// Search close to v1
	results, err := s.Search([]float32{0.9, 0.1, 0}, 1, 0.5)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "v1", results[0].Data)

	// Search close to v3 (should match both v1 and v2 with sim ~0.707)
	results, err = s.Search(v3, 2, 0.5)
	assert.NoError(t, err)
	assert.Len(t, results, 2)
}

func TestSimpleStore_Eviction(t *testing.T) {
	s := vectorstore.NewSimpleStore(2) // Capacity 2

	v1 := []float32{1, 0}
	v2 := []float32{0, 1}
	v3 := []float32{1, 1}

	s.Add(v1, "v1", time.Hour)
	s.Add(v2, "v2", time.Hour)
	// Full now.

	// Add v3. Should evict v1 (FIFO).
	s.Add(v3, "v3", time.Hour)

	// Check
	res, _ := s.Search(v1, 1, 0.99)
	assert.Len(t, res, 0, "v1 should be evicted")

	res, _ = s.Search(v2, 1, 0.99)
	assert.Len(t, res, 1, "v2 should remain")

	res, _ = s.Search(v3, 1, 0.99)
	assert.Len(t, res, 1, "v3 should exist")
}

func TestSimpleStore_Expiration(t *testing.T) {
	s := vectorstore.NewSimpleStore(10)

	v1 := []float32{1, 0}
	s.Add(v1, "v1", 1 * time.Millisecond)

	time.Sleep(10 * time.Millisecond)

	res, _ := s.Search(v1, 1, 0.1)
	assert.Len(t, res, 0, "v1 should be expired")
}

func TestSimpleStore_Eviction_PrioritizesExpired(t *testing.T) {
	s := vectorstore.NewSimpleStore(2)

	v1 := []float32{1, 0}
	v2 := []float32{0, 1}
	v3 := []float32{1, 1}

	s.Add(v1, "v1", 1 * time.Millisecond) // Will expire soon
	s.Add(v2, "v2", time.Hour)            // Won't expire

	time.Sleep(10 * time.Millisecond)
	// v1 expired

	// Add v3. Should evict v1 because it's expired
	s.Add(v3, "v3", time.Hour)

	// v1 gone (expired/evicted)
	res, _ := s.Search(v1, 1, 0.99)
	assert.Len(t, res, 0)

	// v2 should remain
	res, _ = s.Search(v2, 1, 0.99)
	assert.Len(t, res, 1)

	// v3 should exist
	res, _ = s.Search(v3, 1, 0.99)
	assert.Len(t, res, 1)
}
