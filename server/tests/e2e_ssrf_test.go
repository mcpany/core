// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
)

// mockResolver implements util.IPResolver
type mockResolver struct {
	ip net.IP
}

func (m *mockResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	if m.ip == nil {
		return nil, &net.DNSError{Err: "no such host", Name: host}
	}
	return []net.IP{m.ip}, nil
}

func TestSafeDialer_BlocksDiscardOnly(t *testing.T) {
	// This E2E/Integration test verifies that the SafeDialer (which powers the HTTP client)
	// correctly blocks connections to Discard-Only IPv6 addresses (100::/64) which we added
	// to the validation logic.

	dialer := util.NewSafeDialer()
	// By default, SafeDialer blocks Private/LinkLocal/Loopback.

	// Mock the resolver to force resolution to a Discard-Only IP
	discardIP := net.ParseIP("100::1")
	dialer.Resolver = &mockResolver{ip: discardIP}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Try to dial
	_, err := dialer.DialContext(ctx, "tcp", "example.com:80")

	if err == nil {
		t.Fatal("Expected DialContext to fail, but it succeeded")
	}

	expectedErr := "ssrf attempt blocked"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error containing %q, got %q", expectedErr, err.Error())
	}
}

func TestSafeDialer_BlocksBenchmarkingIPv6(t *testing.T) {
	// Verify blocking of 2001:2::/48
	dialer := util.NewSafeDialer()
	benchIP := net.ParseIP("2001:2::1")
	dialer.Resolver = &mockResolver{ip: benchIP}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := dialer.DialContext(ctx, "tcp", "bench.example.com:80")

	if err == nil {
		t.Fatal("Expected DialContext to fail, but it succeeded")
	}

	expectedErr := "ssrf attempt blocked"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("Expected error containing %q, got %q", expectedErr, err.Error())
	}
}
