// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// TestGlobalSettings_InvalidBindAddress tests that an invalid bind address in global settings is rejected.
// This serves as an "E2E-like" test for the validation logic integration.
func TestGlobalSettings_InvalidBindAddress(t *testing.T) {
	tests := []struct {
		name          string
		bindAddress   string
		expectedError string
	}{
		{
			name:          "Invalid Port -1",
			bindAddress:   "127.0.0.1:-1",
			expectedError: "invalid mcp_listen_address: port must be between 0 and 65535",
		},
		{
			name:          "Invalid Port 70000",
			bindAddress:   "127.0.0.1:70000",
			expectedError: "invalid mcp_listen_address: port must be between 0 and 65535",
		},
		{
			name:          "Valid Address",
			bindAddress:   "127.0.0.1:8080",
			expectedError: "",
		},
		{
			name:          "Remote Address Guard",
			bindAddress:   "0.0.0.0:8080",
			expectedError: "remote access guard: binding to 0.0.0.0:8080 requires MCPANY_ATTESTATION_TOKEN to be set",
		},
		{
			name:          "Remote Address Guard Port Only",
			bindAddress:   "8080",
			expectedError: "remote access guard: binding to 8080 requires MCPANY_ATTESTATION_TOKEN to be set",
		},
		{
			name:          "Remote Address Guard Empty Host",
			bindAddress:   ":8080",
			expectedError: "remote access guard: binding to :8080 requires MCPANY_ATTESTATION_TOKEN to be set",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := configv1.McpAnyServerConfig_builder{
				GlobalSettings: configv1.GlobalSettings_builder{
					McpListenAddress: proto.String(tc.bindAddress),
				}.Build(),
			}.Build()

			// We use Server binary type to trigger the bind address validation
			errs := Validate(context.Background(), cfg, Server)

			if tc.expectedError != "" {
				assert.NotEmpty(t, errs)
				found := false
				for _, err := range errs {
					if ae, ok := err.Err.(*ActionableError); ok {
						if ae.Err.Error() == tc.expectedError {
							found = true
							break
						}
					} else {
						if err.Err.Error() == tc.expectedError {
							found = true
							break
						}
					}
				}
				assert.True(t, found, "Expected error %q not found in %v", tc.expectedError, errs)
			} else {
				// We expect some errors because other fields might be missing (like audit config?),
				// but we shouldn't see the bind address error.
				for _, err := range errs {
					if err.Err.Error() == "invalid mcp_listen_address: port must be between 0 and 65535" {
						t.Errorf("Unexpected bind address error: %v", err)
					}
				}
			}
		})
	}
}
