// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	httppkg "github.com/mcpany/core/pkg/upstream/http"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSSRF_Protection(t *testing.T) {
	// 1. Setup a local server (simulating a sensitive internal service)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("secret data"))
	}))
	defer ts.Close()

	// 2. Create the HTTP pool using the implementation
	config := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				TlsConfig: &configv1.TLSConfig{InsecureSkipVerify: proto.Bool(true)},
			},
		},
	}

	pool, err := httppkg.NewHTTPPool(1, 1, time.Second, config)
	require.NoError(t, err)

	client, err := pool.Get(context.Background())
	require.NoError(t, err)

	// 3. Attempt to connect to the local server (Loopback)
	req, _ := http.NewRequest("GET", ts.URL, nil)
	resp, err := client.Do(req)

	// 4. Verification: Should be BLOCKED (Loopback is blocked by default)
	require.Error(t, err, "Connection to localhost should be blocked by SSRF protection")
	require.Contains(t, err.Error(), "ssrf attempt blocked")
	if resp != nil {
		resp.Body.Close()
	}

	// 5. Verify bypass with Env Var
	t.Run("AllowLoopback", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK", "true")

		resp2, err2 := client.Do(req)
		require.NoError(t, err2, "Connection to localhost should be allowed when MCPANY_ALLOW_LOOPBACK=true")
		require.Equal(t, 200, resp2.StatusCode)
		resp2.Body.Close()
	})

	// 6. Verify Private IP behavior
	t.Run("PrivateIP_AllowedByDefault", func(t *testing.T) {
		// Default config: Private allowed
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		// 10.0.0.1 is reserved private IP, likely unreachable
		reqPrivate, _ := http.NewRequestWithContext(ctx, "GET", "http://10.0.0.1", nil)
		_, err := client.Do(reqPrivate)

		// Should NOT be ssrf blocked. Should be timeout or network unreachable.
		if err != nil {
			require.NotContains(t, err.Error(), "ssrf attempt blocked", "Private IPs should be allowed by default")
		}
	})

	t.Run("PrivateIP_BlockedWhenConfigured", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK", "false")

		// Re-dial
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		defer cancel()

		reqPrivate, _ := http.NewRequestWithContext(ctx, "GET", "http://10.0.0.1", nil)
		_, err := client.Do(reqPrivate)

		require.Error(t, err)
		require.Contains(t, err.Error(), "ssrf attempt blocked", "Private IPs should be blocked when configured")
	})
}
