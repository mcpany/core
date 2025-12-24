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

// SQLiteVectorStore implements VectorStore using SQLite.
// Since SQLite doesn't have native vector support without extensions,
// and we want to avoid complex CGO dependencies for the basic setup,
// we will store vectors as JSON or BLOBs and perform the cosine similarity
// search in Go by fetching candidates.
// To optimize, we can rely on the fact that we query by 'key' (service/tool identifier),
// which limits the search space significantly.
type SQLiteVectorStore struct {
	db   *sql.DB
	mu   sync.RWMutex
	done chan struct{}
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
	CREATE TABLE IF NOT EXISTS semantic_cache (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key TEXT NOT NULL,
		vector BLOB NOT NULL,
		result TEXT NOT NULL,
		norm REAL NOT NULL,
		expires_at DATETIME NOT NULL
	);
	CREATE INDEX IF NOT EXISTS idx_semantic_cache_key ON semantic_cache(key);
	CREATE INDEX IF NOT EXISTS idx_semantic_cache_expires_at ON semantic_cache(expires_at);
	`
	if _, err := db.Exec(schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create semantic_cache table: %w", err)
	}

	store := &SQLiteVectorStore{
		db:   db,
		done: make(chan struct{}),
	}

	// Start background cleanup
	go store.backgroundCleanup()

	return store, nil
}

// Add adds a vector and its associated result to the store.
func (s *SQLiteVectorStore) Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Serialize vector
	vectorBytes, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed to marshal vector: %w", err)
	}

	// Serialize result
	resultBytes, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	norm := vectorNorm(vector)
	expiresAt := time.Now().Add(ttl)

	query := `
	INSERT INTO semantic_cache (key, vector, result, norm, expires_at)
	VALUES (?, ?, ?, ?, ?)
	`
	_, err = s.db.ExecContext(ctx, query, key, vectorBytes, resultBytes, norm, expiresAt)
	return err
}

// Search finds the nearest neighbor for the given query vector.
func (s *SQLiteVectorStore) Search(ctx context.Context, key string, query []float32) (any, float32, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Fetch all valid entries for the key
	rows, err := s.db.QueryContext(ctx, "SELECT vector, result, norm FROM semantic_cache WHERE key = ? AND expires_at > ?", key, time.Now())
	if err != nil {
		return nil, 0, false, fmt.Errorf("failed to query cache: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var bestResult any
	var bestScore float32 = -1.0
	queryNorm := vectorNorm(query)
	found := false

	for rows.Next() {
		var vectorBytes []byte
		var resultBytes []byte
		var norm float32

		if err := rows.Scan(&vectorBytes, &resultBytes, &norm); err != nil {
			continue
		}

		var vector []float32
		if err := json.Unmarshal(vectorBytes, &vector); err != nil {
			continue
		}

		score := cosineSimilarityOptimized(query, vector, queryNorm, norm)
		if score > bestScore {
			bestScore = score
			var result any
			if err := json.Unmarshal(resultBytes, &result); err == nil {
				bestResult = result
				found = true
			}
		}
	}

	return bestResult, bestScore, found, nil
}

// Close closes the database connection.
func (s *SQLiteVectorStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	close(s.done)
	return s.db.Close()
}

func (s *SQLiteVectorStore) backgroundCleanup() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.cleanup()
		case <-s.done:
			return
		}
	}
}

func (s *SQLiteVectorStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.db.Exec("DELETE FROM semantic_cache WHERE expires_at < ?", time.Now())
}
