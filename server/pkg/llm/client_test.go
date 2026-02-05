// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

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

func TestNewOpenAIClient(t *testing.T) {
	client := NewOpenAIClient("test-key", "")
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
	assert.NotNil(t, client.client)

	clientCustom := NewOpenAIClient("test-key", "https://custom.api/v1")
	assert.Equal(t, "https://custom.api/v1", clientCustom.baseURL)
}

func TestChatCompletion_HappyPath(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req openAIChatRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", req.Model)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, "user", req.Messages[0].Role)
		assert.Equal(t, "hello", req.Messages[0].Content)

		resp := openAIChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{
					Message: struct {
						Content string `json:"content"`
					}{
						Content: "Hi there!",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)
	resp, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	})

	require.NoError(t, err)
	assert.Equal(t, "Hi there!", resp.Content)
}

func TestChatCompletion_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			Error: &struct {
				Message string `json:"message"`
			}{
				Message: "Rate limit exceeded",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)
	_, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openai error: Rate limit exceeded")
}

func TestChatCompletion_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)
	_, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "openai api error (status 401)")
}

func TestChatCompletion_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)
	_, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestChatCompletion_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := openAIChatResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)
	_, err := client.ChatCompletion(context.Background(), ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no choices returned")
}

func TestChatCompletion_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Wait longer than context
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := client.ChatCompletion(ctx, ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "hello"},
		},
	})

	assert.Error(t, err)
	// Exact error message depends on Go version and platform, but should be context deadline exceeded or request canceled
	// assert.Contains(t, err.Error(), "context deadline exceeded")
}
