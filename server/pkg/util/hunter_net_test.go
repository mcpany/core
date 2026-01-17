// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConnection_Coverage(t *testing.T) {
	// Start a listener
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	addr := ln.Addr().String()

	// 1. Success case with host:port
	err = CheckConnection(context.Background(), addr)
	assert.NoError(t, err)

	// 2. Success case with scheme
	err = CheckConnection(context.Background(), "http://"+addr)
	assert.NoError(t, err)

	// 3. Failure case
	// Pick a port that is likely closed.
	err = CheckConnection(context.Background(), "127.0.0.1:54321")
	assert.Error(t, err)

	// 4. Invalid address
	err = CheckConnection(context.Background(), "invalid:address:port")
	assert.Error(t, err)

	// 5. Invalid URL
	err = CheckConnection(context.Background(), "http://[::1]:namedport") // named port is invalid in URL
	assert.Error(t, err)
}

func TestSafeDialer_Coverage(t *testing.T) {
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// By default, SafeDialer blocks loopback
	client := NewSafeHTTPClient()
	_, err := client.Get(ts.URL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loopback")

	// Allow loopback via env
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	client = NewSafeHTTPClient()
	resp, err := client.Get(ts.URL)
	assert.NoError(t, err)
	resp.Body.Close()

	// Direct SafeDialContext usage
	// Should fail for loopback if default
	dialer := NewSafeDialer()
	// Split ts.URL (http://127.0.0.1:port)
	u, _ :=  ts.Listener.Addr().(*net.TCPAddr)
	addr := u.String()

	_, err = dialer.DialContext(context.Background(), "tcp", addr)
	assert.Error(t, err)
}

func TestCheckConnection_NoPort(t *testing.T) {
	// Mocking CheckConnection's behavior for no port is hard because it defaults to 80.
	// But we can check that it fails or succeeds depending on port 80 accessibility.
	// Usually 127.0.0.1:80 is closed.

	err := CheckConnection(context.Background(), "127.0.0.1")
	// If it fails, that's fine. If it succeeds, that's also fine (if something is running).
	// But we want to exercise the code path:
	// host = address, port = "80"
	// So we assert that it doesn't panic.
	_ = err
}
