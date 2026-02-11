// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSRFProtection(t *testing.T) {
	// Start a local server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer ts.Close()

	// 1. Default behavior: Block local IPs
	t.Run("BlockLocalIPs", func(t *testing.T) {
		// Ensure environment variables are unset
		os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
		os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
		os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

		client := NewSafeHTTPClient()
		// Reduce timeout for test speed
		client.Timeout = 1 * time.Second

		req, err := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		// We expect an error because ts.URL is localhost (loopback)
		assert.Error(t, err, "Should fail to connect to loopback address")
		if err != nil {
			assert.Contains(t, err.Error(), "ssrf attempt blocked", "Error message should mention SSRF block")
		}
		if resp != nil {
			resp.Body.Close()
		}
	})

	// 2. Allow Loopback behavior
	t.Run("AllowLoopback", func(t *testing.T) {
		os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
		defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

		client := NewSafeHTTPClient()
		client.Timeout = 1 * time.Second

		req, err := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
		require.NoError(t, err)

		resp, err := client.Do(req)
		assert.NoError(t, err, "Should connect to loopback address when allowed")
		if resp != nil {
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			resp.Body.Close()
		}
	})
}
