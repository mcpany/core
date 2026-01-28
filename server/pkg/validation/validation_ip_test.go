// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLoopback(t *testing.T) {
	tests := []struct {
		name     string
		ipStr    string
		expected bool
	}{
		{"IPv4 Loopback", "127.0.0.1", true},
		{"IPv4 Loopback Range", "127.10.20.30", true},
		{"IPv6 Loopback", "::1", true},
		{"IPv4-Mapped Loopback", "::ffff:127.0.0.1", true},
		{"IPv4-Compatible Loopback", "::127.0.0.1", true},
		{"NAT64 Loopback", "64:ff9b::127.0.0.1", true},
		{"IPv4 Public", "8.8.8.8", false},
		{"IPv4 Private", "192.168.1.1", false},
		{"IPv6 Public", "2001:4860:4860::8888", false},
		{"IPv6 LinkLocal", "fe80::1", false},
		{"Unspecified", "0.0.0.0", false},
		{"IPv6 Unspecified", "::", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ipStr)
			assert.NotNil(t, ip)
			assert.Equal(t, tc.expected, IsLoopback(ip), "IP: %s", tc.ipStr)
		})
	}
}

func TestIsLinkLocal(t *testing.T) {
	tests := []struct {
		name     string
		ipStr    string
		expected bool
	}{
		{"IPv4 LinkLocal", "169.254.1.1", true},
		{"IPv6 LinkLocal", "fe80::1", true},
		{"IPv4-Mapped LinkLocal", "::ffff:169.254.1.1", true},
		{"IPv4-Compatible LinkLocal", "::169.254.1.1", true},
		{"NAT64 LinkLocal", "64:ff9b::169.254.1.1", true},
		{"IPv4 Loopback", "127.0.0.1", false},
		{"IPv4 Public", "8.8.8.8", false},
		{"IPv4 Private", "10.0.0.1", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ipStr)
			assert.NotNil(t, ip)
			assert.Equal(t, tc.expected, IsLinkLocal(ip), "IP: %s", tc.ipStr)
		})
	}
}

func TestIsPrivateNetworkIPv4_Branches(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
		reason   string
	}{
		// 100.64.0.0/10 (CGNAT) -> 100.64.0.0 - 100.127.255.255
		{"100.63.255.255", false, "Below CGNAT"},
		{"100.64.0.0", true, "Start CGNAT"},
		{"100.127.255.255", true, "End CGNAT"},
		{"100.128.0.0", false, "Above CGNAT"},

		// 172.16.0.0/12 -> 172.16.0.0 - 172.31.255.255
		{"172.15.255.255", false, "Below 172.16"},
		{"172.16.0.0", true, "Start 172.16"},
		{"172.31.255.255", true, "End 172.16"},
		{"172.32.0.0", false, "Above 172.16"},

		// 192.168.0.0/16
		{"192.168.0.0", true, "192.168.x.x"},
		{"192.169.0.0", false, "Not 192.168"},

		// 192.0.0.0/24 and 192.0.2.0/24
		{"192.0.0.1", true, "IETF Protocol"},
		{"192.0.2.1", true, "TEST-NET-1"},
		{"192.0.1.1", false, "Between IETF and TEST-NET-1"},
		{"192.0.3.1", false, "Above TEST-NET-1"},
		{"192.1.0.0", false, "192.1.x.x"},

		// 198.18.0.0/15 -> 198.18.0.0 - 198.19.255.255
		{"198.17.255.255", false, "Below Benchmarking"},
		{"198.18.0.0", true, "Start Benchmarking"},
		{"198.19.255.255", true, "End Benchmarking"},
		{"198.20.0.0", false, "Above Benchmarking"},

		// 198.51.100.0/24 (TEST-NET-2)
		{"198.51.100.0", true, "TEST-NET-2"},
		{"198.51.101.0", false, "Not TEST-NET-2"},
		{"198.50.100.0", false, "Not 198.51"},

		// 203.0.113.0/24 (TEST-NET-3)
		{"203.0.113.0", true, "TEST-NET-3"},
		{"203.0.114.0", false, "Not TEST-NET-3"},
		{"203.1.113.0", false, "Not 203.0"},
	}

	for _, tc := range tests {
		t.Run(tc.ip, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			assert.NotNil(t, ip)
			ip4 := ip.To4()
			assert.NotNil(t, ip4)
			assert.Equal(t, tc.expected, IsPrivateNetworkIPv4(ip4), tc.reason)
		})
	}
}
