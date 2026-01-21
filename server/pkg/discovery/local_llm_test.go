/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

package discovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOllamaProvider_Discover(t *testing.T) {
	// Mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/tags" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	provider := &OllamaProvider{Endpoint: server.URL}
	assert.Equal(t, "ollama", provider.Name())

	discovered, err := provider.Discover(context.Background())
	assert.NoError(t, err)
	assert.Len(t, discovered, 1)
	assert.Equal(t, "Local Ollama", discovered[0].GetName())
	assert.Equal(t, "v1", discovered[0].GetVersion())
	assert.Equal(t, server.URL+"/v1", discovered[0].GetHttpService().GetAddress())
}

func TestOllamaProvider_Discover_NotFound(t *testing.T) {
	provider := &OllamaProvider{Endpoint: "http://localhost:54321"} // Random unused port
	discovered, err := provider.Discover(context.Background())
	assert.Error(t, err)
	assert.Nil(t, discovered)
}

func TestLMStudioProvider_Discover(t *testing.T) {
	// Mock LM Studio server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/models" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	provider := &LMStudioProvider{Endpoint: server.URL}
	assert.Equal(t, "lm-studio", provider.Name())

	discovered, err := provider.Discover(context.Background())
	assert.NoError(t, err)
	assert.Len(t, discovered, 1)
	assert.Equal(t, "Local LM Studio", discovered[0].GetName())
	assert.Equal(t, "v1", discovered[0].GetVersion())
	assert.Equal(t, server.URL+"/v1", discovered[0].GetHttpService().GetAddress())
}

func TestLMStudioProvider_Discover_NotFound(t *testing.T) {
	provider := &LMStudioProvider{Endpoint: "http://localhost:54321"} // Random unused port
	discovered, err := provider.Discover(context.Background())
	assert.Error(t, err)
	assert.Nil(t, discovered)
}
