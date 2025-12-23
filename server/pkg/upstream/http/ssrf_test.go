// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPPool_SSRFProtection(t *testing.T) {
	// Start a local server (loopback)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	// Configure pool
	config := &configv1.UpstreamServiceConfig{}

	// Create the pool
	pool, err := NewHTTPPool(1, 1, 10*time.Second, config)
	require.NoError(t, err)
	defer func() { _ = pool.Close() }()

	client, err := pool.Get(context.Background())
	require.NoError(t, err)

	// Attempt to connect to the local server
	// Since SafeDialer defaults to blocking loopback, this SHOULD fail if SSRF protection is enabled.
	// Currently (before fix), it should succeed.
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)

	// We expect an error due to SSRF protection (loopback blocked)
	assert.Error(t, err, "Should have blocked connection to loopback address")
	if err == nil {
		_ = resp.Body.Close()
	} else {
		assert.Contains(t, err.Error(), "ssrf attempt blocked", "Error message should indicate SSRF block")
	}
}

func TestHTTPPool_SSRFProtection_AllowLoopback(t *testing.T) {
	// Set env var to allow loopback
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Start a local server (loopback)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	// Configure pool
	config := &configv1.UpstreamServiceConfig{}

	// Create the pool - it should pick up the env var
	pool, err := NewHTTPPool(1, 1, 10*time.Second, config)
	require.NoError(t, err)
	defer func() { _ = pool.Close() }()

	client, err := pool.Get(context.Background())
	require.NoError(t, err)

	// Attempt to connect to the local server
	// This SHOULD succeed because we enabled loopback resources
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := client.Do(req)
	assert.NoError(t, err, "Should allow connection to loopback address when env var is set")
	if err == nil {
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}
