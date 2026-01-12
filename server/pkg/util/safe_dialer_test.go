// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeDialer(t *testing.T) {
	t.Parallel()
	t.Run("Default Strict", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		d := NewSafeDialer()
		// Test blocks loopback
		_, err := d.DialContext(ctx, "tcp", "127.0.0.1:8080")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("Allow Private", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		d := NewSafeDialer()
		d.AllowPrivate = true
		// 10.0.0.1 is private
		// Note: Dialing will fail with connection error (timeout), but should NOT be blocked by SSRF check.
		// We use a random port to ensure connection failure.
		_, err := d.DialContext(ctx, "tcp", "10.0.0.1:45678")
		require.Error(t, err)
		// Error should be a network error (e.g. timeout or unreachable), NOT "ssrf attempt blocked"
		assert.False(t, strings.Contains(err.Error(), "ssrf attempt blocked"), "Should not be blocked by SSRF check. Got: %v", err)
	})

	t.Run("Allow Loopback", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		d := NewSafeDialer()
		d.AllowLoopback = true
		// 127.0.0.1 is loopback
		// Start a listener to accept connection if we want success.
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		defer l.Close()
		addr := l.Addr().String()

		conn, err := d.DialContext(ctx, "tcp", addr)
		if err != nil {
			// If connection fails (e.g. environment issues), ensure it's NOT blocked by SSRF check.
			assert.False(t, strings.Contains(err.Error(), "ssrf attempt blocked"), "Should not be blocked by SSRF check. Got: %v", err)
		} else {
			conn.Close()
		}
	})

	t.Run("Block LinkLocal", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		d := NewSafeDialer()
		d.AllowPrivate = true  // Even if private allowed
		d.AllowLoopback = true // Even if loopback allowed
		// 169.254.169.254 is LinkLocal
		_, err := d.DialContext(ctx, "tcp", "169.254.169.254:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("Allow LinkLocal Explicitly", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		d := NewSafeDialer()
		d.AllowLinkLocal = true
		// 169.254.169.254 is LinkLocal
		// It will fail to connect (route unreachable or timeout), but shouldn't be blocked by SSRF.
		_, err := d.DialContext(ctx, "tcp", "169.254.169.254:45678")
		require.Error(t, err)
		assert.False(t, strings.Contains(err.Error(), "ssrf attempt blocked"), "Should not be blocked by SSRF check. Got: %v", err)
	})
}
