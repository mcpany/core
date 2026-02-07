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
)

func TestNewOpenAIClient(t *testing.T) {
	t.Run("DefaultBaseURL", func(t *testing.T) {
		client := NewOpenAIClient("test-key", "")
		if client.baseURL != "https://api.openai.com/v1" {
			t.Errorf("expected default base URL, got %s", client.baseURL)
		}
	})

	t.Run("CustomBaseURL", func(t *testing.T) {
		customURL := "https://custom.openai.com/v1"
		client := NewOpenAIClient("test-key", customURL)
		if client.baseURL != customURL {
			t.Errorf("expected custom base URL %s, got %s", customURL, client.baseURL)
		}
	})
}

func TestChatCompletion_HappyPath(t *testing.T) {
	apiKey := "test-api-key"
	expectedContent := "Hello, world!"
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hi"},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify URL
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected path /chat/completions, got %s", r.URL.Path)
		}

		// Verify Method
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		// Verify Headers
		if r.Header.Get("Authorization") != "Bearer "+apiKey {
			t.Errorf("expected Authorization header Bearer %s, got %s", apiKey, r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Verify Body
		var decodedReq openAIChatRequest
		if err := json.NewDecoder(r.Body).Decode(&decodedReq); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if decodedReq.Model != req.Model {
			t.Errorf("expected model %s, got %s", req.Model, decodedReq.Model)
		}
		if len(decodedReq.Messages) != len(req.Messages) {
			t.Errorf("expected %d messages, got %d", len(req.Messages), len(decodedReq.Messages))
		}

		// Return Response
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
						Content: expectedContent,
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewOpenAIClient(apiKey, server.URL)
	resp, err := client.ChatCompletion(context.Background(), req)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Content != expectedContent {
		t.Errorf("expected content %s, got %s", expectedContent, resp.Content)
	}
}

func TestChatCompletion_ErrorScenarios(t *testing.T) {
	apiKey := "test-api-key"
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hi"},
		},
	}

	t.Run("NetworkError", func(t *testing.T) {
		// Use an invalid URL to simulate network error
		client := NewOpenAIClient(apiKey, "http://invalid-url-that-does-not-exist")
		_, err := client.ChatCompletion(context.Background(), req)
		if err == nil {
			t.Error("expected network error, got nil")
		}
	})

	t.Run("HTTPErrorStatus", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		}))
		defer server.Close()

		client := NewOpenAIClient(apiKey, server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		if err == nil {
			t.Error("expected error, got nil")
		}
		// Error message should contain the status code
		if err != nil && !strings.Contains(err.Error(), "500") {
			t.Errorf("expected error to contain 500, got %v", err)
		}
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("invalid json"))
		}))
		defer server.Close()

		client := NewOpenAIClient(apiKey, server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		if err == nil {
			t.Error("expected error, got nil")
		}
	})

	t.Run("APIErrorResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openAIChatResponse{
				Error: &struct {
					Message string `json:"message"`
				}{
					Message: "Something went wrong",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient(apiKey, server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "Something went wrong") {
			t.Errorf("expected error message 'Something went wrong', got %v", err)
		}
	})

	t.Run("EmptyChoices", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := openAIChatResponse{
				Choices: []struct {
					Message struct {
						Content string `json:"content"`
					} `json:"message"`
				}{}, // Empty slice
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client := NewOpenAIClient(apiKey, server.URL)
		_, err := client.ChatCompletion(context.Background(), req)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "no choices returned") {
			t.Errorf("expected error 'no choices returned', got %v", err)
		}
	})
}
