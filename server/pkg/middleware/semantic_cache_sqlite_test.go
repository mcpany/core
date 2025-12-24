// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteVectorStore(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "sqlite-vector-store-*.db")
	require.NoError(t, err)
	dbPath := tmpFile.Name()
	defer os.Remove(dbPath)
	_ = tmpFile.Close()

	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	key := "test_key"
	vector := []float32{1.0, 0.0, 0.0}
	result := "result"
	ttl := 1 * time.Minute

	// Test Add
	err = store.Add(key, vector, result, ttl)
	assert.NoError(t, err)

	// Test Search (Hit)
	// Exact vector should yield score 1.0 (or very close)
	res, score, found := store.Search(key, vector)
	assert.True(t, found)
	assert.InDelta(t, 1.0, score, 0.001)
	assert.Equal(t, result, res)

	// Test Search (Miss)
	_, _, found = store.Search("other_key", vector)
	assert.False(t, found)

	// Test Persistence
	// Close and reopen
	store.Close()

	store2, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Verify data loaded from DB
	res, score, found = store2.Search(key, vector)
	assert.True(t, found)
	assert.InDelta(t, 1.0, score, 0.001)
	assert.Equal(t, result, res)
}

func TestSQLiteVectorStore_Expiry(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "sqlite-vector-store-expiry-*.db")
	require.NoError(t, err)
	dbPath := tmpFile.Name()
	defer os.Remove(dbPath)
	_ = tmpFile.Close()

	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	key := "test_key"
	vector := []float32{1.0, 0.0, 0.0}
	result := "result"
	ttl := 1 * time.Millisecond // Short TTL

	err = store.Add(key, vector, result, ttl)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// In-memory check should filter expired
	_, _, found := store.Search(key, vector)
	assert.False(t, found)

	// Reopen to check if expired entries are loaded
	store.Close()
	store2, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Should not be loaded
	_, _, found = store2.Search(key, vector)
	assert.False(t, found)
}
