package util

import (
	"net"
	"testing"
)

func TestIsPrivateIP_NAT64_Fix(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// Private IPv4 in NAT64
		{"NAT64 192.168.1.1", "64:ff9b::192.168.1.1", true},
		{"NAT64 10.0.0.1", "64:ff9b::10.0.0.1", true},

		// Public IPv4 in NAT64
		{"NAT64 8.8.8.8", "64:ff9b::8.8.8.8", false},

		// Unspecified
		{"Unspecified IPv4", "0.0.0.0", true},
		{"Unspecified IPv6", "::", true},

		// Loopback
		{"Loopback IPv4", "127.0.0.1", true},
		{"Loopback IPv6", "::1", true},

		// Link-local
		{"Link-local IPv4", "169.254.1.1", true},
		{"Link-local IPv6", "fe80::1", true},

		// Private Network Blocks (random selection)
		{"Private 10.x", "10.1.2.3", true},
		{"Private 172.16.x", "172.16.1.1", true},
		{"Private 192.168.x", "192.168.0.1", true},

		// Public
		{"Public Google DNS", "8.8.8.8", false},
		{"Public Cloudflare DNS", "1.1.1.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			if got := IsPrivateIP(ip); got != tt.expected {
				t.Errorf("IsPrivateIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}
