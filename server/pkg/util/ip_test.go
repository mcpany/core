// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"context"
	"net"
	"testing"
)

func TestContextWithRemoteIP(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ip := "127.0.0.1"

	ctx = ContextWithRemoteIP(ctx, ip)

	retrievedIP, ok := RemoteIPFromContext(ctx)
	if !ok {
		t.Errorf("expected IP to be present in context")
	}
	if retrievedIP != ip {
		t.Errorf("expected IP %s, got %s", ip, retrievedIP)
	}

	// Test missing IP
	_, ok = RemoteIPFromContext(context.Background())
	if ok {
		t.Errorf("expected IP to be absent in empty context")
	}
}

func TestExtractIP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "IPv4 with port",
			input:    "192.168.1.1:8080",
			expected: "192.168.1.1",
		},
		{
			name:     "IPv4 without port",
			input:    "192.168.1.1",
			expected: "192.168.1.1",
		},
		{
			name:     "IPv6 with port",
			input:    "[2001:db8::1]:8080",
			expected: "2001:db8::1",
		},
		{
			name:     "IPv6 without port with brackets",
			input:    "[2001:db8::1]",
			expected: "2001:db8::1",
		},
		{
			name:     "IPv6 without port without brackets",
			input:    "2001:db8::1",
			expected: "2001:db8::1",
		},
		{
			name:     "Localhost with port",
			input:    "localhost:3000",
			expected: "localhost",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractIP(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractIP(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsPrivateNetworkIP_Unspecified(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"0.0.0.0", true},
		{"::", true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsPrivateNetworkIP(ip); got != tt.expected {
				t.Errorf("IsPrivateNetworkIP(%q) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}
