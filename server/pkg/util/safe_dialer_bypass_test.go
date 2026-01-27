// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeDialer_Bypass(t *testing.T) {
	// Create a SafeDialer with strict defaults (No Loopback, No Private)
	// Defaults: AllowLoopback=false, AllowPrivate=false, AllowLinkLocal=false

	t.Run("Block IPv4-compatible IPv6 Loopback", func(t *testing.T) {
		client := util.NewSafeHTTPClient()
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// [::127.0.0.1] is IPv4-compatible IPv6 address for 127.0.0.1
		// It should be blocked as loopback.
		_, err := transport.DialContext(ctx, "tcp", "[::127.0.0.1]:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("Block 0.0.0.0 as Loopback when Private is Allowed", func(t *testing.T) {
		// Allow Private, but Block Loopback
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
		t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
		client := util.NewSafeHTTPClient()
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// 0.0.0.0 refers to "this host", essentially loopback.
		// If 0.0.0.0 is treated as "Private" but not "Loopback", and Private is allowed, this might pass.
		// It SHOULD be blocked because it bypasses loopback restriction.
		_, err := transport.DialContext(ctx, "tcp", "0.0.0.0:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("Block IPv6 Unspecified as Loopback", func(t *testing.T) {
		// Allow Private, Block Loopback
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
		t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
		client := util.NewSafeHTTPClient()
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// [::] is unspecified, effectively loopback
		_, err := transport.DialContext(ctx, "tcp", "[::]:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("Block IPv4-mapped 0.0.0.0 as Loopback", func(t *testing.T) {
		// Allow Private, Block Loopback
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "false")
		t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
		client := util.NewSafeHTTPClient()
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// [::ffff:0.0.0.0]
		_, err := transport.DialContext(ctx, "tcp", "[::ffff:0.0.0.0]:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})
}
