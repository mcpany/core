// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func boolPtr(b bool) *bool                                                                { return &b }
func storageTypePtr(t configv1.AuditConfig_StorageType) *configv1.AuditConfig_StorageType { return &t }

func TestValidateUsers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		users        []*configv1.User
		expectErr    bool
		errSubstring string
	}{
		{
			name: "Valid User",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
					Authentication: &configv1.AuthenticationConfig{
						AuthMethod: &configv1.AuthenticationConfig_ApiKey{
							ApiKey: &configv1.APIKeyAuth{
								ParamName: strPtr("key"),
								KeyValue:  strPtr("secret"),
							},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Missing ID",
			users: []*configv1.User{
				{
					Id: strPtr(""), // Empty string pointer
				},
			},
			expectErr:    true,
			errSubstring: "user has empty id",
		},
		{
			name: "Duplicate ID",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
					Authentication: &configv1.AuthenticationConfig{
						AuthMethod: &configv1.AuthenticationConfig_ApiKey{
							ApiKey: &configv1.APIKeyAuth{ParamName: strPtr("k"), KeyValue: strPtr("v")},
						},
					},
				},
				{
					Id: strPtr("user1"), // Duplicate
					Authentication: &configv1.AuthenticationConfig{
						AuthMethod: &configv1.AuthenticationConfig_ApiKey{
							ApiKey: &configv1.APIKeyAuth{ParamName: strPtr("k"), KeyValue: strPtr("v")},
						},
					},
				},
			},
			expectErr:    true,
			errSubstring: "duplicate user id",
		},
		{
			name: "Missing Authentication",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid OAuth2",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
					Authentication: &configv1.AuthenticationConfig{
						AuthMethod: &configv1.AuthenticationConfig_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								TokenUrl: strPtr("invalid-url"),
							},
						},
					},
				},
			},
			expectErr:    true,
			errSubstring: "invalid oauth2 token_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configv1.McpAnyServerConfig{
				Users: tt.users,
			}
			errs := Validate(ctx, config, Server)
			if tt.expectErr {
				assert.NotEmpty(t, errs)
				found := false
				for _, e := range errs {
					if assert.Contains(t, e.Err.Error(), tt.errSubstring) {
						found = true
						break
					}
				}
				if !found && len(errs) > 0 {
					// Check if substring match failed but error existed
					// Actually strict check:
					assert.Fail(t, "expected error substring not found", "substring: %s, errors: %v", tt.errSubstring, errs)
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateGlobalSettings_Extended(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		gs           *configv1.GlobalSettings
		expectErr    bool
		errSubstring string
	}{
		{
			name: "Valid Audit File",
			gs: &configv1.GlobalSettings{
				Audit: &configv1.AuditConfig{
					Enabled:     boolPtr(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_FILE),
					OutputPath:  strPtr("/var/log/audit.log"),
				},
			},
			expectErr: false,
		},
		{
			name: "Audit File Missing Path",
			gs: &configv1.GlobalSettings{
				Audit: &configv1.AuditConfig{
					Enabled:     boolPtr(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_FILE),
				},
			},
			expectErr:    true,
			errSubstring: "output_path is required",
		},
		{
			name: "Audit Webhook Invalid URL",
			gs: &configv1.GlobalSettings{
				Audit: &configv1.AuditConfig{
					Enabled:     boolPtr(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_WEBHOOK),
					WebhookUrl:  strPtr("not-a-url"),
				},
			},
			expectErr:    true,
			errSubstring: "invalid webhook_url",
		},
		{
			name: "DLP Invalid Regex",
			gs: &configv1.GlobalSettings{
				Dlp: &configv1.DLPConfig{
					Enabled:        boolPtr(true),
					CustomPatterns: []string{"["}, // Invalid regex
				},
			},
			expectErr:    true,
			errSubstring: "invalid regex pattern",
		},
		{
			name: "GC Invalid Interval",
			gs: &configv1.GlobalSettings{
				GcSettings: &configv1.GCSettings{
					Enabled:  boolPtr(true),
					Interval: strPtr("not-a-duration"),
				},
			},
			expectErr:    true,
			errSubstring: "invalid interval",
		},
		{
			name: "GC Insecure Path",
			gs: &configv1.GlobalSettings{
				GcSettings: &configv1.GCSettings{
					Enabled: boolPtr(true),
					Paths:   []string{"../etc"},
				},
			},
			expectErr:    true,
			errSubstring: "not secure",
		},
		{
			name: "GC Relative Path (Not Allowed)",
			gs: &configv1.GlobalSettings{
				GcSettings: &configv1.GCSettings{
					Enabled: boolPtr(true),
					Paths:   []string{"relative/path"},
				},
			},
			expectErr:    true,
			errSubstring: "must be absolute",
		},
		{
			name: "Duplicate Profile Name",
			gs: &configv1.GlobalSettings{
				ProfileDefinitions: []*configv1.ProfileDefinition{
					{Name: strPtr("p1")},
					{Name: strPtr("p1")},
				},
			},
			expectErr:    true,
			errSubstring: "duplicate profile definition name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configv1.McpAnyServerConfig{
				GlobalSettings: tt.gs,
			}
			errs := Validate(ctx, config, Server)
			if tt.expectErr {
				assert.NotEmpty(t, errs)
				found := false
				for _, e := range errs {
					if len(e.Err.Error()) > 0 && (tt.errSubstring == "" || assert.Contains(t, e.Err.Error(), tt.errSubstring)) {
						found = true
						break
					}
				}
				if !found {
					t.Logf("Errors found: %v", errs)
					assert.Fail(t, "expected error not found")
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
