// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"os"
	"testing"
)

func TestIsSafeURL_PrivateIPs(t *testing.T) {
	os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "Private IP 10.x",
			url:     "http://10.0.0.1",
			wantErr: true,
		},
		{
			name:    "Private IP 192.168.x",
			url:     "http://192.168.1.1",
			wantErr: true,
		},
		{
			name:    "Private IP 172.16.x",
			url:     "http://172.16.0.1",
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
