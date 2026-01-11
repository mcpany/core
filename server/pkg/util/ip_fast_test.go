// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
)

func TestIsPrivateNetworkIPv4_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		// Class A Private: 10.0.0.0/8
		{"10.0.0.0", "10.0.0.0", true},
		{"10.255.255.255", "10.255.255.255", true},

		// Class B Private: 172.16.0.0/12 (172.16.x.x - 172.31.x.x)
		{"172.16.0.0", "172.16.0.0", true},
		{"172.31.255.255", "172.31.255.255", true},
		{"172.15.255.255", "172.15.255.255", false}, // Lower bound check
		{"172.32.0.0", "172.32.0.0", false},         // Upper bound check

		// Class C Private: 192.168.0.0/16
		{"192.168.0.0", "192.168.0.0", true},
		{"192.168.255.255", "192.168.255.255", true},
		{"192.167.255.255", "192.167.255.255", false},
		{"192.169.0.0", "192.169.0.0", false},

		// CGNAT: 100.64.0.0/10 (100.64.x.x - 100.127.x.x)
		{"100.64.0.0", "100.64.0.0", true},
		{"100.127.255.255", "100.127.255.255", true},
		{"100.63.255.255", "100.63.255.255", false},
		{"100.128.0.0", "100.128.0.0", false},

		// Documentation: 192.0.2.0/24 (TEST-NET-1)
		{"192.0.2.0", "192.0.2.0", true},
		{"192.0.2.255", "192.0.2.255", true},
		{"192.0.3.0", "192.0.3.0", false},

		// Documentation: 198.51.100.0/24 (TEST-NET-2)
		{"198.51.100.0", "198.51.100.0", true},
		{"198.51.100.255", "198.51.100.255", true},
		{"198.51.101.0", "198.51.101.0", false},

		// Documentation: 203.0.113.0/24 (TEST-NET-3)
		{"203.0.113.0", "203.0.113.0", true},
		{"203.0.113.255", "203.0.113.255", true},
		{"203.0.114.0", "203.0.114.0", false},

		// Benchmarking: 198.18.0.0/15 (198.18.x.x - 198.19.x.x)
		{"198.18.0.0", "198.18.0.0", true},
		{"198.19.255.255", "198.19.255.255", true},
		{"198.17.255.255", "198.17.255.255", false},
		{"198.20.0.0", "198.20.0.0", false},

		// 0.0.0.0/8
		{"0.0.0.0", "0.0.0.0", true},
		{"0.255.255.255", "0.255.255.255", true},

		// Class E and Broadcast
		{"240.0.0.0", "240.0.0.0", true},
		{"255.255.255.255", "255.255.255.255", true},
		{"239.255.255.255", "239.255.255.255", false}, // Multicast is typically not "private" in this specific sense, though usually restricted.
		// Note: The original implementation iterated privateNetworkBlocks. Multicast (224.0.0.0/4) was NOT in that list.
		// So strict equality to old behavior is important. My new check:
		// if ip[0] >= 240. So Multicast (< 240) returns false. Correct.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("invalid ip: %s", tt.ip)
			}
			got := IsPrivateNetworkIP(ip)
			if got != tt.expected {
				t.Errorf("IsPrivateNetworkIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}
