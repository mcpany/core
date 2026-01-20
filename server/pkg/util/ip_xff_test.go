// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"net"
	"net/http"
	"testing"
)

func TestGetClientIP_XFF_Normalization(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		trustProxy bool
		want       string
	}{
		{
			name:       "XFF with brackets",
			remoteAddr: "1.2.3.4:1234",
			xff:        "[2001:db8::1]",
			trustProxy: true,
			want:       "2001:db8::1",
		},
		{
			name:       "XFF with brackets and port",
			remoteAddr: "1.2.3.4:1234",
			xff:        "[2001:db8::1]:8080",
			trustProxy: true,
			want:       "2001:db8::1",
		},
		{
			name:       "XFF standard IPv6",
			remoteAddr: "1.2.3.4:1234",
			xff:        "2001:db8::1",
			trustProxy: true,
			want:       "2001:db8::1",
		},
		{
			name:       "XFF IPv4",
			remoteAddr: "1.2.3.4:1234",
			xff:        "10.0.0.1",
			trustProxy: true,
			want:       "10.0.0.1",
		},
		{
			name:       "XFF multiple IPs",
			remoteAddr: "1.2.3.4:1234",
			xff:        "[2001:db8::1], 10.0.0.1",
			trustProxy: true,
			want:       "2001:db8::1",
		},
		{
			name:       "No trust proxy",
			remoteAddr: "[::1]:1234",
			xff:        "10.0.0.1",
			trustProxy: false,
			want:       "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				req.Header.Set("X-Forwarded-For", tt.xff)
			}

			got := GetClientIP(req, tt.trustProxy)
			if got != tt.want {
				t.Errorf("GetClientIP() = %v, want %v", got, tt.want)
			}

			// Verify it parses
			if net.ParseIP(got) == nil {
				t.Errorf("GetClientIP() returned unparseable IP: %v", got)
			}
		})
	}
}
