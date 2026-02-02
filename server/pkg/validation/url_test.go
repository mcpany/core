// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
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
