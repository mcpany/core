package util

import (
	"net"
	"testing"
)

func TestIsPrivateIP_IPv4Compatible(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"::127.0.0.1", true},   // IPv4-compatible Loopback
		{"::169.254.1.1", true}, // IPv4-compatible Link-local
		{"::192.168.1.1", true}, // IPv4-compatible Private
		{"::10.0.0.1", true},    // IPv4-compatible Private
		{"::1.1.1.1", false},    // IPv4-compatible Public
		{"::1:192.168.1.1", false}, // Not compatible (has non-zero prefix)
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Fatalf("Failed to parse IP: %s", tc.ip)
		}
		if got := IsPrivateIP(ip); got != tc.expected {
			t.Errorf("IsPrivateIP(%s) = %v; want %v", tc.ip, got, tc.expected)
		}
	}
}

func TestIsPrivateNetworkIP_IPv4Compatible(t *testing.T) {
	tests := []struct {
		ip       string
		expected bool
	}{
		{"::192.168.1.1", true}, // IPv4-compatible Private
		{"::1.1.1.1", false},    // IPv4-compatible Public
	}

	for _, tc := range tests {
		ip := net.ParseIP(tc.ip)
		if ip == nil {
			t.Fatalf("Failed to parse IP: %s", tc.ip)
		}
		if got := IsPrivateNetworkIP(ip); got != tc.expected {
			t.Errorf("IsPrivateNetworkIP(%s) = %v; want %v", tc.ip, got, tc.expected)
		}
	}
}
