package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidate_Security_ApiKey(t *testing.T) {
	tests := []struct {
		name      string
		apiKey    string
		wantError bool
	}{
		{
			name:      "weak api key",
			apiKey:    "weak",
			wantError: true,
		},
		{
			name:      "strong api key",
			apiKey:    "this-is-a-very-strong-api-key-123456",
			wantError: false,
		},
		{
			name:      "empty api key",
			apiKey:    "",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &configv1.McpAnyServerConfig{
				GlobalSettings: &configv1.GlobalSettings{
					ApiKey: proto.String(tt.apiKey),
				},
			}
			validationErrors := Validate(context.Background(), cfg, Server)
			if tt.wantError {
				assert.NotEmpty(t, validationErrors)
				if len(validationErrors) > 0 {
					assert.Contains(t, validationErrors[0].Error(), "API key must be at least 16 characters long")
				}
			} else {
				assert.Empty(t, validationErrors)
			}
		})
	}
}
