// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"net"
	"testing"

	"google.golang.org/grpc/peer"
)

func TestIPAllowlistInterceptor_checkIP(t *testing.T) {
	tests := []struct {
		name        string
		allowedIPs  []string
		peerAddr    net.Addr
		expectError bool
	}{
		{
			name:       "Empty allowlist - Allow",
			allowedIPs: []string{},
			peerAddr:   &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234},
			expectError: false,
		},
		{
			name:       "Allowed IP",
			allowedIPs: []string{"127.0.0.1"},
			peerAddr:   &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5000},
			expectError: false,
		},
		{
			name:       "Denied IP",
			allowedIPs: []string{"127.0.0.1"},
			peerAddr:   &net.TCPAddr{IP: net.ParseIP("1.2.3.4"), Port: 1234},
			expectError: true,
		},
		{
			name:       "Allowed CIDR",
			allowedIPs: []string{"10.0.0.0/8"},
			peerAddr:   &net.TCPAddr{IP: net.ParseIP("10.1.2.3"), Port: 1234},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := NewIPAllowlistInterceptor(tt.allowedIPs)
			ctx := peer.NewContext(context.Background(), &peer.Peer{Addr: tt.peerAddr})
			err := interceptor.checkIP(ctx)
			if (err != nil) != tt.expectError {
				t.Errorf("checkIP() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
