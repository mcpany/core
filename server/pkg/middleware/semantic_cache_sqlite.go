// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteVectorStore implements a persistent vector store using SQLite.
// It loads all entries into memory on startup for fast search,
// and persists updates to the database.
type SQLiteVectorStore struct {
	db     *sql.DB
	memory *MemoryVectorStore
	mu     sync.Mutex
}

// NewSQLiteVectorStore creates a new SQLiteVectorStore.
func NewSQLiteVectorStore(path string) (*SQLiteVectorStore, error) {
	if path == "" {
		return nil, fmt.Errorf("sqlite path is required")
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Create table if not exists
	schema := `
	CREATE TABLE IF NOT EXISTS semantic_cache_entries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		vector TEXT NOT NULL,
		result TEXT NOT NULL,
		expires_at INTEGER NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_semantic_cache_key ON semantic_cache_entries(key);
	CREATE INDEX IF NOT EXISTS idx_semantic_cache_expires_at ON semantic_cache_entries(expires_at);
	`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create semantic_cache_entries table: %w", err)
	}

	store := &SQLiteVectorStore{
		db:     db,
		memory: NewMemoryVectorStore(),
	}

	// Load entries into memory
	if err := store.load(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to load cache entries: %w", err)
	}

	// Start background cleanup
	go store.backgroundCleanup()

	return store, nil
}

func (s *SQLiteVectorStore) load() error {
	rows, err := s.db.Query("SELECT key, vector, result, expires_at FROM semantic_cache_entries WHERE expires_at > ?", time.Now().UnixNano())
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var key string
		var vectorJSON, resultJSON string
		var expiresAtNano int64

		if err := rows.Scan(&key, &vectorJSON, &resultJSON, &expiresAtNano); err != nil {
			return err
		}

		var vector []float32
		if err := json.Unmarshal([]byte(vectorJSON), &vector); err != nil {
			continue // Skip invalid entries
		}

		var result any
		if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
			continue // Skip invalid entries
		}

		// Add to memory without persistence (to avoid circular write)
		expiresAt := time.Unix(0, expiresAtNano)
		// We use the internal Add of MemoryVectorStore which is safe
		// Calculate norm since MemoryVectorStore expects it if we bypassed Add,
		// but we can just use Add. MemoryVectorStore.Add does not persist, so it's safe.
		// Wait, MemoryVectorStore.Add takes duration. We have expiration time.
		// We need to calculate remaining TTL or modify MemoryVectorStore to accept absolute time.
		// Or just manually insert.
		s.memory.mu.Lock()
		entries := s.memory.items[key]
		entry := &VectorEntry{
			Vector:    vector,
			Result:    result,
			ExpiresAt: expiresAt,
			Norm:      vectorNorm(vector),
		}
		s.memory.items[key] = append(entries, entry)
		s.memory.mu.Unlock()
	}

	return nil
}

// Add adds an entry to the store (memory + DB).
func (s *SQLiteVectorStore) Add(key string, vector []float32, result any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 1. Add to Memory
	if err := s.memory.Add(key, vector, result, ttl); err != nil {
		return err
	}

	// 2. Add to DB
	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed to marshal vector: %w", err)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	expiresAt := time.Now().Add(ttl).UnixNano()

	_, err = s.db.Exec("INSERT INTO semantic_cache_entries (key, vector, result, expires_at) VALUES (?, ?, ?, ?)",
		key, string(vectorJSON), string(resultJSON), expiresAt)

	return err
}

// Search searches for similar vectors in memory.
func (s *SQLiteVectorStore) Search(key string, query []float32) (any, float32, bool) {
	return s.memory.Search(key, query)
}

// Clear clears the cache.
func (s *SQLiteVectorStore) Clear(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.memory.Clear(ctx); err != nil {
		return err
	}

	_, err := s.db.ExecContext(ctx, "DELETE FROM semantic_cache_entries")
	return err
}

// Close closes the database connection.
func (s *SQLiteVectorStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteVectorStore) backgroundCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now().UnixNano()
		_, _ = s.db.Exec("DELETE FROM semantic_cache_entries WHERE expires_at < ?", now)
		s.mu.Unlock()
	}
}
