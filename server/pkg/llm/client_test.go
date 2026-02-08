// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0


package llm_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOpenAIClient(t *testing.T) {
	t.Run("defaults", func(t *testing.T) {
		client := llm.NewOpenAIClient("test-key", "")
		assert.NotNil(t, client)
	})

	t.Run("custom base url", func(t *testing.T) {
		client := llm.NewOpenAIClient("test-key", "http://custom-url")
		assert.NotNil(t, client)
	})
}

func TestChatCompletion(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/chat/completions", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

			var req map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			assert.Equal(t, "gpt-4", req["model"])

			w.WriteHeader(http.StatusOK)
			// Matches openAIChatResponse structure
			response := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"content": "Hello world",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		client := llm.NewOpenAIClient("test-key", ts.URL)
		resp, err := client.ChatCompletion(ctx, llm.ChatRequest{
			Model: "gpt-4",
			Messages: []llm.Message{
				{Role: "user", Content: "Hi"},
			},
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Hello world", resp.Content)
	})

	t.Run("api error status", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request"))
		}))
		defer ts.Close()

		client := llm.NewOpenAIClient("test-key", ts.URL)
		resp, err := client.ChatCompletion(ctx, llm.ChatRequest{Model: "gpt-4"})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "openai api error (status 400)")
		assert.Contains(t, err.Error(), "invalid request")
	})

	t.Run("response decode error", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{invalid-json"))
		}))
		defer ts.Close()

		client := llm.NewOpenAIClient("test-key", ts.URL)
		resp, err := client.ChatCompletion(ctx, llm.ChatRequest{Model: "gpt-4"})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("openai error field", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Model overloaded",
				},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		client := llm.NewOpenAIClient("test-key", ts.URL)
		resp, err := client.ChatCompletion(ctx, llm.ChatRequest{Model: "gpt-4"})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "openai error: Model overloaded")
	})

	t.Run("no choices returned", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"choices": []interface{}{},
			}
			json.NewEncoder(w).Encode(response)
		}))
		defer ts.Close()

		client := llm.NewOpenAIClient("test-key", ts.URL)
		resp, err := client.ChatCompletion(ctx, llm.ChatRequest{Model: "gpt-4"})

		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no choices returned")
	})

	t.Run("context cancellation", func(t *testing.T) {
		// Create a server that sleeps
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		client := llm.NewOpenAIClient("test-key", ts.URL)

		ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer cancel()

		resp, err := client.ChatCompletion(ctx, llm.ChatRequest{Model: "gpt-4"})
		require.Error(t, err)
		assert.Nil(t, resp)
		// Error message might vary depending on OS/platform (context deadline exceeded or connection error),
		// but we just assert it's an error.
	})
}
