// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"net"
	"testing"
)

func TestIsPrivateNetworkIP_Extra(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip       string
		expected bool
	}{
		// Existing checks (sanity)
		{"127.0.0.1", false}, // Loopback is not "PrivateNetwork" (RFC 1918)
		{"10.0.0.1", true},   // Private
		{"8.8.8.8", false},   // Public

		// New checks
		{"100::1", true},                    // Discard-Only (RFC 6666)
		{"100::ffff:ffff:ffff:ffff", true},  // Discard-Only Max
		{"2001:2::1", true},                 // Benchmarking (RFC 5180)
		{"2001:2:0:ffff:ffff:ffff:ffff:ffff", true}, // Benchmarking Max

		// Edge cases around new ranges
		{"100:0:0:1::1", false}, // Not 100::/64
		{"2001:3::1", false},    // Not 2001:2::/48

		// Global Unicast (Verification)
		{"2001:4860:4860::8888", false}, // Google DNS IPv6
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Invalid IP in test case: %s", tt.ip)
			}
			if got := IsPrivateNetworkIP(ip); got != tt.expected {
				t.Errorf("IsPrivateNetworkIP(%q) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}
