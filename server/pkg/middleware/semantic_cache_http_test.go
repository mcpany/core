// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHttpEmbeddingProvider(t *testing.T) {
	tests := []struct {
		name             string
		url              string
		bodyTemplate     string
		responseJSONPath string
		expectError      bool
	}{
		{
			name:             "Valid config",
			url:              "http://example.com/embed",
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			expectError:      false,
		},
		{
			name:             "Missing URL",
			url:              "",
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			expectError:      true,
		},
		{
			name:             "Invalid template",
			url:              "http://example.com/embed",
			bodyTemplate:     `{"input": "{{.Input}"`, // Missing closing brace
			responseJSONPath: "$.embedding",
			expectError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider, err := NewHttpEmbeddingProvider(tt.url, nil, tt.bodyTemplate, tt.responseJSONPath)
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, provider)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, provider)
			}
		})
	}
}

func TestHttpEmbeddingProvider_Embed(t *testing.T) {
	tests := []struct {
		name             string
		handler          http.HandlerFunc
		bodyTemplate     string
		responseJSONPath string
		inputText        string
		expectedEmbed    []float32
		expectError      bool
		errorContains    string
	}{
		{
			name: "Success",
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
				// Check body
				var body map[string]string
				err := json.NewDecoder(r.Body).Decode(&body)
				require.NoError(t, err)
				assert.Equal(t, "hello", body["input"])

				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"embedding": [0.1, 0.2, 0.3]}}`))
			},
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.data.embedding",
			inputText:        "hello",
			expectedEmbed:    []float32{0.1, 0.2, 0.3},
			expectError:      false,
		},
		{
			name: "HTTP Error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			},
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			inputText:        "test",
			expectError:      true,
			errorContains:    "http api error",
		},
		{
			name: "Invalid JSON Response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`invalid json`))
			},
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			inputText:        "test",
			expectError:      true,
			errorContains:    "failed to unmarshal response json",
		},
		{
			name: "JSONPath Not Found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"other": "data"}`))
			},
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			inputText:        "test",
			expectError:      true,
			errorContains:    "jsonpath extraction failed",
		},
		{
			name: "Result Not Array",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"embedding": "not an array"}`))
			},
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			inputText:        "test",
			expectError:      true,
			errorContains:    "jsonpath result is not an array",
		},
		{
			name: "Array Contains Non-Number",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"embedding": [0.1, "string"]}`))
			},
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			inputText:        "test",
			expectError:      true,
			errorContains:    "is not a number",
		},
		{
			name: "Empty Embedding",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"embedding": []}`))
			},
			bodyTemplate:     `{"input": "{{.Input}}"}`,
			responseJSONPath: "$.embedding",
			inputText:        "test",
			expectError:      true,
			errorContains:    "empty embedding returned",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			headers := map[string]string{"Content-Type": "application/json"}
			provider, err := NewHttpEmbeddingProvider(server.URL, headers, tt.bodyTemplate, tt.responseJSONPath)
			require.NoError(t, err)

			embed, err := provider.Embed(context.Background(), tt.inputText)
			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedEmbed, embed)
			}
		})
	}
}

func TestHttpEmbeddingProvider_TemplateError(t *testing.T) {
	// Template that might fail execution?
	// It's hard to make text/template fail at execution time with a simple struct.
	// But let's verify headers are sent.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "custom-value", r.Header.Get("X-Custom-Header"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"e": [1.0]}`))
	}))
	defer server.Close()

	provider, err := NewHttpEmbeddingProvider(server.URL, map[string]string{"X-Custom-Header": "custom-value"}, `{{.Input}}`, "$.e")
	require.NoError(t, err)

	_, err = provider.Embed(context.Background(), "test")
	assert.NoError(t, err)
}
