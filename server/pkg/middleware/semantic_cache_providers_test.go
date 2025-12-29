// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaEmbeddingProvider_Embed(t *testing.T) {
	// Mock Ollama Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/embeddings" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req ollamaEmbeddingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if req.Prompt == "error" {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		resp := ollamaEmbeddingResponse{
			Embedding: []float32{0.1, 0.2, 0.3},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOllamaEmbeddingProvider(server.URL, "test-model")

	// Test Success
	embedding, err := provider.Embed(context.Background(), "test-input")
	assert.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, embedding)

	// Test Error
	embedding, err = provider.Embed(context.Background(), "error")
	assert.Error(t, err)
	assert.Nil(t, embedding)
}

func TestHttpEmbeddingProvider_Embed(t *testing.T) {
	// Mock HTTP Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		bodyBytes, _ := json.Marshal(map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"embedding": []float64{0.1, 0.2, 0.3},
				},
			},
		})
		w.Write(bodyBytes)
	}))
	defer server.Close()

	provider, err := NewHTTPEmbeddingProvider(
		server.URL,
		map[string]string{"Authorization": "Bearer test"},
		`{"input": "{{.input}}"}`,
		"data[0].embedding",
	)
	assert.NoError(t, err)

	// Test Success
	embedding, err := provider.Embed(context.Background(), "test-input")
	assert.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, embedding)
}
