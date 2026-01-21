// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaProvider_Name(t *testing.T) {
	p := &OllamaProvider{Endpoint: "http://localhost:11434"}
	assert.Equal(t, "ollama", p.Name())
}

func TestOllamaProvider_Discover(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		// Mock Ollama server
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/tags", r.URL.Path)
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		p := &OllamaProvider{Endpoint: ts.URL}
		configs, err := p.Discover(context.Background())
		require.NoError(t, err)
		require.Len(t, configs, 1)

		config := configs[0]
		assert.Equal(t, "Local Ollama", config.GetName())
		assert.Equal(t, "v1", config.GetVersion())
		assert.Equal(t, ts.URL+"/v1", config.GetHttpService().GetAddress())
		assert.Equal(t, []string{"local-llm", "ollama", "openai-compatible"}, config.Tags)
	})

	t.Run("failure_not_found", func(t *testing.T) {
		// Using a reserved domain that shouldn't resolve or a port that is likely closed on localhost
		// invalid.test is reserved by RFC 2606
		p := &OllamaProvider{Endpoint: "http://invalid.test:12345"}
		configs, err := p.Discover(context.Background())
		assert.Error(t, err)
		assert.Nil(t, configs)
		assert.Contains(t, err.Error(), "ollama not found")
	})

	t.Run("failure_status_code", func(t *testing.T) {
		// Mock Ollama server returning 500
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer ts.Close()

		p := &OllamaProvider{Endpoint: ts.URL}
		configs, err := p.Discover(context.Background())
		assert.Error(t, err)
		assert.Nil(t, configs)
		assert.Contains(t, err.Error(), "ollama returned status 500")
	})
}
