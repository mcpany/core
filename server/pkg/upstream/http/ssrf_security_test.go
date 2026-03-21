// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestSSRFProtection(t *testing.T) {
	// 1. Start a local server (target for SSRF)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("secret data"))
	}))
	defer server.Close()

	// 2. Configure pool to point to it
	// Note: httptest server listens on loopback
	config := configv1.UpstreamServiceConfig_builder{
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
	}.Build()

	// 3. Create pool
	// Ensure env vars are cleared so we test default secure behavior
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "")
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "")

	p, err := NewHTTPPool(1, 1, 10, config)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	clientWrapper, err := p.Get(context.Background())
	require.NoError(t, err)

	// 4. Try to make a request
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := clientWrapper.Do(req)

	// 5. Assert failure
	// We expect an error because loopback should be blocked by SafeDialer
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	}
	if resp != nil {
		_ = resp.Body.Close()
	}
}

func TestSSRFProtection_Allowed(t *testing.T) {
	// 1. Start a local server (target for SSRF)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("secret data"))
	}))
	defer server.Close()

	// 2. Configure pool to point to it
	config := configv1.UpstreamServiceConfig_builder{
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String(server.URL),
		}.Build(),
	}.Build()

	// 3. Allow loopback via env var
	// 3. Allow loopback via env var
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	p, err := NewHTTPPool(1, 1, 10, config)
	require.NoError(t, err)
	defer func() { _ = p.Close() }()

	clientWrapper, err := p.Get(context.Background())
	require.NoError(t, err)

	// 4. Try to make a request
	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := clientWrapper.Do(req)

	// 5. Assert success
	assert.NoError(t, err)
	if resp != nil {
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}
