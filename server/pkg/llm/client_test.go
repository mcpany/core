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

func TestOpenAIClient_ChatCompletion(t *testing.T) {
	t.Run("HappyPath", func(t *testing.T) {
		// Mock the OpenAI API
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Validate Request
			assert.Equal(t, "/chat/completions", r.URL.Path)
			assert.Equal(t, "POST", r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
			assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))

			var req openAIChatRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, "gpt-4", req.Model)
			assert.Equal(t, "user", req.Messages[0].Role)
			assert.Equal(t, "Hello", req.Messages[0].Content)

			// Send Response
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
							Content: "Hello there!",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-api-key", ts.URL)
		req := ChatRequest{
			Model: "gpt-4",
			Messages: []Message{
				{Role: "user", Content: "Hello"},
			},
		}

		resp, err := client.ChatCompletion(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, "Hello there!", resp.Content)
	})

	t.Run("HTTPError", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-api-key", ts.URL)
		req := ChatRequest{Model: "gpt-4"}

		resp, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "openai api error (status 500)")
	})

	t.Run("MalformedJSONResponse", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("{invalid json"))
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-api-key", ts.URL)
		req := ChatRequest{Model: "gpt-4"}

		resp, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "failed to decode response")
	})

	t.Run("APIErrorResponse", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK) // 200 OK but with error field
			resp := openAIChatResponse{
				Error: &struct {
					Message string `json:"message"`
				}{
					Message: "Rate limit exceeded",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-api-key", ts.URL)
		req := ChatRequest{Model: "gpt-4"}

		resp, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "openai error: Rate limit exceeded")
	})

	t.Run("EmptyChoices", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openAIChatResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{}, // Empty
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-api-key", ts.URL)
		req := ChatRequest{Model: "gpt-4"}

		resp, err := client.ChatCompletion(context.Background(), req)
		require.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "no choices returned")
	})

	t.Run("ContextCancellation", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond) // Delay response
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-api-key", ts.URL)
		req := ChatRequest{Model: "gpt-4"}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		resp, err := client.ChatCompletion(ctx, req)
		require.Error(t, err)
		assert.Nil(t, resp)
		// Error message might vary depending on exact timing/OS, but should be context deadline
		assert.Contains(t, err.Error(), "context deadline exceeded")
	})
}
