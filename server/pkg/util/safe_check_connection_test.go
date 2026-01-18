package util_test

import (
	"context"
	"net"
	"os"
	"testing"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCheckConnection_Safe(t *testing.T) {
	// 1. Start a local listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().String()

	// 2. Verify SafeDialContext BLOCKS it (default behavior)
	conn, err := util.SafeDialContext(context.Background(), "tcp", addr)
	if err == nil {
		conn.Close()
		t.Fatal("SafeDialContext should have blocked loopback address")
	}
	assert.Contains(t, err.Error(), "ssrf attempt blocked")

	// 3. Verify CheckConnection BLOCKS it (Fix verified)
	err = util.CheckConnection(context.Background(), addr)
	assert.Error(t, err, "CheckConnection should have failed")
	assert.Contains(t, err.Error(), "ssrf attempt blocked")
}

func TestCheckConnection_WithEnvVar(t *testing.T) {
	// 1. Start a local listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().String()

	// 2. Set Env Var to allow loopback
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// 3. Verify CheckConnection ALLOWS it
	err = util.CheckConnection(context.Background(), addr)
	assert.NoError(t, err, "CheckConnection should have succeeded with env var set")
}
