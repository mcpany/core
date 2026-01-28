// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	"net"
	"testing"

	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
)

func TestIsIPv4CompatibleLoopback(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"::127.0.0.1", true},
		{"::7f00:0001", true}, // Same as ::127.0.0.1
		{"::1", false},        // Pure IPv6 Loopback
		{"::ffff:127.0.0.1", false}, // IPv4-mapped (not compatible)
		{"::127.0.0.2", true},
		{"::126.0.0.1", false},
		{"::10.0.0.1", false}, // Compatible but not loopback
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.ip)
		assert.Equal(t, tc.expected, validation.IsIPv4CompatibleLoopback(ip), "IP: %s", tc.ip)
	}
}

func TestIsPrivateNetworkIP_IPv6(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"fc00::1", true},      // Unique Local
		{"fd00::1", true},      // Unique Local
		{"2001:db8::1", true},  // Documentation
		{"fe80::1", false},     // Link Local (excluded)
		{"::1", false},         // Loopback (excluded)
		{"::", true},           // Unspecified (included as private/unsafe)
		{"2001:4860:4860::8888", false}, // Google DNS (Public)
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.ip)
		assert.Equal(t, tc.expected, validation.IsPrivateNetworkIP(ip), "IP: %s", tc.ip)
	}
}

func TestIsPrivateNetworkIPv4_Coverage(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"0.0.0.0", true},
		{"10.0.0.1", true},
		{"100.64.0.1", true}, // Shared Address Space
		{"172.16.0.1", true},
		{"192.168.0.1", true},
		{"192.0.0.1", true},
		{"192.0.2.1", true},
		{"198.18.0.1", true},
		{"198.51.100.1", true},
		{"203.0.113.1", true},
		{"240.0.0.1", true}, // Class E
		{"8.8.8.8", false},
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.ip).To4()
		assert.Equal(t, tc.expected, validation.IsPrivateNetworkIPv4(ip), "IP: %s", tc.ip)
	}
}

func TestIsPrivateIP_Coverage(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"::1", true},        // Loopback
		{"::", true},         // Unspecified
		{"127.0.0.1", true},  // IPv4 Loopback
		{"169.254.1.1", true},// IPv4 Link-local
		{"fe80::1", true},    // IPv6 Link-local
		{"::127.0.0.1", true}, // IPv4-compatible Loopback
		{"::169.254.1.1", true}, // IPv4-compatible Link-local
		{"64:ff9b::7f00:0001", true}, // NAT64 Loopback (127.0.0.1)
		{"64:ff9b::a9fe:0101", true}, // NAT64 Link-local (169.254.1.1)
		{"10.0.0.1", true},   // Private Network
		{"8.8.8.8", false},   // Public
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.ip)
		assert.Equal(t, tc.expected, validation.IsPrivateIP(ip), "IP: %s", tc.ip)
	}
}
