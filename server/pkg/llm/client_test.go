// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewOpenAIClient(t *testing.T) {
	client := NewOpenAIClient("test-key", "")
	if client.baseURL != "https://api.openai.com/v1" {
		t.Errorf("expected default base URL, got %s", client.baseURL)
	}
	if client.apiKey != "test-key" {
		t.Errorf("expected api key test-key, got %s", client.apiKey)
	}

	customClient := NewOpenAIClient("test-key", "http://custom-url")
	if customClient.baseURL != "http://custom-url" {
		t.Errorf("expected custom base URL, got %s", customClient.baseURL)
	}
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		handler        http.HandlerFunc
		req            ChatRequest
		expectedResult *ChatResponse
		expectError    bool
		errorContains  string
	}{
		{
			name: "Happy Path",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.Header.Get("Authorization") != "Bearer test-key" {
					t.Errorf("expected Authorization header, got %s", r.Header.Get("Authorization"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type header, got %s", r.Header.Get("Content-Type"))
				}

				var req openAIChatRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					t.Errorf("failed to decode request body: %v", err)
				}
				if req.Model != "gpt-4" {
					t.Errorf("expected model gpt-4, got %s", req.Model)
				}

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
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			expectedResult: &ChatResponse{
				Content: "Hello world",
			},
			expectError: false,
		},
		{
			name: "API Error (500)",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			req: ChatRequest{
				Model: "gpt-4",
			},
			expectError:   true,
			errorContains: "openai api error (status 500)",
		},
		{
			name: "OpenAI Specific Error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				resp := openAIChatResponse{
					Error: &struct {
						Message string `json:"message"`
					}{
						Message: "Rate limit exceeded",
					},
				}
				json.NewEncoder(w).Encode(resp)
			},
			req: ChatRequest{
				Model: "gpt-4",
			},
			expectError:   true,
			errorContains: "openai error: Rate limit exceeded",
		},
		{
			name: "Empty Choices",
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
			req: ChatRequest{
				Model: "gpt-4",
			},
			expectError:   true,
			errorContains: "no choices returned",
		},
		{
			name: "Malformed Response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("{invalid-json"))
			},
			req: ChatRequest{
				Model: "gpt-4",
			},
			expectError:   true,
			errorContains: "failed to decode response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)
			resp, err := client.ChatCompletion(context.Background(), tt.req)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("expected error containing %q, got %q", tt.errorContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if resp == nil {
					t.Error("expected response, got nil")
				} else if resp.Content != tt.expectedResult.Content {
					t.Errorf("expected content %q, got %q", tt.expectedResult.Content, resp.Content)
				}
			}
		})
	}
}

func TestChatCompletion_ContextCancellation(t *testing.T) {
	// Create a handler that simulates a long running request
	handler := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	client := NewOpenAIClient("test-key", server.URL)

	// Create a context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	req := ChatRequest{Model: "gpt-4"}
	_, err := client.ChatCompletion(ctx, req)

	if err == nil {
		t.Error("expected error due to context cancellation, got nil")
	}
}
