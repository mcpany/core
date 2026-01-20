// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
)

func TestIsPrivateNetworkIP_Fix(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{"::1", false},              // IPv6 Loopback
		{"::127.0.0.1", false},      // IPv4-compatible Loopback
		{"127.0.0.1", false},        // IPv4 Loopback
		{"10.0.0.1", true},          // Private Class A
		{"192.168.1.1", true},       // Private Class C
		{"8.8.8.8", false},          // Public
		{"100.64.0.1", true},        // CGNAT
		{"0.0.0.0", true},           // Unspecified
		{"::", true},                // Unspecified IPv6
		{"fc00::1", true},           // IPv6 ULA
		{"2001:db8::1", true},       // IPv6 Doc
		{"fe80::1", false},          // IPv6 Link-Local (IsPrivateNetworkIP excludes link-local)
		{"169.254.1.1", false},      // IPv4 Link-Local
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			if got := IsPrivateNetworkIP(ip); got != tt.want {
				t.Errorf("IsPrivateNetworkIP(%s) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}
