// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"testing"

	config "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestInjectProvenance(t *testing.T) {
	r := &ServiceRegistry{}

	tests := []struct {
		name           string
		serviceName    string
		expectVerified bool
		expectSigner   string
	}{
		{
			name:           "Official Service",
			serviceName:    "github",
			expectVerified: true,
			expectSigner:   "MCP Official",
		},
		{
			name:           "Verified Prefix",
			serviceName:    "mcp-weather",
			expectVerified: true,
			expectSigner:   "MCP Community Verified",
		},
		{
			name:           "Official Prefix",
			serviceName:    "official-weather",
			expectVerified: true,
			expectSigner:   "MCP Community Verified",
		},
		{
			name:           "Unverified Service",
			serviceName:    "random-service",
			expectVerified: false,
			expectSigner:   "Unverified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.UpstreamServiceConfig{}
			cfg.SetName(tt.serviceName)
			r.injectProvenance(cfg)

			assert.NotNil(t, cfg.GetProvenance())
			assert.Equal(t, tt.expectVerified, cfg.GetProvenance().GetVerified())
			if tt.expectVerified {
				assert.Equal(t, tt.expectSigner, cfg.GetProvenance().GetSignerIdentity())
				assert.NotNil(t, cfg.GetProvenance().GetAttestationTime())
				assert.Equal(t, "ecdsa-p256-sha256", cfg.GetProvenance().GetSignatureAlgorithm())
			} else {
				assert.Equal(t, "Unverified", cfg.GetProvenance().GetSignerIdentity())
			}
		})
	}
}
