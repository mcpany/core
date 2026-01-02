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
	"google.golang.org/protobuf/proto"
)

func TestNewPineconeClient(t *testing.T) {
	t.Run("missing_api_key", func(t *testing.T) {
		cfg := &configv1.PineconeVectorDB{}
		_, err := NewPineconeClient(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "api_key is required")
	})

	t.Run("missing_host_info", func(t *testing.T) {
		cfg := &configv1.PineconeVectorDB{
			ApiKey: proto.String("test-key"),
		}
		_, err := NewPineconeClient(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "host OR (index_name, project_id, environment) must be provided")
	})

	t.Run("valid_host", func(t *testing.T) {
		cfg := &configv1.PineconeVectorDB{
			ApiKey: proto.String("test-key"),
			Host:   proto.String("https://example.com"),
		}
		client, err := NewPineconeClient(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "https://example.com", client.baseURL)
	})

	t.Run("construct_url", func(t *testing.T) {
		cfg := &configv1.PineconeVectorDB{
			ApiKey:      proto.String("test-key"),
			IndexName:   proto.String("my-index"),
			ProjectId:   proto.String("my-project"),
			Environment: proto.String("us-west1"),
		}
		client, err := NewPineconeClient(cfg)
		assert.NoError(t, err)
		assert.Equal(t, "https://my-index-my-project.svc.us-west1.pinecone.io", client.baseURL)
	})
}

func TestPineconeClient_Operations(t *testing.T) {
	// Setup mock server
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		assert.Equal(t, "test-key", r.Header.Get("Api-Key"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "application/json", r.Header.Get("Accept"))

		switch r.URL.Path {
		case "/query":
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			assert.Contains(t, req, "vector")
			assert.Contains(t, req, "topK")

			json.NewEncoder(w).Encode(map[string]interface{}{
				"matches": []interface{}{},
			})
		case "/vectors/upsert":
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			assert.Contains(t, req, "vectors")

			json.NewEncoder(w).Encode(map[string]interface{}{
				"upsertedCount": 1,
			})
		case "/vectors/delete":
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			// Check logic for delete all vs ids
			if _, ok := req["deleteAll"]; ok {
				//
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
			})
		case "/describe_index_stats":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"totalVectorCount": 100,
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	cfg := &configv1.PineconeVectorDB{
		ApiKey: proto.String("test-key"),
		Host:   proto.String(server.URL),
	}
	client, err := NewPineconeClient(cfg)
	assert.NoError(t, err)

	ctx := context.Background()

	t.Run("Query", func(t *testing.T) {
		vector := []float32{0.1, 0.2}
		res, err := client.Query(ctx, vector, 10, nil, "")
		assert.NoError(t, err)
		assert.Contains(t, res, "matches")
	})

	t.Run("QueryWithFilterAndNamespace", func(t *testing.T) {
		vector := []float32{0.1, 0.2}
		filter := map[string]interface{}{"foo": "bar"}
		res, err := client.Query(ctx, vector, 10, filter, "ns1")
		assert.NoError(t, err)
		assert.Contains(t, res, "matches")
	})

	t.Run("Upsert", func(t *testing.T) {
		vectors := []map[string]interface{}{
			{"id": "1", "values": []float64{0.1}},
		}
		res, err := client.Upsert(ctx, vectors, "")
		assert.NoError(t, err)
		assert.Contains(t, res, "upsertedCount")
	})

	t.Run("UpsertWithNamespace", func(t *testing.T) {
		vectors := []map[string]interface{}{
			{"id": "1", "values": []float64{0.1}},
		}
		res, err := client.Upsert(ctx, vectors, "ns1")
		assert.NoError(t, err)
		assert.Contains(t, res, "upsertedCount")
	})

	t.Run("Delete", func(t *testing.T) {
		ids := []string{"1"}
		res, err := client.Delete(ctx, ids, "", nil)
		assert.NoError(t, err)
		assert.Contains(t, res, "success")
	})

	t.Run("DeleteWithFilterAndNamespace", func(t *testing.T) {
		ids := []string{"1"}
		filter := map[string]interface{}{"foo": "bar"}
		res, err := client.Delete(ctx, ids, "ns1", filter)
		assert.NoError(t, err)
		assert.Contains(t, res, "success")
	})

	t.Run("DeleteAll", func(t *testing.T) {
		res, err := client.Delete(ctx, nil, "", nil)
		assert.NoError(t, err)
		assert.Contains(t, res, "success")
	})

	t.Run("DescribeIndexStats", func(t *testing.T) {
		res, err := client.DescribeIndexStats(ctx, nil)
		assert.NoError(t, err)
		assert.Contains(t, res, "totalVectorCount")
	})

	t.Run("DescribeIndexStatsWithFilter", func(t *testing.T) {
		filter := map[string]interface{}{"foo": "bar"}
		res, err := client.DescribeIndexStats(ctx, filter)
		assert.NoError(t, err)
		assert.Contains(t, res, "totalVectorCount")
	})
}

func TestPineconeClient_DoRequestErrors(t *testing.T) {
	// 1. Invalid URL (force doRequest failure)
	// Since we construct the URL from config, we can inject a bad host or use a method that makes a bad URL?
	// But `url.JoinPath` handles a lot. Let's try an invalid host that causes http client to fail (e.g. invalid port).

	cfg := &configv1.PineconeVectorDB{
		ApiKey: proto.String("test-key"),
		Host:   proto.String("http://[::1]:namedport"), // Invalid
	}
	client, err := NewPineconeClient(cfg)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = client.Query(ctx, []float32{0.1}, 10, nil, "")
	assert.Error(t, err)
	// Depending on Go version and platform the error might vary, but it should fail in client.Do or NewRequest
}

func TestPineconeClient_Errors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "bad request"}`))
	}))
	defer server.Close()

	cfg := &configv1.PineconeVectorDB{
		ApiKey: proto.String("test-key"),
		Host:   proto.String(server.URL),
	}
	client, err := NewPineconeClient(cfg)
	assert.NoError(t, err)

	ctx := context.Background()
	_, err = client.Query(ctx, []float32{0.1}, 10, nil, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "pinecone request failed: status=400")
}
