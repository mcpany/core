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
	tests := []struct {
		name     string
		apiKey   string
		baseURL  string
		expected string
	}{
		{
			name:     "Default URL",
			apiKey:   "test-key",
			baseURL:  "",
			expected: "https://api.openai.com/v1",
		},
		{
			name:     "Custom URL",
			apiKey:   "test-key",
			baseURL:  "http://localhost:8080",
			expected: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.baseURL)
			assert.Equal(t, tt.expected, client.baseURL)
			assert.Equal(t, tt.apiKey, client.apiKey)
			assert.NotNil(t, client.client)
		})
	}
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name         string
		request      ChatRequest
		mockHandler  func(t *testing.T, w http.ResponseWriter, r *http.Request)
		ctxTimeout   time.Duration
		expectedResp *ChatResponse
		expectedErr  string
	}{
		{
			name: "Happy Path",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

				var req openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&req)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", req.Model)
				assert.Equal(t, "Hello", req.Messages[0].Content)

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
			},
			expectedResp: &ChatResponse{
				Content: "Hi there!",
			},
		},
		{
			name: "API Error (Non-200)",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte("Unauthorized"))
			},
			expectedErr: "openai api error (status 401): Unauthorized",
		},
		{
			name: "OpenAI Logic Error",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Invalid model",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "openai error: Invalid model",
		},
		{
			name: "Invalid JSON Response",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("{invalid-json"))
			},
			expectedErr: "failed to decode response",
		},
		{
			name: "Empty Choices",
			request: ChatRequest{
				Model: "gpt-4",
			},
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErr: "no choices returned",
		},
		{
			name:        "Network Error",
			request:     ChatRequest{Model: "gpt-4"},
			mockHandler: nil, // Indicates no server, bad URL
			expectedErr: "request failed",
		},
		{
			name:       "Context Timeout",
			request:    ChatRequest{Model: "gpt-4"},
			ctxTimeout: 1 * time.Millisecond,
			mockHandler: func(t *testing.T, w http.ResponseWriter, r *http.Request) {
				time.Sleep(50 * time.Millisecond)
			},
			expectedErr: "request failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var baseURL string
			if tt.mockHandler != nil {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					tt.mockHandler(t, w, r)
				}))
				defer server.Close()
				baseURL = server.URL
			} else {
				// Use an invalid port to force connection error
				baseURL = "http://127.0.0.1:0"
			}

			client := NewOpenAIClient("test-key", baseURL)

			ctx := context.Background()
			if tt.ctxTimeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.ctxTimeout)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.request)

			if tt.expectedErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
		})
	}
}
