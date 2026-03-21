// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util //nolint:revive

import (
	"context"
	"net"
	"net/http"
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
			input:    "127.0.0.1:3000",
			expected: "127.0.0.1",
		},
		{
			name:     "IPv6 with zone index",
			input:    "fe80::1%eth0",
			expected: "fe80::1",
		},
		{
			name:     "Invalid IP string",
			input:    "not-an-ip",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "IP with extra spaces",
			input:    " 192.168.1.1 ", // ParseIP handles trimming? No, ParseIP returns nil for spaces.
			expected: "",            // net.ParseIP(" 192.168.1.1 ") returns nil
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

func TestGetClientIP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		xri        string
		trustProxy bool
		expected   string
	}{
		{
			name:       "No Proxy, No XFF",
			remoteAddr: "1.2.3.4:1234",
			xff:        "",
			trustProxy: false,
			expected:   "1.2.3.4",
		},
		{
			name:       "No Proxy, XFF Present (Ignored)",
			remoteAddr: "1.2.3.4:1234",
			xff:        "5.6.7.8",
			trustProxy: false,
			expected:   "1.2.3.4",
		},
		{
			name:       "Trust Proxy, No XFF",
			remoteAddr: "1.2.3.4:1234",
			xff:        "",
			trustProxy: true,
			expected:   "1.2.3.4",
		},
		{
			name:       "Trust Proxy, XFF Present",
			remoteAddr: "1.2.3.4:1234",
			xff:        "5.6.7.8",
			trustProxy: true,
			expected:   "5.6.7.8",
		},
		{
			name:       "Trust Proxy, Multi-XFF",
			remoteAddr: "1.2.3.4:1234",
			xff:        "5.6.7.8, 9.10.11.12",
			trustProxy: true,
			expected:   "5.6.7.8",
		},
		{
			name:       "Trust Proxy, XFF with Spaces",
			remoteAddr: "1.2.3.4:1234",
			xff:        " 5.6.7.8 , 9.10.11.12 ",
			trustProxy: true,
			expected:   "5.6.7.8",
		},
		{
			name:       "Trust Proxy, Invalid XFF, Fallback to RemoteAddr",
			remoteAddr: "1.2.3.4:1234",
			xff:        "invalid-ip",
			trustProxy: true,
			expected:   "1.2.3.4",
		},
		{
			name:       "Trust Proxy, X-Real-IP takes precedence",
			remoteAddr: "1.2.3.4:1234",
			xff:        "5.6.7.8",
			xri:        "9.9.9.9",
			trustProxy: true,
			expected:   "9.9.9.9",
		},
		{
			name:       "Trust Proxy, Invalid X-Real-IP, Fallback to XFF",
			remoteAddr: "1.2.3.4:1234",
			xff:        "5.6.7.8",
			xri:        "invalid-ip",
			trustProxy: true,
			expected:   "5.6.7.8",
		},
		{
			name:       "Trust Proxy, Invalid X-Real-IP and XFF, Fallback to RemoteAddr",
			remoteAddr: "1.2.3.4:1234",
			xff:        "invalid-xff",
			xri:        "invalid-xri",
			trustProxy: true,
			expected:   "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}
			if tt.xri != "" {
				req.Header.Set("X-Real-IP", tt.xri)
			}

			got := GetClientIP(req, tt.trustProxy)
			if got != tt.expected {
				t.Errorf("GetClientIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIsPrivateNetworkIP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip       string
		expected bool
	}{
		// Unspecified
		{"0.0.0.0", true},
		{"::", true},

		// IPv4 Private
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.0.1", true},
		{"192.168.255.255", true},

		// IPv4 CGNAT
		{"100.64.0.1", true},
		{"100.127.255.255", true},

		// IPv4 Loopback (handled by separate function usually, but IsPrivateIP calls IsPrivateNetworkIP for others)
		// IsPrivateNetworkIP logic for IPv4 includes:
		// 127.0.0.0/8 is NOT in isPrivateNetworkIPv4 list explicitly?
		// Wait, look at ip.go implementation of isPrivateNetworkIPv4.
		// It handles 0, 10, 100, 172, 192, 198, 203, >=240.
		// It does NOT handle 127.
		// So IsPrivateNetworkIP("127.0.0.1") might be false?
		// IsPrivateIP checks IsLoopback first.
		// Let's check ranges handled by isPrivateNetworkIPv4.
		{"127.0.0.1", false}, // Not a "Private Network" IP in RFC1918 sense, it's Loopback.

		// IPv4 Public
		{"8.8.8.8", false},
		{"1.1.1.1", false},

		// IPv6 Unique Local
		{"fc00::1", true},
		{"fd00::1", true},

		// IPv6 Documentation
		{"2001:db8::1", true},

		// IPv6 Public
		{"2001:4860:4860::8888", false},

		// NAT64 (64:ff9b::/96) + Private IPv4
		// 64:ff9b::192.168.1.1 -> 64:ff9b::c0a8:0101
		{"64:ff9b::c0a8:0101", true},
		// NAT64 + Public IPv4
		// 64:ff9b::8.8.8.8 -> 64:ff9b::0808:0808
		{"64:ff9b::0808:0808", false},

		// IPv4-Compatible (::a.b.c.d) + Private IPv4
		{"::192.168.1.1", true},
		// IPv4-Compatible + Public IPv4
		{"::8.8.8.8", false},
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

func TestIsPrivateIP(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ip       string
		expected bool
	}{
		// Loopback
		{"127.0.0.1", true},
		{"::1", true},

		// Link Local
		{"169.254.1.1", true},
		{"fe80::1", true},

		// Private Network
		{"192.168.1.1", true},
		{"10.0.0.1", true},

		// Public
		{"8.8.8.8", false},
		{"2607:f8b0:4005:805::200e", false},

		// IPv4-Compatible Loopback
		{"::127.0.0.1", true},
		// IPv4-Compatible Link-Local
		{"::169.254.1.1", true},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if got := IsPrivateIP(ip); got != tt.expected {
				t.Errorf("IsPrivateIP(%q) = %v, want %v", tt.ip, got, tt.expected)
			}
		})
	}
}
