// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	// modernc.org/sqlite is a pure Go SQLite driver.
	_ "modernc.org/sqlite"
)

// SQLiteVectorStore implements VectorStore using SQLite for persistence
// and an in-memory cache for fast search.
type SQLiteVectorStore struct {
	memoryStore *SimpleVectorStore
	db          *sql.DB
}

// NewSQLiteVectorStore creates a new SQLiteVectorStore.
//
// Summary: Initializes a SQLite-backed vector store, loading persisted entries into memory for fast access.
//
// Parameters:
//   - path: string. The path to the SQLite database file.
//
// Returns:
//   - *SQLiteVectorStore: The initialized store.
//   - error: An error if database creation or loading fails.
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
		vector BLOB,
		result TEXT,
		expires_at INTEGER
	);
	CREATE INDEX IF NOT EXISTS idx_key ON semantic_cache_entries(key);
	CREATE INDEX IF NOT EXISTS idx_expires_at ON semantic_cache_entries(expires_at);
	`
	ctxSchema, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()
	if _, err := db.ExecContext(ctxSchema, schema); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create semantic_cache_entries table: %w", err)
	}

	// Optimize SQLite performance
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode=WAL;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set WAL mode: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA synchronous=NORMAL;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set synchronous mode: %w", err)
	}
	if _, err := db.ExecContext(ctx, "PRAGMA busy_timeout=5000;"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	store := &SQLiteVectorStore{
		memoryStore: NewSimpleVectorStore(),
		db:          db,
	}

	if err := store.loadFromDB(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to load cache entries from db: %w", err)
	}

	return store, nil
}

// loadFromDB loads unexpired entries from DB to memory.
func (s *SQLiteVectorStore) loadFromDB(ctx context.Context) error {
	now := time.Now().UnixNano()
	// Order by ID ASC to maintain insertion order, helping SimpleVectorStore's FIFO eviction policy
	// work consistently with the persistent state.
	// Order by ID ASC to maintain insertion order, helping SimpleVectorStore's FIFO eviction policy
	// work consistently with the persistent state.
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, "SELECT key, vector, result, expires_at FROM semantic_cache_entries WHERE expires_at > ? ORDER BY id ASC", now)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var key string
		var vectorRaw []byte // Can be JSON string (legacy) or binary (new)
		var resultJSON []byte
		var expiresAtNano int64

		// We scan vector into []byte to handle both TEXT and BLOB column types
		// Scanning resultJSON into []byte avoids allocating a string if it's large.
		if err := rows.Scan(&key, &vectorRaw, &resultJSON, &expiresAtNano); err != nil {
			return err
		}

		var vector []float32

		// Handle legacy JSON and new Binary formats.
		// Heuristic: Check if it looks like JSON.
		isJSON := len(vectorRaw) > 0 && vectorRaw[0] == '['
		if isJSON {
			if err := json.Unmarshal(vectorRaw, &vector); err != nil {
				// If JSON parsing fails, it might be a binary blob that happened to start with '['
				// (unlikely but possible).
				if len(vectorRaw)%4 == 0 {
					vector = bytesToFloat32(vectorRaw)
				} else {
					continue // Skip malformed
				}
			}
		} else {
			// Not JSON, assume binary
			if len(vectorRaw)%4 == 0 {
				vector = bytesToFloat32(vectorRaw)
			}
			// If not a multiple of 4, it's corrupt or unsupported format, so we skip.
			if vector == nil {
				continue
			}
		}

		var result any
		if err := json.Unmarshal(resultJSON, &result); err != nil {
			continue // Skip malformed
		}

		ttl := time.Duration(expiresAtNano - time.Now().UnixNano())
		if ttl > 0 {
			// Add to memory store without writing back to DB
			_ = s.memoryStore.Add(ctx, key, vector, result, ttl)
		}
	}

	if err := rows.Err(); err != nil {
		return err
	}

	return nil
}

// Add adds a new entry to both memory and DB.
//
// Summary: Stores a vector embedding and its associated result in both the in-memory cache and the SQLite database.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - key: string. The partition key.
//   - vector: []float32. The embedding vector.
//   - result: any. The result object to store.
//   - ttl: time.Duration. The time-to-live for the entry.
//
// Returns:
//   - error: An error if persistence fails.
func (s *SQLiteVectorStore) Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error {
	// Add to memory first
	if err := s.memoryStore.Add(ctx, key, vector, result, ttl); err != nil {
		return err
	}

	// Add to DB
	// We store vector as binary for performance
	vectorBytes := float32ToBytes(vector)

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	expiresAt := time.Now().Add(ttl).UnixNano()

	_, err = s.db.ExecContext(ctx, "INSERT INTO semantic_cache_entries (key, vector, result, expires_at) VALUES (?, ?, ?, ?)",
		key, vectorBytes, resultJSON, expiresAt)
	if err != nil {
		return err
	}

	// Probabilistic pruning (1 in 100 chance) to prevent unbound growth
	// without impacting performance on every write.
	if time.Now().UnixNano()%100 == 0 {
		go func() {
			// Best effort prune
			now := time.Now().UnixNano()
			// Use background context with short timeout for async cleanup so it doesn't block unrelated things,
			// but also doesn't hang forever.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = s.db.ExecContext(ctx, "DELETE FROM semantic_cache_entries WHERE expires_at <= ?", now)
		}()
	}

	return nil
}

// Search searches in memory.
//
// Summary: Performs an in-memory nearest neighbor search for the given query vector.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - key: string. The partition key to filter by.
//   - query: []float32. The query embedding vector.
//
// Returns:
//   - any: The best matching cached result.
//   - float32: The similarity score (cosine similarity).
//   - bool: True if a match was found.
func (s *SQLiteVectorStore) Search(ctx context.Context, key string, query []float32) (any, float32, bool) {
	return s.memoryStore.Search(ctx, key, query)
}

// Prune removes expired entries from both memory and DB.
//
// Summary: Removes expired cache entries from the in-memory store and the SQLite database.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - key: string. The partition key (optional).
func (s *SQLiteVectorStore) Prune(ctx context.Context, key string) {
	s.memoryStore.Prune(ctx, key)

	now := time.Now().UnixNano()
	_, _ = s.db.ExecContext(ctx, "DELETE FROM semantic_cache_entries WHERE expires_at <= ?", now)
}

// Close closes the database connection.
//
// Summary: Closes the SQLite database connection.
//
// Returns:
//   - error: An error if closing fails.
func (s *SQLiteVectorStore) Close() error {
	return s.db.Close()
}
