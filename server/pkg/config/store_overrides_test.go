package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestApplySetOverrides(t *testing.T) {
	tests := []struct {
		name      string
		initial   map[string]interface{}
		setValues []string
		expected  map[string]interface{}
		protoMsg  proto.Message
	}{
		{
			name:      "Simple key value",
			initial:   map[string]interface{}{},
			setValues: []string{"global_settings.log_level=debug"},
			expected: map[string]interface{}{
				"global_settings": map[string]interface{}{
					"log_level": "debug",
				},
			},
			protoMsg: configv1.McpAnyServerConfig_builder{}.Build(),
		},
		{
			name:      "Nested key value",
			initial:   map[string]interface{}{},
			setValues: []string{"upstream_services[0].name=service1", "upstream_services[0].http_service.address=http://localhost:8080"},
			expected: map[string]interface{}{
				"upstream_services": map[string]interface{}{
					"0": map[string]interface{}{
						"name": "service1",
						"http_service": map[string]interface{}{
							"address": "http://localhost:8080",
						},
					},
				},
			},
			protoMsg: configv1.McpAnyServerConfig_builder{}.Build(),
		},
		{
			name: "Override existing value",
			initial: map[string]interface{}{
				"global_settings": map[string]interface{}{
					"log_level": "info",
				},
			},
			setValues: []string{"global_settings.log_level=debug"},
			expected: map[string]interface{}{
				"global_settings": map[string]interface{}{
					"log_level": "debug",
				},
			},
			protoMsg: configv1.McpAnyServerConfig_builder{}.Build(),
		},
		{
			name:      "Invalid set format",
			initial:   map[string]interface{}{},
			setValues: []string{"invalid_format"},
			expected:  map[string]interface{}{},
			protoMsg:  configv1.McpAnyServerConfig_builder{}.Build(),
		},
		{
			name: "Oneof clearing",
			initial: map[string]interface{}{
				"upstream_services": map[string]interface{}{
					"0": map[string]interface{}{
						"http_service": map[string]interface{}{
							"address": "http://localhost",
						},
					},
				},
			},
			// Setting grpc_service should clear http_service because they are in a oneof
			setValues: []string{"upstream_services[0].grpc_service.address=grpc://localhost"},
			expected: map[string]interface{}{
				"upstream_services": map[string]interface{}{
					"0": map[string]interface{}{
						"grpc_service": map[string]interface{}{
							"address": "grpc://localhost",
						},
					},
				},
			},
			protoMsg: configv1.McpAnyServerConfig_builder{}.Build(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			applySetOverrides(tc.initial, tc.setValues, tc.protoMsg)
			assert.Equal(t, tc.expected, tc.initial)
		})
	}
}
