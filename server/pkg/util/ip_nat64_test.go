// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
)

func TestIsPrivateNetworkIP_NAT64(t *testing.T) {
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "NAT64 Public IP (Google DNS)",
			ip:       "64:ff9b::8.8.8.8",
			expected: false, // Should be treated as public
		},
		{
			name:     "NAT64 Private IP (192.168.1.1)",
			ip:       "64:ff9b::192.168.1.1",
			expected: true, // Should be treated as private
		},
		{
			name:     "NAT64 Private IP (10.0.0.1)",
			ip:       "64:ff9b::10.0.0.1",
			expected: true, // Should be treated as private
		},
		{
			name:     "Regular Private IPv4",
			ip:       "192.168.1.1",
			expected: true,
		},
		{
			name:     "Regular Public IPv4",
			ip:       "8.8.8.8",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			if got := IsPrivateNetworkIP(ip); got != tt.expected {
				t.Errorf("IsPrivateNetworkIP(%s) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}

func TestIsPrivateIP_NAT64(t *testing.T) {
	// IsPrivateIP should also handle NAT64 correctly by delegating to IsPrivateNetworkIP
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{
			name:     "NAT64 Private IP (192.168.1.1)",
			ip:       "64:ff9b::192.168.1.1",
			expected: true,
		},
		{
			name:     "NAT64 Public IP (8.8.8.8)",
			ip:       "64:ff9b::8.8.8.8",
			expected: false,
		},
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
