package util

import (
	"net"
	"testing"
)

func TestIsPrivateNetworkIP_MissingRanges(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
		desc     string
	}{
		{"2001:2::1", true, "IPv6 Benchmarking (RFC 5180)"},
		{"64:ff9b:1::1", true, "Local-Use NAT64 (RFC 8215)"},
		// Control cases
		{"8.8.8.8", false, "Public IPv4"},
		{"2607:f8b0:4005:805::200e", false, "Public IPv6 (Google)"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
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
