// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteVectorStore(t *testing.T) {
	// Setup temporary DB
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vectors.db")

	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	key := "tool_a"
	vec := []float32{1.0, 0.0}
	result := map[string]interface{}{"foo": "bar"}
	ttl := 1 * time.Minute

	// Add entry
	err = store.Add(context.Background(), key, vec, result, ttl)
	assert.NoError(t, err)

	// Search in same session (should be in memory)
	res, score, found := store.Search(context.Background(), key, []float32{1.0, 0.0})
	assert.True(t, found)
	assert.Equal(t, float32(1.0), score)
	assert.Equal(t, result, res)

	// Close and reopen to test persistence
	err = store.Close()
	assert.NoError(t, err)

	store2, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Search in new session (should be loaded from DB)
	res, score, found = store2.Search(context.Background(), key, []float32{1.0, 0.0})
	assert.True(t, found)
	assert.Equal(t, float32(1.0), score)

	// Check result content - we need to handle map[string]interface{} equality carefully if JSON marshaling changes types (e.g. numbers)
	// result is {"foo": "bar"}, which should be stable.
	assert.Equal(t, result, res)
}

func TestSQLiteVectorStore_Expiry(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vectors_expiry.db")

	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	key := "tool_b"
	vec := []float32{0.0, 1.0}
	result := "expired"
	ttl := 1 * time.Millisecond

	err = store.Add(context.Background(), key, vec, result, ttl)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Search should miss in memory (memory store handles expiry check)
	_, _, found := store.Search(context.Background(), key, vec)
	assert.False(t, found)

	// Close and reopen
	store.Close()

	store2, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Should not load expired item
	_, _, found = store2.Search(context.Background(), key, vec)
	assert.False(t, found)
}

func TestSQLiteVectorStore_InvalidPath(t *testing.T) {
	_, err := NewSQLiteVectorStore("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "sqlite path is required")

	// Test path that cannot be created (e.g. directory that doesn't exist)
	// /nonexistent/db.sqlite
	_, err = NewSQLiteVectorStore("/nonexistent/dir/db.sqlite")
	assert.Error(t, err)
}

func TestSQLiteVectorStore_CorruptedData(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vectors_corrupt.db")

	// Create valid DB first
	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)

	key := "tool_c"
	vec := []float32{0.5, 0.5}
	result := "valid"
	ttl := 1 * time.Hour

	err = store.Add(context.Background(), key, vec, result, ttl)
	assert.NoError(t, err)
	store.Close()

	// Corrupt the file? Or insert garbage via raw SQL?
	// Can't easily use raw SQL without opening it again.
	// Let's overwrite with garbage.
	// Actually, if we overwrite with garbage, NewSQLiteVectorStore might fail to open or query.

	// Let's try to simulate bad JSON in DB by manually inserting if possible,
	// or just test that loadFromDB handles scan errors if we drop the table?
	// Dropping table would cause query error.

	// Better test: Insert invalid JSON into the vector/result column using a raw connection.
	// But we need to use the same driver.

	store2, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store2.Close()

	// Insert invalid vector data (not multiple of 4 bytes)
	_, err = store2.db.Exec("INSERT INTO semantic_cache_entries (key, vector, result, expires_at) VALUES (?, ?, ?, ?)",
		"bad_vec", []byte{0x01, 0x02, 0x03}, "{}", time.Now().Add(time.Hour).UnixNano())
	assert.NoError(t, err)

	// Insert invalid result JSON
	vectorBytes := float32ToBytes([]float32{0.0})
	_, err = store2.db.Exec("INSERT INTO semantic_cache_entries (key, vector, result, expires_at) VALUES (?, ?, ?, ?)",
		"bad_res", vectorBytes, "{invalid_json}", time.Now().Add(time.Hour).UnixNano())
	assert.NoError(t, err)

	store2.Close()

	// Reopen and check that it doesn't crash and loads valid entries
	store3, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store3.Close()

	// Should find the original valid entry
	res, _, found := store3.Search(context.Background(), key, vec)
	assert.True(t, found)
	assert.Equal(t, result, res)

	// Should NOT find bad_vec (loadFromDB skips it)
	res, _, found = store3.Search(context.Background(), "bad_vec", []float32{0.0})
	assert.False(t, found)

	// Should NOT find bad_res
	res, _, found = store3.Search(context.Background(), "bad_res", []float32{0.0})
	assert.False(t, found)
}

func TestSQLiteVectorStore_Prune(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "vectors_prune.db")

	store, err := NewSQLiteVectorStore(dbPath)
	require.NoError(t, err)
	defer store.Close()

	// Add entry that expires immediately
	key := "tool_prune"
	vec := []float32{0.1}
	store.Add(context.Background(), key, vec, "data", 1*time.Millisecond)

	time.Sleep(10 * time.Millisecond)

	// Manual Prune
	store.Prune(context.Background(), key)

	// Check DB directly
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM semantic_cache_entries").Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}
