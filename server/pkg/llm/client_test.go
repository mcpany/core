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
	tests := []struct {
		name    string
		apiKey  string
		baseURL string
		wantURL string
	}{
		{
			name:    "Default Base URL",
			apiKey:  "test-key",
			baseURL: "",
			wantURL: "https://api.openai.com/v1",
		},
		{
			name:    "Custom Base URL",
			apiKey:  "test-key",
			baseURL: "https://custom.api.com",
			wantURL: "https://custom.api.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewOpenAIClient(tt.apiKey, tt.baseURL)
			if client.apiKey != tt.apiKey {
				t.Errorf("NewOpenAIClient() apiKey = %v, want %v", client.apiKey, tt.apiKey)
			}
			if client.baseURL != tt.wantURL {
				t.Errorf("NewOpenAIClient() baseURL = %v, want %v", client.baseURL, tt.wantURL)
			}
			if client.client == nil {
				t.Error("NewOpenAIClient() client is nil")
			}
		})
	}
}

func TestChatCompletion(t *testing.T) {
	tests := []struct {
		name           string
		req            ChatRequest
		mockStatus     int
		mockBody       string
		mockDelay      time.Duration
		expectError    bool
		errorContains  string
		expectContent  string
		validateReq    func(*testing.T, *http.Request)
	}{
		{
			name: "Success",
			req: ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			mockStatus: http.StatusOK,
			mockBody: `{
				"choices": [
					{
						"message": {
							"content": "Hi there!"
						}
					}
				]
			}`,
			expectContent: "Hi there!",
			validateReq: func(t *testing.T, r *http.Request) {
				if r.Header.Get("Authorization") != "Bearer test-key" {
					t.Errorf("Authorization header = %v, want Bearer test-key", r.Header.Get("Authorization"))
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Content-Type header = %v, want application/json", r.Header.Get("Content-Type"))
				}
				var body openAIChatRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("Failed to decode request body: %v", err)
				}
				if body.Model != "gpt-4" {
					t.Errorf("Request model = %v, want gpt-4", body.Model)
				}
			},
		},
		{
			name: "API Error 401",
			req: ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusUnauthorized,
			mockBody:   `{"error": {"message": "Invalid API key"}}`,
			expectError: true,
			errorContains: "openai api error (status 401)",
		},
		{
			name: "API Error 500",
			req: ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusInternalServerError,
			mockBody:   `Internal Server Error`,
			expectError: true,
			errorContains: "openai api error (status 500)",
		},
		{
			name: "OpenAI Logic Error",
			req: ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockBody:   `{"error": {"message": "Model not found"}}`,
			expectError: true,
			errorContains: "openai error: Model not found",
		},
		{
			name: "Empty Choices",
			req: ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockBody:   `{"choices": []}`,
			expectError: true,
			errorContains: "no choices returned",
		},
		{
			name: "Invalid JSON Response",
			req: ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockBody:   `{invalid json`,
			expectError: true,
			errorContains: "failed to decode response",
		},
		{
			name: "Context Timeout",
			req: ChatRequest{Model: "gpt-4"},
			mockStatus: http.StatusOK,
			mockDelay: 100 * time.Millisecond,
			expectError: true,
			errorContains: "context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.validateReq != nil {
					tt.validateReq(t, r)
				}
				if tt.mockDelay > 0 {
					time.Sleep(tt.mockDelay)
				}
				w.WriteHeader(tt.mockStatus)
				w.Write([]byte(tt.mockBody))
			}))
			defer server.Close()

			client := NewOpenAIClient("test-key", server.URL)

			ctx := context.Background()
			if tt.mockDelay > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, 50*time.Millisecond)
				defer cancel()
			}

			resp, err := client.ChatCompletion(ctx, tt.req)

			if tt.expectError {
				if err == nil {
					t.Error("ChatCompletion() expected error, got nil")
				} else if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("ChatCompletion() error = %v, want substring %v", err, tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ChatCompletion() unexpected error: %v", err)
				} else if resp.Content != tt.expectContent {
					t.Errorf("ChatCompletion() content = %v, want %v", resp.Content, tt.expectContent)
				}
			}
		})
	}
}
