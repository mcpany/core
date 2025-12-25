// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockEmbeddingProvider is a mock implementation of EmbeddingProvider
type MockEmbeddingProvider struct {
	EmbedFunc func(ctx context.Context, text string) ([]float32, error)
}

func (m *MockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	if m.EmbedFunc != nil {
		return m.EmbedFunc(ctx, text)
	}
	return nil, nil
}

func TestSemanticCache_Get_Set(t *testing.T) {
	mockProvider := &MockEmbeddingProvider{
		EmbedFunc: func(ctx context.Context, text string) ([]float32, error) {
			// Return a fixed embedding for simplicity
			if text == "test query" {
				return []float32{1.0, 0.0, 0.0}, nil
			}
			return []float32{0.0, 1.0, 0.0}, nil
		},
	}

	cache := NewSemanticCache(mockProvider, 0.9)

	ctx := context.Background()
	key := "test-key"

	// 1. Get on empty cache
	result, _, hit, err := cache.Get(ctx, key, "test query")
	require.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, result)

	// 2. Set a value
	embedding := []float32{1.0, 0.0, 0.0}
	err = cache.Set(ctx, key, embedding, "cached result", time.Minute)
	require.NoError(t, err)

	// 3. Get with high similarity
	result, _, hit, err = cache.Get(ctx, key, "test query") // embed -> {1, 0, 0}
	require.NoError(t, err)
	assert.True(t, hit)
	assert.Equal(t, "cached result", result)

	// 4. Get with low similarity
	result, _, hit, err = cache.Get(ctx, key, "other query") // embed -> {0, 1, 0}
	require.NoError(t, err)
	assert.False(t, hit)
	assert.Nil(t, result)
}

func TestSemanticCache_Expiration(t *testing.T) {
	mockProvider := &MockEmbeddingProvider{
		EmbedFunc: func(ctx context.Context, text string) ([]float32, error) {
			return []float32{1.0}, nil
		},
	}
	cache := NewSemanticCache(mockProvider, 0.9)
	ctx := context.Background()
	key := "expire-key"

	err := cache.Set(ctx, key, []float32{1.0}, "value", 10*time.Millisecond)
	require.NoError(t, err)

	// Immediate get should hit
	_, _, hit, _ := cache.Get(ctx, key, "any")
	assert.True(t, hit)

	// Wait for expiration
	time.Sleep(20 * time.Millisecond)

	// Get should miss
	_, _, hit, _ = cache.Get(ctx, key, "any")
	assert.False(t, hit)
}

func TestOpenAIEmbeddingProvider(t *testing.T) {
	// Mock OpenAI API
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req openAIEmbeddingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "test-model", req.Model)

		resp := openAIEmbeddingResponse{
			Data: []struct {
				Embedding []float32 `json:"embedding"`
			}{
				{Embedding: []float32{0.1, 0.2, 0.3}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIEmbeddingProvider("test-key", "test-model")
	provider.baseURL = server.URL // Override base URL for testing

	ctx := context.Background()
	embedding, err := provider.Embed(ctx, "test input")
	require.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, embedding)
}

func TestOpenAIEmbeddingProvider_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer server.Close()

	provider := NewOpenAIEmbeddingProvider("test-key", "")
	provider.baseURL = server.URL

	ctx := context.Background()
	_, err := provider.Embed(ctx, "test input")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "openai api error (status 500)")
}

func TestOpenAIEmbeddingProvider_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIEmbeddingResponse{
			Error: &struct {
				Message string `json:"message"`
			}{
				Message: "invalid api key",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIEmbeddingProvider("test-key", "")
	provider.baseURL = server.URL

	ctx := context.Background()
	_, err := provider.Embed(ctx, "test input")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "openai error: invalid api key")
}
