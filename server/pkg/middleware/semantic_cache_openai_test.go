// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// RoundTripFunc is a type for mocking http.RoundTripper
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip implements http.RoundTripper
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid network calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
		Timeout:   10 * time.Second,
	}
}

func TestNewOpenAIEmbeddingProvider(t *testing.T) {
	// Test default model
	p1 := NewOpenAIEmbeddingProvider("key", "")
	assert.Equal(t, "text-embedding-3-small", p1.model)
	assert.Equal(t, "key", p1.apiKey)
	assert.NotNil(t, p1.client)

	// Test custom model
	p2 := NewOpenAIEmbeddingProvider("key", "custom-model")
	assert.Equal(t, "custom-model", p2.model)
}

func TestOpenAIEmbeddingProvider_Embed_Success(t *testing.T) {
	mockResponse := openAIEmbeddingResponse{
		Data: []struct {
			Embedding []float32 `json:"embedding"`
		}{
			{
				Embedding: []float32{0.1, 0.2, 0.3},
			},
		},
	}
	respBytes, _ := json.Marshal(mockResponse)

	client := NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "POST", req.Method)
		assert.Equal(t, "https://api.openai.com/v1/embeddings", req.URL.String())
		assert.Equal(t, "Bearer test-key", req.Header.Get("Authorization"))
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

		// Check request body
		var body openAIEmbeddingRequest
		_ = json.NewDecoder(req.Body).Decode(&body)
		assert.Equal(t, "test input", body.Input)
		assert.Equal(t, "test-model", body.Model)

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer(respBytes)),
			Header:     make(http.Header),
		}
	})

	provider := NewOpenAIEmbeddingProvider("test-key", "test-model")
	provider.client = client // Inject mock client

	emb, err := provider.Embed(context.Background(), "test input")
	assert.NoError(t, err)
	assert.Equal(t, []float32{0.1, 0.2, 0.3}, emb)
}

func TestOpenAIEmbeddingProvider_Embed_Error(t *testing.T) {
	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 500,
			Body:       io.NopCloser(bytes.NewBufferString("Internal Server Error")),
			Header:     make(http.Header),
		}
	})

	provider := NewOpenAIEmbeddingProvider("test-key", "")
	provider.client = client

	emb, err := provider.Embed(context.Background(), "test")
	assert.Error(t, err)
	assert.Nil(t, emb)
	assert.Contains(t, err.Error(), "openai api error (status 500)")
}

func TestOpenAIEmbeddingProvider_Embed_APIError(t *testing.T) {
	mockError := openAIEmbeddingResponse{
		Error: &struct {
			Message string `json:"message"`
		}{
			Message: "Invalid API Key",
		},
	}
	respBytes, _ := json.Marshal(mockError)

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200, // API returns 200 but with error field? Usually it returns 4xx, but let's test logic
			Body:       io.NopCloser(bytes.NewBuffer(respBytes)),
			Header:     make(http.Header),
		}
	})

	provider := NewOpenAIEmbeddingProvider("test-key", "")
	provider.client = client

	emb, err := provider.Embed(context.Background(), "test")
	assert.Error(t, err)
	assert.Nil(t, emb)
	assert.Contains(t, err.Error(), "openai error: Invalid API Key")
}

func TestOpenAIEmbeddingProvider_Embed_NoData(t *testing.T) {
	mockResponse := openAIEmbeddingResponse{
		Data: []struct {
			Embedding []float32 `json:"embedding"`
		}{},
	}
	respBytes, _ := json.Marshal(mockResponse)

	client := NewTestClient(func(req *http.Request) *http.Response {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer(respBytes)),
			Header:     make(http.Header),
		}
	})

	provider := NewOpenAIEmbeddingProvider("test-key", "")
	provider.client = client

	emb, err := provider.Embed(context.Background(), "test")
	assert.Error(t, err)
	assert.Nil(t, emb)
	assert.Contains(t, err.Error(), "no embedding data returned")
}
