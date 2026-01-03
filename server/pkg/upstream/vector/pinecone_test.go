// Copyright 2026 Author(s) of MCP Any
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
)

func ptr(s string) *string {
	return &s
}

func TestNewPineconeClient(t *testing.T) {
	tests := []struct {
		name      string
		config    *configv1.PineconeVectorDB
		expectErr bool
		check     func(*testing.T, *PineconeClient)
	}{
		{
			name: "valid config with host",
			config: &configv1.PineconeVectorDB{
				ApiKey: ptr("test-key"),
				Host:   ptr("https://test-index.pinecone.io"),
			},
			expectErr: false,
			check: func(t *testing.T, c *PineconeClient) {
				assert.Equal(t, "https://test-index.pinecone.io", c.baseURL)
			},
		},
		{
			name: "valid config with components",
			config: &configv1.PineconeVectorDB{
				ApiKey:      ptr("test-key"),
				IndexName:   ptr("my-index"),
				ProjectId:   ptr("proj123"),
				Environment: ptr("us-west1"),
			},
			expectErr: false,
			check: func(t *testing.T, c *PineconeClient) {
				assert.Equal(t, "https://my-index-proj123.svc.us-west1.pinecone.io", c.baseURL)
			},
		},
		{
			name:      "missing api key",
			config:    &configv1.PineconeVectorDB{},
			expectErr: true,
		},
		{
			name: "missing host and components",
			config: &configv1.PineconeVectorDB{
				ApiKey: ptr("test-key"),
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewPineconeClient(tt.config)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.check != nil {
					tt.check(t, c)
				}
			}
		})
	}
}

func TestPineconeClient_Operations(t *testing.T) {
	// Setup mock server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Headers
		assert.Equal(t, "test-key", r.Header.Get("Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var reqBody map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil && err.Error() != "EOF" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		switch r.URL.Path {
		case "/query":
			assert.Equal(t, "POST", r.Method)
			// Check request body
			assert.Contains(t, reqBody, "vector")
			assert.Contains(t, reqBody, "topK")

			json.NewEncoder(w).Encode(map[string]interface{}{
				"matches": []interface{}{
					map[string]interface{}{"id": "1", "score": 0.9},
				},
			})
		case "/vectors/upsert":
			assert.Equal(t, "POST", r.Method)
			assert.Contains(t, reqBody, "vectors")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"upsertedCount": 1,
			})
		case "/vectors/delete":
			assert.Equal(t, "POST", r.Method)
			if _, ok := reqBody["deleteAll"]; ok {
				json.NewEncoder(w).Encode(map[string]interface{}{})
			} else {
				assert.Contains(t, reqBody, "ids")
				json.NewEncoder(w).Encode(map[string]interface{}{})
			}
		case "/describe_index_stats":
			assert.Equal(t, "POST", r.Method)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"namespaces": map[string]interface{}{
					"ns1": map[string]interface{}{"vectorCount": 10},
				},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	client, err := NewPineconeClient(&configv1.PineconeVectorDB{
		ApiKey: ptr("test-key"),
		Host:   ptr(server.URL),
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test Query
	t.Run("Query", func(t *testing.T) {
		res, err := client.Query(ctx, []float32{0.1, 0.2}, 10, nil, "ns1")
		assert.NoError(t, err)
		assert.NotNil(t, res["matches"])
	})

	// Test Upsert
	t.Run("Upsert", func(t *testing.T) {
		vectors := []map[string]interface{}{
			{"id": "1", "values": []float32{0.1, 0.2}},
		}
		res, err := client.Upsert(ctx, vectors, "ns1")
		assert.NoError(t, err)
		assert.Equal(t, float64(1), res["upsertedCount"])
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		res, err := client.Delete(ctx, []string{"1"}, "ns1", nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	// Test Delete All
	t.Run("DeleteAll", func(t *testing.T) {
		res, err := client.Delete(ctx, nil, "ns1", nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	// Test DescribeIndexStats
	t.Run("DescribeIndexStats", func(t *testing.T) {
		res, err := client.DescribeIndexStats(ctx, nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})
}

func TestPineconeClient_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	client, err := NewPineconeClient(&configv1.PineconeVectorDB{
		ApiKey: ptr("test-key"),
		Host:   ptr(server.URL),
	})
	require.NoError(t, err)

	ctx := context.Background()

	_, err = client.Query(ctx, []float32{0.1}, 5, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pinecone request failed")
}
