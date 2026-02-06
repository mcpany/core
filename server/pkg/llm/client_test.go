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
	client := NewOpenAIClient("test-key", "https://custom.api.com")
	assert.Equal(t, "test-key", client.apiKey)
	assert.Equal(t, "https://custom.api.com", client.baseURL)
	assert.NotNil(t, client.client)
	assert.Equal(t, 30*time.Second, client.client.Timeout)

	clientDefault := NewOpenAIClient("test-key", "")
	assert.Equal(t, "https://api.openai.com/v1", clientDefault.baseURL)
}

func TestOpenAIClient_ChatCompletion_Success(t *testing.T) {
	// Mock OpenAI API Response
	mockResponse := openAIChatResponse{
		Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{
			{
				Message: struct {
					Content string `json:"content"`
				}{
					Content: "Hello, world!",
				},
			},
		},
	}
	respBytes, _ := json.Marshal(mockResponse)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Request Headers
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		// Verify Request Body
		var req openAIChatRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "gpt-4", req.Model)
		assert.Len(t, req.Messages, 1)
		assert.Equal(t, "user", req.Messages[0].Role)
		assert.Equal(t, "Hello", req.Messages[0].Content)

		// Send Response
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}))
	defer ts.Close()

	client := NewOpenAIClient("test-key", ts.URL)
	ctx := context.Background()
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Hello, world!", resp.Content)
}

func TestOpenAIClient_ChatCompletion_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer ts.Close()

	client := NewOpenAIClient("test-key", ts.URL)
	ctx := context.Background()
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "status 500")
	assert.Contains(t, err.Error(), "Internal Server Error")
}

func TestOpenAIClient_ChatCompletion_OpenAIError(t *testing.T) {
	mockResponse := openAIChatResponse{
		Error: &struct {
			Message string `json:"message"`
		}{
			Message: "Rate limit exceeded",
		},
	}
	respBytes, _ := json.Marshal(mockResponse)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}))
	defer ts.Close()

	client := NewOpenAIClient("test-key", ts.URL)
	ctx := context.Background()
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "openai error: Rate limit exceeded")
}

func TestOpenAIClient_ChatCompletion_EmptyChoices(t *testing.T) {
	mockResponse := openAIChatResponse{
		Choices: []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		}{}, // Empty slice
	}
	respBytes, _ := json.Marshal(mockResponse)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(respBytes)
	}))
	defer ts.Close()

	client := NewOpenAIClient("test-key", ts.URL)
	ctx := context.Background()
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "no choices returned")
}

func TestOpenAIClient_ChatCompletion_MalformedResponse(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{invalid-json"))
	}))
	defer ts.Close()

	client := NewOpenAIClient("test-key", ts.URL)
	ctx := context.Background()
	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to decode response")
}

func TestOpenAIClient_ChatCompletion_ContextCancellation(t *testing.T) {
	// Server that sleeps longer than the context timeout
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	client := NewOpenAIClient("test-key", ts.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req := ChatRequest{
		Model: "gpt-4",
		Messages: []Message{
			{Role: "user", Content: "Hello"},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	require.Error(t, err)
	assert.Nil(t, resp)
	// The error message depends on the OS/Go version, but it should be context deadline exceeded or similar
	assert.Contains(t, err.Error(), "request failed")
}
