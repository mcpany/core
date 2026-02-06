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
	t.Run("default base URL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "")
		assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
		assert.NotNil(t, client.client)
		assert.Equal(t, 30*time.Second, client.client.Timeout)
	})

	t.Run("custom base URL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "https://custom.api.com")
		assert.Equal(t, "https://custom.api.com", client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
	})
}

func TestChatCompletion(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/chat/completions", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

			var req openAIChatRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "gpt-4", req.Model)
			assert.Len(t, req.Messages, 1)
			assert.Equal(t, "user", req.Messages[0].Role)
			assert.Equal(t, "hello", req.Messages[0].Content)

			// Use a map to generate the response JSON to decouple from internal structs
			response := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"content": "world",
						},
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)

		req := ChatRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: "user", Content: "hello"},
			},
		}

		resp, err := client.ChatCompletion(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "world", resp.Content)
	})

	t.Run("http error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)

		req := ChatRequest{Model: "gpt-4"}
		resp, err := client.ChatCompletion(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "openai api error (status 500)")
		assert.Contains(t, err.Error(), "internal error")
	})

	t.Run("api error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "invalid model",
				},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)

		req := ChatRequest{Model: "gpt-4"}
		resp, err := client.ChatCompletion(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "openai error: invalid model")
	})

	t.Run("empty choices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			response := map[string]interface{}{
				"choices": []interface{}{},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(response)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)

		req := ChatRequest{Model: "gpt-4"}
		resp, err := client.ChatCompletion(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no choices returned")
	})

	t.Run("invalid json response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{invalid"))
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)

		req := ChatRequest{Model: "gpt-4"}
		resp, err := client.ChatCompletion(context.Background(), req)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("context cancellation", func(t *testing.T) {
		// Create a server that sleeps to allow cancellation to trigger
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		req := ChatRequest{Model: "gpt-4"}
		resp, err := client.ChatCompletion(ctx, req)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
