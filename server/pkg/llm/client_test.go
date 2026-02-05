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

func TestChatCompletion(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		expectedContent := "Hello world"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "/chat/completions", r.URL.Path)
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var req openAIChatRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "gpt-4", req.Model)
			assert.Len(t, req.Messages, 1)
			assert.Equal(t, "user", req.Messages[0].Role)
			assert.Equal(t, "Hi", req.Messages[0].Content)

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"content": expectedContent,
						},
					},
				},
			})
		}))
		defer server.Close()

		client := NewOpenAIClient("test-api-key", server.URL)
		resp, err := client.ChatCompletion(context.Background(), ChatRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: "user", Content: "Hi"},
			},
		})
		require.NoError(t, err)
		assert.Equal(t, expectedContent, resp.Content)
	})

	t.Run("NetworkError", func(t *testing.T) {
		// Use a closed port/invalid URL to force connection error
		client := NewOpenAIClient("key", "http://127.0.0.1:0")
		_, err := client.ChatCompletion(context.Background(), ChatRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request failed")
	})

	t.Run("APIError_500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := NewOpenAIClient("key", server.URL)
		_, err := client.ChatCompletion(context.Background(), ChatRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "status 500")
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`{invalid json`))
		}))
		defer server.Close()

		client := NewOpenAIClient("key", server.URL)
		_, err := client.ChatCompletion(context.Background(), ChatRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("OpenAIErrorField", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{
					"message": "Model not found",
				},
			})
		}))
		defer server.Close()

		client := NewOpenAIClient("key", server.URL)
		_, err := client.ChatCompletion(context.Background(), ChatRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "openai error: Model not found")
	})

	t.Run("EmptyChoices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []interface{}{},
			})
		}))
		defer server.Close()

		client := NewOpenAIClient("key", server.URL)
		_, err := client.ChatCompletion(context.Background(), ChatRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no choices returned")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Delay to ensure context cancels first
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewOpenAIClient("key", server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.ChatCompletion(ctx, ChatRequest{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
