// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSSRFPrevention(t *testing.T) {
	// Start a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	t.Run("DefaultBlockLoopback", func(t *testing.T) {
		// Ensure environment variable is unset
		os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
		os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

		client := util.NewSafeHTTPClient()
		// Reduce timeout for faster test failure if it tries to connect but hangs (shouldn't happen with block)
		client.Timeout = 2 * time.Second

		resp, err := client.Get(ts.URL)
		// It should fail because loopback is blocked
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
		if resp != nil {
			resp.Body.Close()
		}
	})

	t.Run("AllowLoopback", func(t *testing.T) {
		// Set environment variable to allow loopback
		os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
		defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

		client := util.NewSafeHTTPClient()
		client.Timeout = 2 * time.Second

		resp, err := client.Get(ts.URL)
		require.NoError(t, err)
		if resp != nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}
	})
}
