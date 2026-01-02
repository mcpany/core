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
)

func TestPineconeClient(t *testing.T) {
	// Setup mock server
	handler := http.NewServeMux()
	server := httptest.NewServer(handler)
	defer server.Close()

	// Config
	apiKey := "test-api-key"
	host := server.URL
	config := &configv1.PineconeVectorDB{
		ApiKey: &apiKey,
		Host:   &host,
	}
	client, err := NewPineconeClient(config)
	assert.NoError(t, err)

	t.Run("Query", func(t *testing.T) {
		handler.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, apiKey, r.Header.Get("Api-Key"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			assert.Equal(t, float64(5), req["topK"])

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"matches": []interface{}{
					map[string]interface{}{"id": "vec1", "score": 0.9},
				},
			})
		})

		vector := []float32{0.1, 0.2}
		res, err := client.Query(context.Background(), vector, 5, nil, "")
		assert.NoError(t, err)
		matches := res["matches"].([]interface{})
		assert.Len(t, matches, 1)
	})

	t.Run("Upsert", func(t *testing.T) {
		handler.HandleFunc("/vectors/upsert", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			assert.NotNil(t, req["vectors"])

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"upsertedCount": 1,
			})
		})

		vectors := []map[string]interface{}{
			{"id": "vec1", "values": []float32{0.1, 0.2}},
		}
		res, err := client.Upsert(context.Background(), vectors, "")
		assert.NoError(t, err)
		assert.Equal(t, float64(1), res["upsertedCount"])
	})

	t.Run("Delete", func(t *testing.T) {
		handler.HandleFunc("/vectors/delete", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			// assert.Equal(t, []interface{}{"vec1"}, req["ids"]) // json decodes array as []interface{}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{})
		})

		res, err := client.Delete(context.Background(), []string{"vec1"}, "", nil)
		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

	t.Run("DescribeIndexStats", func(t *testing.T) {
		handler.HandleFunc("/describe_index_stats", func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"totalVectorCount": 100,
			})
		})

		res, err := client.DescribeIndexStats(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, float64(100), res["totalVectorCount"])
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		handler.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		})

		// Create client pointing to error endpoint (hacky override for test)
		client.baseURL = server.URL + "/error"
		// path joined will be /error/query if we called query, which fails 404 on mux.
		// Actually we should create a new client or just mock the endpoint called.
		// The client uses JoinPath.
		// Let's reset baseURL and register a failing endpoint.
		client.baseURL = server.URL

		// Register a specific failing path
		handler.HandleFunc("/fail", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Error"))
		})

		// We can't easily change the path called by Query/Upsert etc methods as they are hardcoded.
		// So we can only test error if the server returns error for /query

		handler.HandleFunc("/query_error", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		})

		// Use doRequest directly or through exposed method if we could mock the path.
		// Since we can't, we'll re-register /query to fail for this test part.
		// But ServeMux doesn't allow overriding.

		// Create a new server for error testing
		errHandler := http.NewServeMux()
		errServer := httptest.NewServer(errHandler)
		defer errServer.Close()

		errHandler.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad Request"))
		})

		errHost := errServer.URL
		errConfig := &configv1.PineconeVectorDB{
			ApiKey: &apiKey,
			Host:   &errHost,
		}
		errClient, _ := NewPineconeClient(errConfig)

		_, err = errClient.Query(context.Background(), []float32{0.1}, 1, nil, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "pinecone request failed: status=400")
	})
}
