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
	tests := []struct {
		name           string
		request        ChatRequest
		handler        http.HandlerFunc
		expectedResp   *ChatResponse
		expectedErrStr string
	}{
		{
			name: "Happy Path",
			request: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/chat/completions", r.URL.Path)
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "Bearer test-api-key", r.Header.Get("Authorization"))
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

				var reqBody openAIChatRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				assert.NoError(t, err)
				assert.Equal(t, "gpt-4", reqBody.Model)
				assert.Equal(t, "user", reqBody.Messages[0].Role)
				assert.Equal(t, "Hello", reqBody.Messages[0].Content)

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
			name: "API Error (500)",
			request: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			expectedErrStr: "openai api error (status 500): Internal Server Error",
		},
		{
			name: "OpenAI Error Field",
			request: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Invalid model",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErrStr: "openai error: Invalid model",
		},
		{
			name: "Empty Choices",
			request: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Choices: []struct {
						Message struct {
							Content string `json:"content"`
						} `json:"message"`
					}{},
				}
				json.NewEncoder(w).Encode(resp)
			},
			expectedErrStr: "no choices returned",
		},
		{
			name: "Malformed Response",
			request: ChatRequest{
				Model: "gpt-4",
			},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("{invalid-json"))
			},
			expectedErrStr: "failed to decode response",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(tc.handler)
			defer server.Close()

			client := NewOpenAIClient("test-api-key", server.URL)
			resp, err := client.ChatCompletion(context.Background(), tc.request)

			if tc.expectedErrStr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErrStr)
				assert.Nil(t, resp)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}
		})
	}
}

func TestOpenAIClient_ChatCompletion_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewOpenAIClient("test-api-key", server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	resp, err := client.ChatCompletion(ctx, ChatRequest{Model: "gpt-4"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "context canceled")
	assert.Nil(t, resp)
}
