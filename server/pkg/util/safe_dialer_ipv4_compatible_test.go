// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeDialer_IPv4Compatible_Bypass(t *testing.T) {
	d := NewSafeDialer()
	// Strict defaults: No Loopback, No Private, No LinkLocal

	tests := []struct {
		name    string
		addr    string
		wantErr string
	}{
		{
			name:    "IPv4-compatible Loopback",
			addr:    "[::127.0.0.1]:80",
			wantErr: "ssrf attempt blocked",
		},
		{
			name:    "IPv4-compatible LinkLocal",
			addr:    "[::169.254.169.254]:80",
			wantErr: "ssrf attempt blocked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := d.DialContext(context.Background(), "tcp", tt.addr)
			require.Error(t, err)
			// If we get a network error, it means we bypassed the SSRF check!
			// We want the error to be the SSRF blocking error.
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got %q", tt.wantErr, err.Error())
			}
		})
	}
}
