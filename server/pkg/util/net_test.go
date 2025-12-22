package util

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeDialContext(t *testing.T) {
	// Use a timeout to prevent hanging on unreachable IPs (like the public IP test case)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	tests := []struct {
		name        string
		addr        string
		wantErr     bool
		errContains string
	}{
		{
			name:        "Block Loopback IP",
			addr:        "127.0.0.1:8080",
			wantErr:     true,
			errContains: "ssrf attempt blocked",
		},
		{
			name:        "Block Private IP Class A",
			addr:        "10.0.0.1:8080",
			wantErr:     true,
			errContains: "ssrf attempt blocked",
		},
		{
			name:        "Block Private IP Class C",
			addr:        "192.168.1.1:8080",
			wantErr:     true,
			errContains: "ssrf attempt blocked",
		},
		{
			name:    "Block Localhost Hostname",
			addr:    "localhost:8080",
			wantErr: true,
			// "ssrf attempt blocked" or "no ip addresses found" depending on environment, usually blocked
			// But localhost usually resolves to 127.0.0.1 or ::1
			errContains: "ssrf attempt blocked",
		},
		{
			name:        "Block Unspecified IP (0.0.0.0)",
			addr:        "0.0.0.0:8080",
			wantErr:     true,
			errContains: "ssrf attempt blocked",
		},
		{
			name:        "Block Unspecified IP (IPv6)",
			addr:        "[::]:8080",
			wantErr:     true,
			errContains: "ssrf attempt blocked",
		},
		{
			name:        "Allow Public IP (Dial fails)",
			addr:        "8.8.8.8:12345", // Cloudflare DNS, likely closed port
			wantErr:     true,
			errContains: "connect", // Expect connection error, NOT ssrf blocked
		},
		{
			name:        "Invalid Address",
			addr:        "invalid-addr",
			wantErr:     true,
			errContains: "missing port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := SafeDialContext(ctx, "tcp", tt.addr)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					// Check if error contains expected string
					// For public IP, we accept standard dial errors
					if tt.name == "Allow Public IP (Dial fails)" {
						assert.False(t, strings.Contains(err.Error(), "ssrf attempt blocked"), "Should not be blocked by SSRF check")
					} else {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				}
			} else {
				require.NoError(t, err)
				if conn != nil {
					_ = conn.Close()
				}
			}
		})
	}
}
