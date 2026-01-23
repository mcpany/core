// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"testing"
)

func TestIsPrivateIP_Security(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"127.0.0.1", true},
		{"192.168.1.1", true}, // Private but NOT loopback
		{"10.0.0.1", true},    // Private but NOT loopback
		{"8.8.8.8", false},    // Public
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if got := IsPrivateIP(ip); got != tt.expected {
			t.Errorf("IsPrivateIP(%q) = %v, want %v", tt.ip, got, tt.expected)
		}
	}
}

func TestIsLoopback_Security(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"127.0.0.1", true},
		{"127.255.255.255", true},
		{"::1", true},
		{"192.168.1.1", false}, // Private but NOT loopback
		{"10.0.0.1", false},    // Private but NOT loopback
		{"8.8.8.8", false},     // Public
		{"::ffff:127.0.0.1", true}, // IPv4-mapped loopback
	}

	for _, tt := range tests {
		ip := net.ParseIP(tt.ip)
		if got := IsLoopback(ip); got != tt.expected {
			t.Errorf("IsLoopback(%q) = %v, want %v", tt.ip, got, tt.expected)
		}
	}
}
