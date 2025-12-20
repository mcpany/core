// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafeDialContext(t *testing.T) {
	ctx := context.Background()

	t.Run("DefaultBehavior", func(t *testing.T) {
		// Clear env vars to ensure defaults
		t.Setenv("MCPANY_ALLOW_LOOPBACK", "")
		t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK", "")

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
				name:        "Block LinkLocal IP",
				addr:        "169.254.169.254:80",
				wantErr:     true,
				errContains: "ssrf attempt blocked",
			},
			{
				name:        "Block Unspecified IP (0.0.0.0)",
				addr:        "0.0.0.0:8080",
				wantErr:     true,
				errContains: "ssrf attempt blocked",
			},
			{
				name:        "Allow Private IP Class A",
				addr:        "10.0.0.1:8080",
				wantErr:     true,
				errContains: "i/o timeout", // Timeout/unreachable implies it tried to connect
			},
			{
				name:        "Allow Public IP (Dial fails)",
				addr:        "8.8.8.8:12345",
				wantErr:     true,
				errContains: "i/o timeout",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// Use short timeout for allowed connections
				ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
				defer cancel()

				conn, err := SafeDialContext(ctx, "tcp", tt.addr)
				if tt.wantErr {
					require.Error(t, err)
					if tt.errContains == "i/o timeout" {
						// Accept timeout OR network unreachable OR connection refused
						// But definitely NOT ssrf blocked
						assert.NotContains(t, err.Error(), "ssrf attempt blocked")
					} else {
						assert.Contains(t, err.Error(), tt.errContains)
					}
				} else {
					require.NoError(t, err)
					if conn != nil {
						_ = conn.Close()
					}
				}
			})
		}
	})

	t.Run("StrictBlocking", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK", "false")

		// Private IPs should now be blocked
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		_, err := SafeDialContext(ctx, "tcp", "10.0.0.1:8080")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "ssrf attempt blocked")
	})

	t.Run("AllowLoopback", func(t *testing.T) {
		t.Setenv("MCPANY_ALLOW_LOOPBACK", "true")

		// Loopback should be allowed (will fail connection refused/timeout)
		ctx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
		defer cancel()
		_, err := SafeDialContext(ctx, "tcp", "127.0.0.1:8080")
		if err != nil {
			assert.NotContains(t, err.Error(), "ssrf attempt blocked")
		}
	})
}
