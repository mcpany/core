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

func TestNewOllamaEmbeddingProvider(t *testing.T) {
	tests := []struct {
		name          string
		baseURL       string
		model         string
		expectedURL   string
		expectedModel string
	}{
		{
			name:          "Valid config",
			baseURL:       "http://custom-ollama:11434",
			model:         "llama3",
			expectedURL:   "http://custom-ollama:11434",
			expectedModel: "llama3",
		},
		{
			name:          "Defaults",
			baseURL:       "",
			model:         "",
			expectedURL:   "http://localhost:11434",
			expectedModel: "nomic-embed-text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewOllamaEmbeddingProvider(tt.baseURL, tt.model)
			assert.NotNil(t, provider)
			assert.Equal(t, tt.expectedURL, provider.baseURL)
			assert.Equal(t, tt.expectedModel, provider.model)
		})
	}
}

func TestOllamaEmbeddingProvider_Embed(t *testing.T) {
	tests := []struct {
		name          string
		handler       http.HandlerFunc
		baseURL       string
		inputText     string
		expectedEmbed []float32
		expectError   bool
		errorContains string
	}{
		{
			name: "Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/api/embeddings", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody map[string]string
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				require.NoError(t, err)
				assert.Equal(t, "test-model", reqBody["model"])
				assert.Equal(t, "hello", reqBody["prompt"])

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"embedding": [0.1, 0.2, 0.3]}`))
			},
			inputText:     "hello",
			expectedEmbed: []float32{0.1, 0.2, 0.3},
			expectError:   false,
		},
		{
			name: "HTTP Error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Ollama Error"))
			},
			inputText:     "test",
			expectError:   true,
			errorContains: "ollama api error",
		},
		{
			name: "Invalid JSON Response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`invalid json`))
			},
			inputText:     "test",
			expectError:   true,
			errorContains: "failed to decode response",
		},
		{
			name: "Empty Embedding in Response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"embedding": []}`))
			},
			inputText:     "test",
			expectError:   true,
			errorContains: "no embedding data returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			provider := NewOllamaEmbeddingProvider(server.URL, "test-model")
			require.NotNil(t, provider.client) // Should be initialized

			embed, err := provider.Embed(context.Background(), tt.inputText)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEmbed, embed)
			}
		})
	}
}
