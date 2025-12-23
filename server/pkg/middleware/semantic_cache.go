// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"
)

// EmbeddingProvider defines the interface for fetching text embeddings.
type EmbeddingProvider interface {
	Embed(ctx context.Context, text string) ([]float32, error)
}

// SemanticCache implements a semantic cache using embeddings and cosine similarity.
type SemanticCache struct {
	provider  EmbeddingProvider
	store     *SimpleVectorStore
	threshold float32
}

// NewSemanticCache creates a new SemanticCache.
func NewSemanticCache(provider EmbeddingProvider, threshold float32) *SemanticCache {
	if threshold <= 0 {
		threshold = 0.9 // Default high threshold
	}
	return &SemanticCache{
		provider:  provider,
		store:     NewSimpleVectorStore(),
		threshold: threshold,
	}
}

// Get attempts to find a semantically similar cached result.
// It returns the result, the computed embedding, a boolean indicating a hit, and an error.
func (c *SemanticCache) Get(ctx context.Context, key string, input string) (any, []float32, bool, error) {
	embedding, err := c.provider.Embed(ctx, input)
	if err != nil {
		return nil, nil, false, err
	}

	result, score, found := c.store.Search(key, embedding)
	if found && score >= c.threshold {
		return result, embedding, true, nil
	}
	return nil, embedding, false, nil
}

// Set adds a result to the cache using the provided embedding.
func (c *SemanticCache) Set(ctx context.Context, key string, embedding []float32, result any, ttl time.Duration) error {
	c.store.Add(key, embedding, result, ttl)
	return nil
}

// SimpleVectorStore is a naive in-memory vector store.
type SimpleVectorStore struct {
	mu         sync.RWMutex
	items      map[string][]*VectorEntry
	maxEntries int
}

type VectorEntry struct {
	Vector    []float32
	Result    any
	ExpiresAt time.Time
}

func NewSimpleVectorStore() *SimpleVectorStore {
	return &SimpleVectorStore{
		items:      make(map[string][]*VectorEntry),
		maxEntries: 100, // Limit per key to prevent OOM
	}
}

func (s *SimpleVectorStore) Add(key string, vector []float32, result any, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cleanup(key)

	entries := s.items[key]
	if len(entries) >= s.maxEntries {
		// Evict oldest (FIFO)
		entries = entries[1:]
	}

	entry := &VectorEntry{
		Vector:    vector,
		Result:    result,
		ExpiresAt: time.Now().Add(ttl),
	}
	s.items[key] = append(entries, entry)
}

func (s *SimpleVectorStore) Search(key string, query []float32) (any, float32, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entries, ok := s.items[key]
	if !ok {
		return nil, 0, false
	}

	now := time.Now()
	var bestResult any
	var bestScore float32 = -1.0

	for _, entry := range entries {
		if now.After(entry.ExpiresAt) {
			continue
		}
		score := cosineSimilarity(query, entry.Vector)
		if score > bestScore {
			bestScore = score
			bestResult = entry.Result
		}
	}

	if bestScore == -1.0 {
		return nil, 0, false
	}

	return bestResult, bestScore, true
}

func (s *SimpleVectorStore) cleanup(key string) {
	entries, ok := s.items[key]
	if !ok {
		return
	}
	now := time.Now()
	// Filter in place
	n := 0
	for _, e := range entries {
		if now.Before(e.ExpiresAt) {
			entries[n] = e
			n++
		}
	}
	// Zero out the rest to help GC
	for i := n; i < len(entries); i++ {
		entries[i] = nil
	}
	s.items[key] = entries[:n]
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) != len(b) || len(a) == 0 {
		return 0
	}
	var dotProduct, normA, normB float32
	for i := range a {
		vA := a[i]
		vB := b[i]
		dotProduct += vA * vB
		normA += vA * vA
		normB += vB * vB
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (float32(math.Sqrt(float64(normA))) * float32(math.Sqrt(float64(normB))))
}

// OpenAIEmbeddingProvider implements EmbeddingProvider for OpenAI.
type OpenAIEmbeddingProvider struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIEmbeddingProvider(apiKey, model string) *OpenAIEmbeddingProvider {
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &OpenAIEmbeddingProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

type openAIEmbeddingRequest struct {
	Input          string `json:"input"`
	Model          string `json:"model"`
	EncodingFormat string `json:"encoding_format"`
}

type openAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *OpenAIEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	reqBody := openAIEmbeddingRequest{
		Input:          text,
		Model:          p.model,
		EncodingFormat: "float",
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai api error (status %d): %s", resp.StatusCode, string(body))
	}

	var response openAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Error != nil {
		return nil, fmt.Errorf("openai error: %s", response.Error.Message)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embedding data returned")
	}

	return response.Data[0].Embedding, nil
}
