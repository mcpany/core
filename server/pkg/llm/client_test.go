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

func TestOpenAIClient_ChatCompletion_HappyPath(t *testing.T) {
	// Mock OpenAI API Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request Headers
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)

		// Verify Request Body
		var reqBody openAIChatRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", reqBody.Model)
		assert.Len(t, reqBody.Messages, 1)
		assert.Equal(t, "user", reqBody.Messages[0].Role)
		assert.Equal(t, "Hello", reqBody.Messages[0].Content)

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
						Content: "Hi there!",
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Initialize Client
	client := NewOpenAIClient("test-api-key", server.URL)

	// Execute Request
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}
	resp, err := client.ChatCompletion(context.Background(), req)

	// Verify Result
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Hi there!", resp.Content)
}

func TestOpenAIClient_ChatCompletion_Errors(t *testing.T) {
	testCases := []struct {
		name          string
		handler       http.HandlerFunc
		expectedError string
	}{
		{
			name: "HTTP 401 Unauthorized",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
			},
			expectedError: "openai api error (status 401): Unauthorized",
		},
		{
			name: "HTTP 500 Internal Server Error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectedError: "openai api error (status 500): Internal Server Error",
		},
		{
			name: "Malformed JSON Response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("{ invalid json"))
			},
			expectedError: "failed to decode response",
		},
		{
			name: "OpenAI Error Structure",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{ "error": { "message": "Rate limit exceeded" } }`))
			},
			expectedError: "openai error: Rate limit exceeded",
		},
		{
			name: "No Choices Returned",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{ "choices": [] }`))
			},
			expectedError: "no choices returned",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			client := NewOpenAIClient("test-api-key", server.URL)
			req := ChatRequest{
				Model:    "gpt-4",
				Messages: []Message{{Role: "user", Content: "Hello"}},
			}

			resp, err := client.ChatCompletion(context.Background(), req)

			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

func TestOpenAIClient_ChatCompletion_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Simulate delay
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-api-key", server.URL)
	req := ChatRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	resp, err := client.ChatCompletion(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	// The error message for context cancellation can vary slightly but usually contains "context canceled"
	assert.Contains(t, err.Error(), "context canceled")
}

func TestOpenAIClient_Headers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer my-secret-key", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "ok"}}]}`))
	}))
	defer server.Close()

	client := NewOpenAIClient("my-secret-key", server.URL)
	req := ChatRequest{
		Model:    "gpt-4",
		Messages: []Message{{Role: "user", Content: "Hello"}},
	}

	_, err := client.ChatCompletion(context.Background(), req)
	assert.NoError(t, err)
}

func TestNewOpenAIClient_Defaults(t *testing.T) {
	client := NewOpenAIClient("key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)

	client2 := NewOpenAIClient("key", "http://custom.url")
	assert.Equal(t, "http://custom.url", client2.baseURL)
}
