// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteVectorStore implements VectorStore using SQLite for persistence
// and an in-memory cache for fast search.
type SQLiteVectorStore struct {
	memoryStore *SimpleVectorStore
	db          *sql.DB
}

// NewSQLiteVectorStore creates a new SQLiteVectorStore.
// It loads existing entries from the database into memory.
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
		key TEXT,
		vector TEXT,
		result TEXT,
		expires_at INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_key ON semantic_cache_entries(key);
	CREATE INDEX IF NOT EXISTS idx_expires_at ON semantic_cache_entries(expires_at);
	`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create semantic_cache_entries table: %w", err)
	}

	store := &SQLiteVectorStore{
		memoryStore: NewSimpleVectorStore(),
		db:          db,
	}

	if err := store.loadFromDB(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to load cache entries from db: %w", err)
	}

	return store, nil
}

// loadFromDB loads unexpired entries from DB to memory.
func (s *SQLiteVectorStore) loadFromDB() error {
	now := time.Now().UnixNano()
	// Order by ID ASC to maintain insertion order, helping SimpleVectorStore's FIFO eviction policy
	// work consistently with the persistent state.
	rows, err := s.db.Query("SELECT key, vector, result, expires_at FROM semantic_cache_entries WHERE expires_at > ? ORDER BY id ASC", now)
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
			continue // Skip malformed
		}

		var result any
		if err := json.Unmarshal([]byte(resultJSON), &result); err != nil {
			continue // Skip malformed
		}

		ttl := time.Duration(expiresAtNano - time.Now().UnixNano())
		if ttl > 0 {
			// Add to memory store without writing back to DB
			_ = s.memoryStore.Add(key, vector, result, ttl)
		}
	}

	return nil
}

// Add adds a new entry to both memory and DB.
func (s *SQLiteVectorStore) Add(key string, vector []float32, result any, ttl time.Duration) error {
	// Add to memory first
	if err := s.memoryStore.Add(key, vector, result, ttl); err != nil {
		return err
	}

	// Add to DB
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
	if err != nil {
		return err
	}

	// Probabilistic pruning (1 in 100 chance) to prevent unbound growth
	// without impacting performance on every write.
	if time.Now().UnixNano()%100 == 0 {
		go func() {
			// Best effort prune
			now := time.Now().UnixNano()
			_, _ = s.db.Exec("DELETE FROM semantic_cache_entries WHERE expires_at <= ?", now)
		}()
	}

	return nil
}

// Search searches in memory.
func (s *SQLiteVectorStore) Search(key string, query []float32) (any, float32, bool) {
	return s.memoryStore.Search(key, query)
}

// Prune removes expired entries from both memory and DB.
func (s *SQLiteVectorStore) Prune(key string) {
	s.memoryStore.Prune(key)

	now := time.Now().UnixNano()
	_, _ = s.db.Exec("DELETE FROM semantic_cache_entries WHERE expires_at <= ?", now)
}

// Close closes the database connection.
func (s *SQLiteVectorStore) Close() error {
	return s.db.Close()
}
