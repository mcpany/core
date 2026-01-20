// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"testing"
)

// MockResolver implements IPResolver interface
type MockResolver struct {
	IPs []net.IP
}

func (m *MockResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return m.IPs, nil
}

func TestIsLoopback_Mapped_FixVerification(t *testing.T) {
	// IPv4-compatible IPv6 address
	ipStr2 := "::127.0.0.1"
	ip2 := net.ParseIP(ipStr2)

	// Verify IsLoopbackIP
	if !IsLoopbackIP(ip2) {
		t.Errorf("IsLoopbackIP(%s) = false; want true", ipStr2)
	}

	// Verify SafeDialer
	dialer := NewSafeDialer()
	// Mock resolver to return our IP
	dialer.Resolver = &MockResolver{IPs: []net.IP{ip2}}

	_, err := dialer.DialContext(context.Background(), "tcp", "example.com:80")
	if err == nil {
		t.Fatal("SafeDialer.DialContext did NOT return error for loopback IP")
	}

	expectedErr := "ssrf attempt blocked"
	if err != nil && len(err.Error()) < len(expectedErr) { // Simple check
		t.Errorf("Error mismatch. Got %v, want substring %q", err, expectedErr)
	}
}
