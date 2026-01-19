// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSafeHTTPClient(t *testing.T) {
	// Test default behavior (blocking private/loopback)
	client := util.NewSafeHTTPClient()
	require.NotNil(t, client)
	require.NotNil(t, client.Transport)

	// We can't easily test the blocking behavior without mocking network or having a real server,
	// but we can test if the environment variables affect the configuration.
	// Since NewSafeHTTPClient reads env vars directly, we need to set them.

	// Save original env vars
	origLoopback := os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
	origPrivate := os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
	defer func() {
		if origLoopback != "" {
			os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", origLoopback)
		} else {
			os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
		}
		if origPrivate != "" {
			os.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", origPrivate)
		} else {
			os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
		}
	}()

	t.Run("Default restricted", func(t *testing.T) {
		os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
		os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
		client := util.NewSafeHTTPClient()
		assert.Equal(t, 10*time.Second, client.Timeout)

		// Attempting to dial localhost should fail (we need to be careful not to actually rely on external network)
		// But here we are just unit testing the creation logic.
		// To verify the dialer logic, we'd need to invoke the DialContext.
		// Let's try to invoke DialContext with a loopback address.

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// 127.0.0.1 is loopback
		_, err := transport.DialContext(ctx, "tcp", "127.0.0.1:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")

		// 0.0.0.0 is unspecified/loopback-ish
		_, err = transport.DialContext(ctx, "tcp", "0.0.0.0:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")

		// [::] is unspecified/loopback-ish
		_, err = transport.DialContext(ctx, "tcp", "[::]:80")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("Allow loopback", func(t *testing.T) {
		os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
		os.Unsetenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES")
		client := util.NewSafeHTTPClient()

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		// It should theoretically allow it, but since there is likely no server listening, it might fail with connection refused or timeout.
		// But it shouldn't be the SSRF error.
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := transport.DialContext(ctx, "tcp", "127.0.0.1:12345")
		// It will likely fail, but check the error message.
		if err != nil {
			assert.NotContains(t, err.Error(), "ssrf attempt blocked", "Should not block loopback when allowed")
		}
	})

	t.Run("Allow private", func(t *testing.T) {
		os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
		os.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")
		client := util.NewSafeHTTPClient()

		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok)

		// 192.168.1.1 is private
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		_, err := transport.DialContext(ctx, "tcp", "192.168.1.1:12345")
		if err != nil {
			assert.NotContains(t, err.Error(), "ssrf attempt blocked", "Should not block private when allowed")
		}
	})
}

func TestSafeDialContext(t *testing.T) {
	// SafeDialContext is just a wrapper around NewSafeDialer().DialContext
	// We verify that it blocks loopback by default.

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// 127.0.0.1 is loopback
	_, err := util.SafeDialContext(ctx, "tcp", "127.0.0.1:80")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ssrf attempt blocked")
}

func TestCheckConnection(t *testing.T) {
	// Start a test server
	server := http.Server{Handler: http.NotFoundHandler()}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	go func() {
		_ = server.Serve(listener)
	}()
	defer func() { _ = server.Close() }()

	addr := listener.Addr().String()

	t.Run("Success host:port", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := util.CheckConnection(ctx, addr)
		assert.NoError(t, err)
	})

	t.Run("Success http://host:port", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := util.CheckConnection(ctx, "http://"+addr)
		assert.NoError(t, err)
	})

	t.Run("Success https://host:port", func(t *testing.T) {
		// It just checks TCP connection, so it should succeed even if we speak HTTP
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := util.CheckConnection(ctx, "https://"+addr)
		assert.NoError(t, err)
	})

	t.Run("Failure connection refused", func(t *testing.T) {
		// Pick a port that is likely closed
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := util.CheckConnection(ctx, "127.0.0.1:54321")
		assert.Error(t, err)
	})

	t.Run("Invalid URL", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		err := util.CheckConnection(ctx, "http://invalid-url\x7f")
		assert.Error(t, err)
	})

	t.Run("Host only defaults to port 80", func(t *testing.T) {
		// This will likely fail connection, but we check if it tried port 80
		// We can't easily check the port without mocking net.Dialer, but we can check error message
		// or just ensure it doesn't error on parsing.
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		err := util.CheckConnection(ctx, "127.0.0.1") // no port
		// Should try 127.0.0.1:80. Likely connection refused or timeout.
		assert.Error(t, err)
		// Error message might contain the address
		assert.Contains(t, err.Error(), ":80")
	})

	t.Run("Scheme defaults", func(t *testing.T) {
		// http -> 80
		// https -> 443
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		err := util.CheckConnection(ctx, "http://127.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ":80")

		err = util.CheckConnection(ctx, "https://127.0.0.1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), ":443")
	})
}
