// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeDialer_Security(t *testing.T) {
	// Setup common variables
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	host := "example.com"
	port := "80"
	addr := net.JoinHostPort(host, port)

	t.Run("BlockLoopback_Default", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		// IP resolves to loopback
		ip := net.ParseIP("127.0.0.1")
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{ip}, nil)

		// Exec
		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
		assert.Nil(t, conn)

		// Ensure DialContext was NOT called
		dialer.AssertNotCalled(t, "DialContext")
	})

	t.Run("BlockPrivate_Default", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		// IP resolves to private
		ip := net.ParseIP("192.168.1.1")
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{ip}, nil)

		// Exec
		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
		assert.Nil(t, conn)
		dialer.AssertNotCalled(t, "DialContext")
	})

	t.Run("BlockLinkLocal_Default", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		// IP resolves to link-local
		ip := net.ParseIP("169.254.1.1")
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{ip}, nil)

		// Exec
		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
		assert.Nil(t, conn)
		dialer.AssertNotCalled(t, "DialContext")
	})

	t.Run("AllowLoopback", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.AllowLoopback = true
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		// IP resolves to loopback
		ip := net.ParseIP("127.0.0.1")
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{ip}, nil)
		dialer.On("DialContext", ctx, "tcp", net.JoinHostPort(ip.String(), port)).Return(&net.TCPConn{}, nil)

		// Exec
		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, conn)
		dialer.AssertCalled(t, "DialContext", ctx, "tcp", net.JoinHostPort(ip.String(), port))
	})

	t.Run("AllowPrivate", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.AllowPrivate = true
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		// IP resolves to private
		ip := net.ParseIP("10.0.0.1")
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{ip}, nil)
		dialer.On("DialContext", ctx, "tcp", net.JoinHostPort(ip.String(), port)).Return(&net.TCPConn{}, nil)

		// Exec
		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		// Verify
		require.NoError(t, err)
		assert.NotNil(t, conn)
		dialer.AssertCalled(t, "DialContext", ctx, "tcp", net.JoinHostPort(ip.String(), port))
	})

	t.Run("MixedIPs_OneBlocked_AllBlocked", func(t *testing.T) {
		// If ANY IP is blocked, the request should be blocked (strict mode to prevent rebinding)
		// Wait, the implementation loops and checks ALL IPs.
		// If any ip fails check, it returns error immediately.

		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		ipPublic := net.ParseIP("1.2.3.4")
		ipPrivate := net.ParseIP("192.168.1.1")

		// Order matters? It iterates ips.
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{ipPublic, ipPrivate}, nil)

		// Exec
		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
		assert.Nil(t, conn)
		dialer.AssertNotCalled(t, "DialContext")
	})

	t.Run("DNSLookupFailure", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		expectedErr := errors.New("dns failure")
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP(nil), expectedErr)

		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "dns lookup failed")
		assert.Nil(t, conn)
	})

	t.Run("NoIPsFound", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{}, nil)

		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no ip addresses found")
		assert.Nil(t, conn)
	})

	t.Run("BlockNAT64LinkLocal_Default", func(t *testing.T) {
		resolver := new(MockIPResolver)
		dialer := new(MockDialer)
		safeDialer := util.NewSafeDialer()
		safeDialer.Resolver = resolver
		safeDialer.Dialer = dialer

		// IP resolves to NAT64 link-local (64:ff9b::169.254.1.1)
		// 169.254.1.1 -> a9.fe.01.01
		ip := net.ParseIP("64:ff9b::a9fe:0101")
		resolver.On("LookupIP", ctx, "ip", host).Return([]net.IP{ip}, nil)

		// Exec
		conn, err := safeDialer.DialContext(ctx, "tcp", addr)

		// Verify
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
		assert.Contains(t, err.Error(), "link-local ip")
		assert.Nil(t, conn)
		dialer.AssertNotCalled(t, "DialContext")
	})
}
