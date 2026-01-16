// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeDialer_ZeroAddressBypass(t *testing.T) {
	// Start a listener on localhost
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer l.Close()

	port := l.Addr().(*net.TCPAddr).Port

	// Go routine to accept connection and close it
	go func() {
		conn, err := l.Accept()
		if err == nil {
			conn.Close()
		}
	}()

	// Configure SafeDialer: Allow Private, Block Loopback
	dialer := &util.SafeDialer{
		AllowLoopback:  false,
		AllowPrivate:   true,
		AllowLinkLocal: false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Try to dial 0.0.0.0:<port>
	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort("0.0.0.0", strconv.Itoa(port)))

	if err == nil {
		conn.Close()
		t.Errorf("Security Bypass: Successfully connected to localhost via 0.0.0.0 when AllowLoopback=false")
	} else {
        // We expect it to fail with "ssrf attempt blocked".
        assert.Contains(t, err.Error(), "ssrf attempt blocked")
	}
}
