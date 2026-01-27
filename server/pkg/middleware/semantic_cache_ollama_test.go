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

func TestNewOllamaEmbeddingProvider(t *testing.T) {
	// Test default values
	p := NewOllamaEmbeddingProvider("", "")
	assert.Equal(t, "http://localhost:11434", p.baseURL)
	assert.Equal(t, "nomic-embed-text", p.model)

	// Test custom values
	p = NewOllamaEmbeddingProvider("http://custom:11434", "llama3")
	assert.Equal(t, "http://custom:11434", p.baseURL)
	assert.Equal(t, "llama3", p.model)
}

func TestOllamaEmbeddingProvider_Embed(t *testing.T) {
	tests := []struct {
		name              string
		handler           http.HandlerFunc
		expectedEmbedding []float32
		expectError       bool
		errorContains     string
	}{
		{
			name: "Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/api/embeddings", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var body map[string]string
				err := json.NewDecoder(r.Body).Decode(&body)
				assert.NoError(t, err)
				assert.Equal(t, "test-model", body["model"])
				assert.Equal(t, "hello", body["prompt"])

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"embedding": [0.1, 0.2, 0.3]}`))
			},
			expectedEmbedding: []float32{0.1, 0.2, 0.3},
			expectError:       false,
		},
		{
			name: "HTTP Error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Ollama Error"))
			},
			expectError:   true,
			errorContains: "ollama api error (status 500)",
		},
		{
			name: "Invalid JSON Response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("invalid json"))
			},
			expectError:   true,
			errorContains: "failed to decode response",
		},
		{
			name: "Empty Embedding",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"embedding": []}`))
			},
			expectError:   true,
			errorContains: "no embedding data returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			provider := NewOllamaEmbeddingProvider(server.URL, "test-model")

			emb, err := provider.Embed(context.Background(), "hello")
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEmbedding, emb)
			}
		})
	}
}

func TestOllamaEmbeddingProvider_Embed_RequestCreationFailure(t *testing.T) {
	provider := NewOllamaEmbeddingProvider("http://[::1]:namedport", "test")
	_, err := provider.Embed(context.Background(), "test")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create request")
}
