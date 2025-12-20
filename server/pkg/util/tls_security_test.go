// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPClientWithTLS_SSRF_Repro(t *testing.T) {
	// Start a local server (loopback)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Hello, client")
	}))
	defer ts.Close()

	t.Run("nil config blocks localhost by default", func(t *testing.T) {
		client, err := NewHTTPClientWithTLS(nil)
		require.NoError(t, err)

		_, err = client.Get(ts.URL)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("empty config blocks localhost by default", func(t *testing.T) {
		client, err := NewHTTPClientWithTLS(&configv1.TLSConfig{})
		require.NoError(t, err)

		_, err = client.Get(ts.URL)
		require.Error(t, err)
		require.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("allows localhost if env var is set", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK", "true")
		client, err := NewHTTPClientWithTLS(nil)
		require.NoError(t, err)

		resp, err := client.Get(ts.URL)
		require.NoError(t, err)
		if resp != nil {
			resp.Body.Close()
		}
	})
}
