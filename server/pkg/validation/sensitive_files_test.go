// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSensitivePath_Hardened(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		// Existing checks
		{"env file", ".env", true},
		{"git dir", ".git", true},
		{"config yaml", "config.yaml", true},
		{"private key", "id_rsa", true},
		{"pem file", "cert.pem", true},

		// New checks
		{"ssh dir", ".ssh", true},
		{"ssh dir path", "/home/user/.ssh", true},
		{"aws dir", ".aws", true},
		{"kube dir", ".kube", true},

		{"bashrc", ".bashrc", true},
		{"bash_profile", ".bash_profile", true},
		{"zshrc", ".zshrc", true},
		{"profile", ".profile", true},

		// Allowed
		{"text file", "hello.txt", false},
		{"hidden file", ".hidden", false},
		{"ssh pub key", "id_rsa.pub", false}, // Technically allowed by current logic, though base logic might flag id_rsa prefix if I'm not careful? No, strict equality for id_rsa
		{"config txt", "config.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsSensitivePath(tt.path)
			if tt.wantErr {
				assert.Error(t, err, "Expected error for path %q", tt.path)
			} else {
				assert.NoError(t, err, "Expected no error for path %q", tt.path)
			}
		})
	}
}
