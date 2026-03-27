// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateConfigAgainstSchema(t *testing.T) {
	// 1. Valid Configuration
	validConfig := map[string]interface{}{
		"global_settings": map[string]interface{}{
			"log_level":          "LOG_LEVEL_INFO", // Enum as string
			"mcp_listen_address": "0.0.0.0:8080",
		},
		"upstream_services": []interface{}{
			map[string]interface{}{
				"name": "service1",
				"http_service": map[string]interface{}{
					"address": "http://example.com",
				},
			},
		},
		"collections": []interface{}{},
		"users":                        []interface{}{},
	}

	err := ValidateConfigAgainstSchema(validConfig)
	assert.NoError(t, err)

	// 2. Invalid Configuration - Wrong Type
	invalidConfigType := map[string]interface{}{
		"global_settings": map[string]interface{}{
			"mcp_listen_address": 12345, // Should be string
		},
	}
	err = ValidateConfigAgainstSchema(invalidConfigType)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema validation failed")

	// 3. Invalid Configuration - Unknown Field (if strict)
	// Our schema generator sets additionalProperties: false
	invalidConfigUnknown := map[string]interface{}{
		"unknown_field": "value",
	}
	err = ValidateConfigAgainstSchema(invalidConfigUnknown)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "additionalProperties")
}

func TestGenerateJSONSchemaBytes(t *testing.T) {
	bytes, err := GenerateJSONSchemaBytes()
	assert.NoError(t, err)
	assert.NotNil(t, bytes)
	assert.Contains(t, string(bytes), "$schema")
}
