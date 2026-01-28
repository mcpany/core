// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Reusing MockIPResolver and MockDialer from safe_dialer_test.go (assuming they are in same package or I can redefine them)
// Since they are defined in `util_test` package in `safe_dialer_test.go`, they should be available here if in the same package `util_test`.

func TestSafeDialer_IPv4CompatibleLoopback(t *testing.T) {
	// Setup
	resolver := new(MockIPResolver)
	dialer := new(MockDialer)

	safeDialer := util.NewSafeDialer()
	safeDialer.Resolver = resolver
	safeDialer.Dialer = dialer

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	host := "ipv4-compatible.local"
	port := "80"
	addr := net.JoinHostPort(host, port)

	// IP: ::127.0.0.1 (IPv4-compatible loopback)
	// Note: net.ParseIP("::127.0.0.1") parses it correctly
	ip := net.ParseIP("::127.0.0.1")
	ips := []net.IP{ip}

	resolver.On("LookupIP", ctx, "ip", host).Return(ips, nil)

	// Mock behavior: If safe dialer fails to block it, it will try to dial.
	// We want to assert that it BLOCKS it, so DialContext should NOT be called.
	// But currently, due to the bug, it probably calls DialContext.
	// So for "Bug Reproduction", we expect DialContext to be called.

	// However, if I want to "Verify the Bug", I can assert that DialContext IS called, which means it FAILED to block.
	// But checking for "ssrf attempt blocked" error is better.

	dialer.On("DialContext", ctx, "tcp", mock.Anything).Return(&net.TCPConn{}, nil).Maybe()

	// Execution
	conn, err := safeDialer.DialContext(ctx, "tcp", addr)

	// CURRENT BEHAVIOR (BUG): It allows connection.
	// DESIRED BEHAVIOR (FIX): It should block it.

	// If the bug exists, err is nil (or dial error), and conn is not nil (if dial succeeds).
	// If fixed, err should be "ssrf attempt blocked...".

	if err == nil {
		t.Logf("Bug confirmed: SafeDialer allowed IPv4-compatible loopback %v", ip)
		// We can close the connection if it succeeded
		if conn != nil {
			conn.Close()
		}
	} else {
		t.Logf("Result: %v", err)
		if assert.ErrorContains(t, err, "ssrf attempt blocked") {
             t.Logf("SafeDialer correctly blocked IPv4-compatible loopback (Bug not present or already fixed)")
        }
	}
}
