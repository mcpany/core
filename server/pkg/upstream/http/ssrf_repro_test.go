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

func TestSSRF_Vulnerability(t *testing.T) {
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
	// Note: We are using the same client for both tests.
	// Since the first request fails to dial, no connection is pooled.

	// 3. Attempt to connect to the local server (Default behavior: Blocked)
	req, _ := http.NewRequest("GET", ts.URL, nil)
	_, err = client.Do(req)

	// 4. Verification: Should fail with SSRF block
	require.Error(t, err, "Should be blocked by SSRF protection")
	require.Contains(t, err.Error(), "ssrf attempt blocked", "Error message should indicate SSRF block")

	// 5. Allow via Environment Variable
	t.Setenv("MCPANY_ALLOW_LOOPBACK", "true")

	// 6. Attempt to connect again (Should succeed)
	req2, _ := http.NewRequest("GET", ts.URL, nil)
	resp2, err2 := client.Do(req2)

	require.NoError(t, err2, "Should be allowed when MCPANY_ALLOW_LOOPBACK is set")
	require.Equal(t, 200, resp2.StatusCode)
}
