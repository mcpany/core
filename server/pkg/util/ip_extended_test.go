// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"net"
	"strings"
	"testing"
)

func TestIsPrivateNetworkIP_Extended(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"10.0.0.1", true},
		{"172.16.0.1", true},
		{"192.168.1.1", true},
		{"100.64.0.1", true}, // CGNAT
		{"fc00::1", true},    // Unique Local
		{"8.8.8.8", false},
		{"127.0.0.1", false}, // Loopback is not "Private Network" in our definition (handled separately)
		{"169.254.1.1", false}, // Link-local is not "Private Network" (handled separately)
		{"fe80::1", false},   // Link-local
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
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

func TestIsPrivateIP_Extended(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"100.64.0.1", true}, // CGNAT should now be true
		{"fe80::1", true},    // Link-local IPv6 should be true
		{"169.254.1.1", true},// Link-local IPv4
		{"127.0.0.1", true},  // Loopback
		{"::", true},         // Unspecified IPv6
		{"0.0.0.0", true},    // Unspecified IPv4
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			got := IsPrivateIP(ip)
			if got != tt.expected {
				t.Errorf("IsPrivateIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}

func TestSafeDialer_CGNAT(t *testing.T) {
	d := NewSafeDialer()
	// d.AllowPrivate is false by default.

	// 100.64.0.1 is in 100.64.0.0/10 (Shared Address Space / CGNAT)
	// It should be blocked if we consider it private.

	addr := "100.64.0.1:80"
	_, err := d.DialContext(context.Background(), "tcp", addr)

	if err == nil {
		t.Errorf("Expected SSRF block for CGNAT IP %s, but dial proceeded", addr)
	} else {
		if strings.Contains(err.Error(), "ssrf attempt blocked") {
			t.Logf("Blocked as expected: %v", err)
		} else {
			// If it is a network error, it means SafeDialer allowed it and tried to connect.
			t.Errorf("Expected SSRF block, but got network error: %v. This implies the IP was considered public.", err)
		}
	}
}
