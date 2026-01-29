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

func TestOpenAIEmbeddingProvider_Embed(t *testing.T) {
	// Setup mock server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		// Verify Body
		var req openAIEmbeddingRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		assert.NoError(t, err)
		assert.Equal(t, "test-input", req.Input)
		assert.Equal(t, "test-model", req.Model)

		// Send Response
		resp := openAIEmbeddingResponse{}
		resp.Data = []struct {
			Embedding []float32 `json:"embedding"`
		}{
			{
				Embedding: []float32{0.1, 0.2, 0.3},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// Initialize provider with mock server URL
	provider := NewOpenAIEmbeddingProvider("test-key", "test-model")
	provider.baseURL = mockServer.URL

	// Execute
	embedding, err := provider.Embed(context.Background(), "test-input")

	// Verify
	require.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, embedding)
}

func TestOpenAIEmbeddingProvider_Embed_Error(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer mockServer.Close()

	provider := NewOpenAIEmbeddingProvider("test-key", "test-model")
	provider.baseURL = mockServer.URL

	embedding, err := provider.Embed(context.Background(), "test-input")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openai api error (status 500)")
	assert.Nil(t, embedding)
}

func TestOpenAIEmbeddingProvider_Embed_APIError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIEmbeddingResponse{}
		resp.Error = &struct {
			Message string `json:"message"`
		}{
			Message: "Rate limit exceeded",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	provider := NewOpenAIEmbeddingProvider("test-key", "test-model")
	provider.baseURL = mockServer.URL

	embedding, err := provider.Embed(context.Background(), "test-input")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openai error: Rate limit exceeded")
	assert.Nil(t, embedding)
}

func TestSimpleVectorStore_Eviction(t *testing.T) {
	store := NewSimpleVectorStore()
	store.maxEntries = 2 // Small limit for testing

	key := "test-key"
	// Add 3 items with orthogonal vectors to avoid cosine similarity collision
	// [1, 0, 0]
	// [0, 1, 0]
	// [0, 0, 1]
	v1 := []float32{1, 0, 0}
	v2 := []float32{0, 1, 0}
	v3 := []float32{0, 0, 1}

	store.Add(context.Background(), key, v1, "result1", time.Minute)
	store.Add(context.Background(), key, v2, "result2", time.Minute)
	store.Add(context.Background(), key, v3, "result3", time.Minute)

	// Verify internal state directly to ensure eviction happened
	store.mu.RLock()
	entries := store.items[key]
	assert.Equal(t, 2, len(entries))
	assert.Equal(t, "result2", entries[0].Result)
	assert.Equal(t, "result3", entries[1].Result)
	store.mu.RUnlock()

	// Search verification
	// Verify that v1 is effectively evicted (or returns a very low score if found)
	// Since we are using orthogonal vectors, searching for v1 (which is no longer in the store)
	// should either return false or a result with score 0.
	res, score, found := store.Search(context.Background(), key, v1)
	if found {
		assert.Equal(t, float32(0), score)
	}

	// Verify that v2 is still present and matches perfectly
	res, score, found = store.Search(context.Background(), key, v2)
	assert.True(t, found)
	assert.InDelta(t, float32(1.0), score, 0.0001)
	assert.Equal(t, "result2", res)

	res, score, found = store.Search(context.Background(), key, v3)
	assert.True(t, found)
	assert.InDelta(t, float32(1.0), score, 0.0001)
	assert.Equal(t, "result3", res)
}

func TestSimpleVectorStore_Cleanup(t *testing.T) {
	store := NewSimpleVectorStore()
	key := "test-key"

	// Add item that expires quickly
	store.Add(context.Background(), key, []float32{0.1}, "result1", 1*time.Millisecond)

	time.Sleep(10 * time.Millisecond)

	// Add another item - this should trigger cleanup of the first one
	store.Add(context.Background(), key, []float32{0.2}, "result2", time.Minute)

	// Check internal storage - should only have 1 item
	store.mu.RLock()
	assert.Equal(t, 1, len(store.items[key]))
	store.mu.RUnlock()

	// Verify the remaining item is result2
	res, _, found := store.Search(context.Background(), key, []float32{0.2})
	assert.True(t, found)
	assert.Equal(t, "result2", res)
}

func TestCosineSimilarity_EdgeCases(t *testing.T) {
	// Empty vectors
	score := cosineSimilarityOptimized([]float32{}, []float32{}, 0, 0)
	assert.Equal(t, float32(0), score)

	// Mismatched lengths
	score = cosineSimilarityOptimized([]float32{1}, []float32{1, 2}, 1, 1)
	assert.Equal(t, float32(0), score)

	// Zero vectors (norm is 0)
	score = cosineSimilarityOptimized([]float32{0, 0}, []float32{0, 0}, 0, 0)
	assert.Equal(t, float32(0), score)
}
