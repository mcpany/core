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
	})

	t.Run("custom base URL", func(t *testing.T) {
		customURL := "https://custom.openai.com"
		client := NewOpenAIClient("test-key", customURL)
		assert.Equal(t, customURL, client.baseURL)
		assert.Equal(t, "test-key", client.apiKey)
	})
}

func TestChatCompletion(t *testing.T) {
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	t.Run("happy path", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/chat/completions", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

			var reqBody openAIChatRequest
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			assert.NoError(t, err)
			assert.Equal(t, "gpt-4", reqBody.Model)
			assert.Equal(t, 1, len(reqBody.Messages))

			resp := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]interface{}{
							"content": "Hi there!",
						},
					},
				},
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)
		resp, err := client.ChatCompletion(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "Hi there!", resp.Content)
	})

	t.Run("API error 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai api error (status 500)")
	})

	t.Run("malformed JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("{invalid-json"))
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("OpenAI business error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"error": map[string]string{
					"message": "Quota exceeded",
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai error: Quota exceeded")
	})

	t.Run("no choices returned", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			resp := map[string]interface{}{
				"choices": []interface{}{},
			}
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no choices returned")
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Simulate delay
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client := NewOpenAIClient("test-key", server.URL)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.ChatCompletion(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})
}
