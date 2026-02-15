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

func TestCheckConnection(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")

	// Start a local test server
	l, err := net.Listen("tcp", "127.0.0.2:0")
	require.NoError(t, err)
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	ts.Listener = l
	ts.Start()
	defer ts.Close()

	ctx := context.Background()

	// 1. Valid URL with scheme
	err = CheckConnection(ctx, ts.URL)
	assert.NoError(t, err)

	// 2. Valid host:port without scheme
	hostPort := ts.Listener.Addr().String()
	err = CheckConnection(ctx, hostPort)
	assert.NoError(t, err)

	// 3. Invalid URL parsing
	// url.Parse won't error easily on just path, but if we give garbage:
	err = CheckConnection(ctx, "http://%gh&%ij") // Invalid encoding
	// Actually url.Parse is very lenient.
	// Let's use control char.
	err = CheckConnection(ctx, "http://\x00")
	assert.Error(t, err)

	// 4. Unreachable
	ln, err := net.Listen("tcp", "127.0.0.2:0")
	require.NoError(t, err)
	closedAddr := ln.Addr().String()
	ln.Close()

	// Attempt to connect to closed port
	err = CheckConnection(ctx, closedAddr)
	assert.Error(t, err)

	// 5. Host without port (defaults to 80)
	// 127.0.0.1:80 is likely closed or filtered, causing error.
	// This exercises the path "net.SplitHostPort fails -> assume port 80".
	// Note: CheckConnection timeout is 5s, so this might slow down tests if it hangs.
	// But usually connection refused is fast.
	// Use a loopback address that definitely has nothing on port 80.
	err = CheckConnection(ctx, "127.0.0.1")
	// We expect error because nothing is listening on port 80
	assert.Error(t, err)

	// 6. URL without port (defaults to 80/443)
	// http://127.0.0.1 -> 127.0.0.1:80
	err = CheckConnection(ctx, "http://127.0.0.1")
	assert.Error(t, err)

    // https -> 443
    err = CheckConnection(ctx, "https://127.0.0.1")
    assert.Error(t, err)
}
