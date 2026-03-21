// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package vector

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestNewMilvusClient_Validation(t *testing.T) {
	t.Run("Missing Address", func(t *testing.T) {
		cfg := configv1.MilvusVectorDB_builder{
			CollectionName: proto.String("coll"),
		}.Build()
		c, err := NewMilvusClient(cfg)
		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "address is required")
	})

	t.Run("Missing Collection", func(t *testing.T) {
		cfg := configv1.MilvusVectorDB_builder{
			Address: proto.String("127.0.0.1:19530"),
		}.Build()
		c, err := NewMilvusClient(cfg)
		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "collection_name is required")
	})

	t.Run("Connection Failure", func(t *testing.T) {
		// Attempt to connect to a random closed port
		cfg := configv1.MilvusVectorDB_builder{
			Address:        proto.String("127.0.0.1:54321"),
			CollectionName: proto.String("test"),
		}.Build()
		c, err := NewMilvusClient(cfg)
		// Milvus client creation might succeed initially (lazy connect) or fail fast.
		// If it connects eagerly and fails, we get error.
		// The current implementation calls c.HasCollection immediately, so it should fail.
		assert.Error(t, err)
		assert.Nil(t, c)
	})
}

func TestNewPineconeClient_Validation(t *testing.T) {
	t.Run("Missing API Key", func(t *testing.T) {
		cfg := configv1.PineconeVectorDB_builder{}.Build()
		c, err := NewPineconeClient(cfg)
		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("Missing Host Info", func(t *testing.T) {
		cfg := configv1.PineconeVectorDB_builder{
			ApiKey: proto.String("key"),
		}.Build()
		c, err := NewPineconeClient(cfg)
		assert.Error(t, err)
		assert.Nil(t, c)
		assert.Contains(t, err.Error(), "host OR (index_name, project_id, environment) must be provided")
	})

	t.Run("Valid with Host", func(t *testing.T) {
		cfg := configv1.PineconeVectorDB_builder{
			ApiKey: proto.String("key"),
			Host:   proto.String("https://custom.pinecone.io"),
		}.Build()
		c, err := NewPineconeClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, "https://custom.pinecone.io", c.baseURL)
	})

	t.Run("Valid with Components", func(t *testing.T) {
		cfg := configv1.PineconeVectorDB_builder{
			ApiKey:      proto.String("key"),
			IndexName:   proto.String("idx"),
			ProjectId:   proto.String("proj"),
			Environment: proto.String("env"),
		}.Build()
		c, err := NewPineconeClient(cfg)
		require.NoError(t, err)
		assert.NotNil(t, c)
		assert.Equal(t, "https://idx-proj.svc.env.pinecone.io", c.baseURL)
	})
}

func TestPineconeClient_Methods(t *testing.T) {
	// Start a mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Headers
		if r.Header.Get("Api-Key") != "test-key" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Handle Routes
		switch r.URL.Path {
		case "/query":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"matches": []map[string]interface{}{
					{"id": "vec1", "score": 0.99},
				},
			})
		case "/vectors/upsert":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"upsertedCount": 1,
			})
		case "/vectors/delete":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true, // Not standard Pinecone response but enough for our generic map return
			})
		case "/describe_index_stats":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"dimension": 1536,
			})
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))
	defer ts.Close()

	cfg := configv1.PineconeVectorDB_builder{
		ApiKey: proto.String("test-key"),
		Host:   proto.String(ts.URL),
	}.Build()
	c, err := NewPineconeClient(cfg)
	require.NoError(t, err)

	ctx := context.Background()

	// Test Query
	res, err := c.Query(ctx, []float32{0.1, 0.2}, 10, nil, "")
	assert.NoError(t, err)
	assert.Contains(t, res, "matches")

	// Test Upsert
	res, err = c.Upsert(ctx, []map[string]interface{}{
		{"id": "v1", "values": []float64{0.1, 0.2}},
	}, "")
	assert.NoError(t, err)
	assert.Contains(t, res, "upsertedCount")

	// Test Delete
	res, err = c.Delete(ctx, []string{"v1"}, "", nil)
	assert.NoError(t, err)
	// assert.Contains(t, res, "success") // Pinecone delete response is empty json {} usually unless checking body

	// Test DescribeIndexStats
	res, err = c.DescribeIndexStats(ctx, nil)
	assert.NoError(t, err)
	// json unmarshals numbers as float64
	assert.Equal(t, float64(1536), res["dimension"])

	// Test Error Case (404)
	c.baseURL = ts.URL + "/invalid"
	_, err = c.Query(ctx, []float32{0.1}, 1, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status=404")
}
