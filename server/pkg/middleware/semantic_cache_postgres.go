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
// It connects to the database and ensures the schema exists.
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
// It ensures the schema exists.
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
// ctx is the context for the request.
// key is the key.
// vector is the vector.
// result is the result.
// ttl is the ttl.
//
// Returns an error if the operation fails.
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
// ctx is the context for the request.
// key is the key.
// query is the query.
//
// Returns the result.
// Returns the result.
// Returns true if successful.
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
// ctx is the context for the request.
// key is the key.
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
// Returns an error if the operation fails.
func (s *PostgresVectorStore) Close() error {
	return s.db.Close()
}

// Ensure interface compatibility.
var _ VectorStore = (*PostgresVectorStore)(nil)
