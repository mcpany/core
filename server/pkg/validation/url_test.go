// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"strings"
	"context"
	"net"
	"os"
	"testing"
)

func TestIsSafeURL(t *testing.T) {
	// Ensure the bypass env var is not set for this test
	originalEnv := os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")
	os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	// Restore original lookupIPFunc
	originalLookupIP := lookupIPFunc
	defer func() {
		lookupIPFunc = originalLookupIP
		if originalEnv != "" {
			os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", originalEnv)
		}
	}()

	// Mock DNS resolution
	lookupIPFunc = func(ctx context.Context, network, host string) ([]net.IP, error) {
		if host == "google.com" {
			return []net.IP{net.ParseIP("8.8.8.8")}, nil
		}
		if host == "localhost" {
			return []net.IP{net.ParseIP("127.0.0.1")}, nil
		}
		return net.DefaultResolver.LookupIP(ctx, network, host)
	}

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Safe Public URL",
			url:     "https://google.com",
			wantErr: false,
		},
		{
			name:    "Safe Public IP",
			url:     "http://8.8.8.8",
			wantErr: false,
		},
		{
			name:    "Localhost IP",
			url:     "http://127.0.0.1",
			wantErr: true,
		},
		{
			name:    "Localhost IPv6",
			url:     "http://[::1]",
			wantErr: true,
		},
		{
			name:    "Link-Local (Metadata)",
			url:     "http://169.254.169.254",
			wantErr: true,
		},
		{
			name:    "Unspecified",
			url:     "http://0.0.0.0",
			wantErr: true,
		},
		{
			name:    "Invalid Scheme",
			url:     "ftp://example.com",
			wantErr: true,
		},
		{
			name:    "Missing Host",
			url:     "http:///foo",
			wantErr: true,
		},
		// Note: We cannot easily test "127.0.0.1" domain resolution in a pure unit test
		// without mocking the resolver or assuming /etc/hosts,
		// but standard environment usually resolves 127.0.0.1 to 127.0.0.1.
		{
			name:    "Localhost Domain",
			url:     "http://127.0.0.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsSafeURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSafeURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateIP(t *testing.T) {
	tests := []struct {
		name          string
		ip            string
		allowLoopback bool
		allowPrivate  bool
		wantErr       bool
		errMsg        string
	}{
		{
			name:          "Safe Public IP",
			ip:            "8.8.8.8",
			allowLoopback: false,
			allowPrivate:  false,
			wantErr:       false,
		},
		{
			name:          "Loopback Not Allowed",
			ip:            "127.0.0.1",
			allowLoopback: false,
			allowPrivate:  false,
			wantErr:       true,
			errMsg:        "loopback address is not allowed",
		},
		{
			name:          "Loopback Allowed",
			ip:            "127.0.0.1",
			allowLoopback: true,
			allowPrivate:  false,
			wantErr:       false,
		},
		{
			name:          "IPv6 Loopback Not Allowed",
			ip:            "::1",
			allowLoopback: false,
			allowPrivate:  false,
			wantErr:       true,
			errMsg:        "loopback address is not allowed",
		},
		{
			name:          "IPv4-Compatible Loopback Not Allowed",
			ip:            "::ffff:127.0.0.1",
			allowLoopback: false,
			allowPrivate:  false,
			wantErr:       true,
			errMsg:        "loopback address is not allowed",
		},
		{
			name:          "Link-Local (Metadata) Not Allowed",
			ip:            "169.254.169.254",
			allowLoopback: true,
			allowPrivate:  true,
			wantErr:       true,
			errMsg:        "link-local address is not allowed",
		},
		{
			name:          "IPv6 Link-Local Not Allowed",
			ip:            "fe80::1",
			allowLoopback: true,
			allowPrivate:  true,
			wantErr:       true,
			errMsg:        "link-local address is not allowed",
		},
		{
			name:          "Link-Local Multicast Not Allowed",
			ip:            "224.0.0.1",
			allowLoopback: true,
			allowPrivate:  true,
			wantErr:       true,
			errMsg:        "link-local multicast address is not allowed",
		},
		{
			name:          "Multicast Not Allowed",
			ip:            "239.255.255.250",
			allowLoopback: true,
			allowPrivate:  true,
			wantErr:       true,
			errMsg:        "multicast address is not allowed",
		},
		{
			name:          "Unspecified Not Allowed",
			ip:            "0.0.0.0",
			allowLoopback: true,
			allowPrivate:  true,
			wantErr:       true,
			errMsg:        "unspecified address (0.0.0.0) is not allowed",
		},
		{
			name:          "Private Network Not Allowed",
			ip:            "10.0.0.1",
			allowLoopback: false,
			allowPrivate:  false,
			wantErr:       true,
			errMsg:        "private network address is not allowed",
		},
		{
			name:          "Private Network Allowed",
			ip:            "192.168.1.1",
			allowLoopback: false,
			allowPrivate:  true,
			wantErr:       false,
		},
		{
			name:          "NAT64 Loopback Not Allowed",
			ip:            "64:ff9b::127.0.0.1",
			allowLoopback: false,
			allowPrivate:  false,
			wantErr:       true,
			errMsg:        "loopback address is not allowed",
		},
		{
			name:          "NAT64 Link-Local Not Allowed",
			ip:            "64:ff9b::169.254.169.254",
			allowLoopback: false,
			allowPrivate:  false,
			wantErr:       true,
			errMsg:        "link-local address is not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}
			err := ValidateIP(ip, tt.allowLoopback, tt.allowPrivate)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateIP() expected error for IP %s, got none", tt.ip)
				} else if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateIP() error = %v, want to contain %v", err, tt.errMsg)
				}
			} else if err != nil {
				t.Errorf("ValidateIP() expected no error for IP %s, got %v", tt.ip, err)
			}
		})
	}
}
