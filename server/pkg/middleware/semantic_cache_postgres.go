// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	// lib/pq is the Postgres driver.
	_ "github.com/lib/pq"
)

// PostgresVectorStore implements VectorStore using PostgreSQL and pgvector.
type PostgresVectorStore struct {
	db *sql.DB
}

// NewPostgresVectorStore creates a new PostgresVectorStore.
//
// Summary: Creates and initializes a new PostgresVectorStore.
//
// It connects to the specified PostgreSQL database using the provided DSN and ensures
// that the required vector extension and schema tables exist.
//
// Parameters:
//   - dsn (string): The data source name for the PostgreSQL connection.
//
// Returns:
//   - (*PostgresVectorStore): The initialized vector store instance.
//   - (error): An error if the connection fails or schema creation fails.
//
// Errors:
//   - Returns an error if the DSN is empty.
//   - Returns an error if the database connection cannot be established.
func NewPostgresVectorStore(dsn string) (*PostgresVectorStore, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is required")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	store, err := NewPostgresVectorStoreWithDB(db)
	if err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

// NewPostgresVectorStoreWithDB creates a new PostgresVectorStore using an existing database connection.
//
// Summary: Initializes the vector store with an existing DB connection.
//
// It verifies the connection and ensures the pgvector extension and schema tables exist.
//
// Parameters:
//   - db (*sql.DB): An existing database connection.
//
// Returns:
//   - (*PostgresVectorStore): The initialized vector store instance.
//   - (error): An error if the connection verification or schema setup fails.
//
// Side Effects:
//   - Creates the `vector` extension if it doesn't exist.
//   - Creates the `semantic_cache_entries` table and indexes if they don't exist.
func NewPostgresVectorStoreWithDB(db *sql.DB) (*PostgresVectorStore, error) {
	// Verify connection
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelPing()
	if err := db.PingContext(ctxPing); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Create extension if not exists
	ctxExt, cancelExt := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelExt()
	if _, err := db.ExecContext(ctxExt, "CREATE EXTENSION IF NOT EXISTS vector"); err != nil {
		return nil, fmt.Errorf("failed to create vector extension: %w", err)
	}

	// Create table if not exists
	// We use unconstrained vector type to allow different embedding models without schema change.
	// Note: Without specified dimension, we cannot create IVFFlat/HNSW indexes easily,
	// but exact nearest neighbor search (KNN) works fine.
	schema := `
	CREATE TABLE IF NOT EXISTS semantic_cache_entries (
		id SERIAL PRIMARY KEY,
		key TEXT NOT NULL,
		vector vector,
		result JSONB,
		expires_at TIMESTAMP WITH TIME ZONE
	);
	CREATE INDEX IF NOT EXISTS idx_semantic_cache_key ON semantic_cache_entries(key);
	CREATE INDEX IF NOT EXISTS idx_semantic_cache_expires_at ON semantic_cache_entries(expires_at);
	`
	ctxSchema, cancelSchema := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelSchema()
	if _, err := db.ExecContext(ctxSchema, schema); err != nil {
		return nil, fmt.Errorf("failed to create semantic_cache_entries table: %w", err)
	}

	return &PostgresVectorStore{
		db: db,
	}, nil
}

// Add adds a new entry to the vector store.
//
// Summary: Inserts a new semantic cache entry into the database.
//
// Parameters:
//   - ctx (context.Context): The context for the database operation.
//   - key (string): The unique identifier for the cache entry (e.g., query hash).
//   - vector ([]float32): The embedding vector associated with the entry.
//   - result (any): The result object to store (serialized to JSON).
//   - ttl (time.Duration): The time-to-live for the cache entry.
//
// Returns:
//   - (error): An error if marshaling fails or the database insert fails.
//
// Side Effects:
//   - Writes a new row to the `semantic_cache_entries` table.
func (s *PostgresVectorStore) Add(ctx context.Context, key string, vector []float32, result any, ttl time.Duration) error {
	vectorJSON, err := json.Marshal(vector)
	if err != nil {
		return fmt.Errorf("failed to marshal vector: %w", err)
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	expiresAt := time.Now().Add(ttl)

	query := `
		INSERT INTO semantic_cache_entries (key, vector, result, expires_at)
		VALUES ($1, $2, $3, $4)
	`

	// pgvector accepts '[1,2,3]' string representation for vector type
	_, err = s.db.ExecContext(ctx, query, key, string(vectorJSON), resultJSON, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	return nil
}

// Search searches for the most similar entry in the vector store.
//
// Summary: Finds the nearest neighbor vector in the cache.
//
// It calculates the cosine distance between the query vector and stored vectors,
// returning the closest match that hasn't expired.
//
// Parameters:
//   - ctx (context.Context): The context for the database query.
//   - key (string): The key to filter by (optional optimization).
//   - query ([]float32): The query embedding vector.
//
// Returns:
//   - (any): The cached result object.
//   - (float32): The similarity score (1.0 - cosine distance).
//   - (bool): True if a match was found, false otherwise.
//
// Errors:
//   - Returns false if no matching row is found or if unmarshaling fails.
func (s *PostgresVectorStore) Search(ctx context.Context, key string, query []float32) (any, float32, bool) {
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, 0, false
	}

	// Use Cosine Distance Operator (<=>)
	// Similarity = 1 - Distance (roughly, assuming normalized vectors or pgvector semantics)
	// pgvector cosine distance is 1 - cosine_similarity.
	// So we order by distance ASC.
	// We filter by key and expiration.
	sqlQuery := `
		SELECT result, (vector <=> $1) as distance
		FROM semantic_cache_entries
		WHERE key = $2 AND expires_at > $3
		ORDER BY distance ASC
		LIMIT 1
	`

	var resultJSON []byte
	var distance float64

	err = s.db.QueryRowContext(ctx, sqlQuery, string(queryJSON), key, time.Now()).Scan(&resultJSON, &distance)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, false
		}
		// Log error?
		return nil, 0, false
	}

	var result any
	if err := json.Unmarshal(resultJSON, &result); err != nil {
		return nil, 0, false
	}

	// Convert distance to similarity
	similarity := float32(1.0 - distance)

	return result, similarity, true
}

// Prune removes expired entries.
//
// Summary: Deletes expired cache entries from the database.
//
// Parameters:
//   - ctx (context.Context): The context for the database operation.
//   - key (string): Optional key to scope the pruning. If empty, prunes all expired entries.
//
// Returns:
//   None.
//
// Side Effects:
//   - Deletes rows from `semantic_cache_entries`.
func (s *PostgresVectorStore) Prune(ctx context.Context, key string) {
	query := "DELETE FROM semantic_cache_entries WHERE expires_at <= $1"
	args := []interface{}{time.Now()}

	if key != "" {
		query += " AND key = $2"
		args = append(args, key)
	}

	_, _ = s.db.ExecContext(ctx, query, args...)
}

// Close closes the database connection.
//
// Summary: Closes the underlying SQL database connection.
//
// Parameters:
//   None.
//
// Returns:
//   - (error): An error if the closing operation fails.
func (s *PostgresVectorStore) Close() error {
	return s.db.Close()
}

// Ensure interface compatibility.
var _ VectorStore = (*PostgresVectorStore)(nil)
