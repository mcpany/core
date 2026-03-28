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
	"github.com/stretchr/testify/require"
)

func TestOllamaEmbeddingProvider_Embed(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/embeddings", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		// Verify Body
		var req ollamaEmbeddingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-input", req.Prompt)
		assert.Equal(t, "test-model", req.Model)

		// Send Response
		resp := ollamaEmbeddingResponse{
			Embedding: []float32{0.1, 0.2, 0.3},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Initialize provider with mock server URL
	provider := NewOllamaEmbeddingProvider(mockServer.URL, "test-model")

	// Execute
	embedding, err := provider.Embed(context.Background(), "test-input")

	// Verify
	require.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, embedding)
}

func TestOllamaEmbeddingProvider_Embed_Error(t *testing.T) {
	// Mock server returning 500
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer mockServer.Close()

	provider := NewOllamaEmbeddingProvider(mockServer.URL, "test-model")

	embedding, err := provider.Embed(context.Background(), "test-input")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ollama api error (status 500)")
	assert.Nil(t, embedding)
}

func TestOllamaEmbeddingProvider_Embed_MalformedResponse(t *testing.T) {
	// Mock server returning invalid JSON
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"embedding": [0.1, 0.2`)) // Incomplete JSON
	}))
	defer mockServer.Close()

	provider := NewOllamaEmbeddingProvider(mockServer.URL, "test-model")

	embedding, err := provider.Embed(context.Background(), "test-input")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
	assert.Nil(t, embedding)
}

func TestOllamaEmbeddingProvider_Embed_Empty(t *testing.T) {
	// Mock server returning empty embedding
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ollamaEmbeddingResponse{
			Embedding: []float32{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	provider := NewOllamaEmbeddingProvider(mockServer.URL, "test-model")

	embedding, err := provider.Embed(context.Background(), "test-input")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no embedding data returned")
	assert.Nil(t, embedding)
}

func TestOllamaEmbeddingProvider_Defaults(t *testing.T) {
	provider := NewOllamaEmbeddingProvider("", "")
	assert.Equal(t, "http://localhost:11434", provider.baseURL)
	assert.Equal(t, "nomic-embed-text", provider.model)
}
