// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
)

func TestIsPrivateNetworkIPv4_FullCoverage(t *testing.T) {
	// These tests target every branch in isPrivateNetworkIPv4 in ip.go
	tests := []struct {
		ip       string
		expected bool
		desc     string
	}{
		// Case 0: 0.0.0.0/8
		{"0.0.0.0", true, "0.0.0.0"},
		{"0.255.255.255", true, "0.255.255.255"},

		// Case 10: 10.0.0.0/8
		{"10.0.0.0", true, "10.0.0.0"},
		{"10.255.255.255", true, "10.255.255.255"},

		// Case 100: 100.64.0.0/10 (64-127)
		{"100.63.255.255", false, "100.63.x.x (below 64)"},
		{"100.64.0.0", true, "100.64.0.0 (start of range)"},
		{"100.127.255.255", true, "100.127.255.255 (end of range)"},
		{"100.128.0.0", false, "100.128.0.0 (above 127)"},

		// Case 172: 172.16.0.0/12 (16-31)
		{"172.15.255.255", false, "172.15.x.x (below 16)"},
		{"172.16.0.0", true, "172.16.0.0 (start of range)"},
		{"172.31.255.255", true, "172.31.255.255 (end of range)"},
		{"172.32.0.0", false, "172.32.0.0 (above 31)"},

		// Case 192:
		// 192.168.0.0/16
		{"192.168.0.0", true, "192.168.0.0"},
		{"192.168.255.255", true, "192.168.255.255"},
		// 192.0.0.0/24 (ip[2] == 0)
		{"192.0.0.0", true, "192.0.0.0"},
		{"192.0.0.255", true, "192.0.0.255"},
		// 192.0.2.0/24 (ip[2] == 2)
		{"192.0.2.0", true, "192.0.2.0"},
		{"192.0.2.255", true, "192.0.2.255"},
		// Negative cases for 192
		{"192.0.1.0", false, "192.0.1.0 (not 0 or 2)"},
		{"192.0.3.0", false, "192.0.3.0 (not 0 or 2)"},
		{"192.1.0.0", false, "192.1.0.0 (not 0 or 168)"},

		// Case 198:
		// 198.18.0.0/15 (ip[1] == 18 or 19)
		{"198.18.0.0", true, "198.18.0.0"},
		{"198.19.255.255", true, "198.19.255.255"},
		{"198.17.255.255", false, "198.17.x.x (below 18)"},
		{"198.20.0.0", false, "198.20.0.0 (above 19)"},
		// 198.51.100.0/24 (ip[1] == 51 && ip[2] == 100)
		{"198.51.100.0", true, "198.51.100.0"},
		{"198.51.100.255", true, "198.51.100.255"},
		{"198.51.99.255", false, "198.51.99.x (ip[2] != 100)"},
		{"198.50.100.0", false, "198.50.100.0 (ip[1] != 51)"},

		// Case 203: 203.0.113.0/24
		{"203.0.113.0", true, "203.0.113.0"},
		{"203.0.113.255", true, "203.0.113.255"},
		{"203.0.112.255", false, "203.0.112.x"},
		{"203.1.113.0", false, "203.1.113.x"},

		// Class E (>= 240)
		{"240.0.0.0", true, "Class E start"},
		{"255.255.255.255", true, "Broadcast"},
		{"239.255.255.255", false, "Multicast (not private per function, except Class E)"},

		// Other public IPs
		{"8.8.8.8", false, "Google DNS"},
		{"1.1.1.1", false, "Cloudflare DNS"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			got := IsPrivateNetworkIP(ip)
			if got != tt.expected {
				t.Errorf("IsPrivateNetworkIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}
