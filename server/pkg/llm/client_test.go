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
	// Common test data
	apiKey := "test-api-key"
	testModel := "gpt-4"
	testMessages := []Message{
		{Role: "user", Content: "Hello"},
	}
	req := ChatRequest{
		Model:    testModel,
		Messages: testMessages,
	}

	tests := []struct {
		name           string
		mockHandler    func(w http.ResponseWriter, r *http.Request)
		expectedResp   *ChatResponse
		expectedError  string
		contextTimeout time.Duration
	}{
		{
			name: "Happy Path",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer "+apiKey, r.Header.Get("Authorization"))

				var receivedReq openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&receivedReq)
				assert.NoError(t, err)
				assert.Equal(t, testModel, receivedReq.Model)
				assert.Equal(t, testMessages, receivedReq.Messages)

				// Send response
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
								Content: "Hello world",
							},
						},
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedResp: &ChatResponse{
				Content: "Hello world",
			},
			expectedError: "",
		},
		{
			name: "OpenAI Error Response",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Invalid API Key",
					},
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedResp:  nil,
			expectedError: "openai error: Invalid API Key",
		},
		{
			name: "Empty Choices",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{}, // Empty choices
				}
				w.WriteHeader(http.StatusOK)
				_ = json.NewEncoder(w).Encode(resp)
			},
			expectedResp:  nil,
			expectedError: "no choices returned",
		},
		{
			name: "Non-200 Status Code",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte("Unauthorized access"))
			},
			expectedResp:  nil,
			expectedError: "openai api error (status 401): Unauthorized access",
		},
		{
			name: "Invalid JSON Response",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("{invalid-json"))
			},
			expectedResp:  nil,
			expectedError: "failed to decode response",
		},
		{
			name: "Context Timeout",
			mockHandler: func(w http.ResponseWriter, r *http.Request) {
				time.Sleep(100 * time.Millisecond) // Simulate delay
				w.WriteHeader(http.StatusOK)
			},
			expectedResp:   nil,
			expectedError:  "context deadline exceeded",
			contextTimeout: 10 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(tt.mockHandler))
			defer server.Close()

			// Initialize client with mock server URL
			client := NewOpenAIClient(apiKey, server.URL)

			// Create context
			ctx := context.Background()
			if tt.contextTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.contextTimeout)
				defer cancel()
			}

			// Execute request
			resp, err := client.ChatCompletion(ctx, req)

			// Verify results
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}

func TestOpenAIClient_NetworkError(t *testing.T) {
	// Test network error by using a closed server URL or invalid URL
	client := NewOpenAIClient("test-key", "http://localhost:12345") // Invalid port/host

	ctx := context.Background()
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "request failed")
	assert.Nil(t, resp)
}

func TestNewOpenAIClient_DefaultURL(t *testing.T) {
	client := NewOpenAIClient("key", "")
	assert.Equal(t, "https://api.openai.com/v1", client.baseURL)
}
