/*
 * Copyright 2025 Author(s) of MCPXY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package util

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateToolID(t *testing.T) {
	tests := []struct {
		name        string
		serviceKey  string
		toolName    string
		expected    string
		expectError bool
	}{
		{"valid tool name", "", "toola", "toola", false},
		{"valid tool name with service key", "service1", "toola", "service1/-/toola", false},
		{"tool name already has service key", "service1", "service1/-/toola", "service1/-/toola", false},
		{"tool name already has service key with different separator", "service1", "service1/toola", "service1/toola", false},
		{"valid tool name with hyphen", "service1", "tool-a", "service1/-/tool-a", false},
		{"empty tool name", "service1", "", "", true},
		{"invalid tool name", "service1", "tool a", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateToolID(tt.serviceKey, tt.toolName)
			if (err != nil) != tt.expectError {
				t.Errorf("GenerateToolID() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.expected {
				t.Errorf("GenerateToolID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateServiceKey(t *testing.T) {
	tests := []struct {
		name        string
		serviceID   string
		expected    string
		expectError bool
	}{
		{"valid service ID", "service1", "service1", false},
		{"valid service ID with hyphen", "service-1", "service-1", false},
		{"empty service ID", "", "", true},
		{"invalid service ID", "service 1", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateServiceKey(tt.serviceID)
			if (err != nil) != tt.expectError {
				t.Errorf("GenerateServiceKey() error = %v, expectError %v", err, tt.expectError)
				return
			}
			if got != tt.expected {
				t.Errorf("GenerateServiceKey() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestGenerateUUID(t *testing.T) {
	generatedUUID := GenerateUUID()
	_, err := uuid.Parse(generatedUUID)
	require.NoError(t, err, "GenerateUUID() should return a valid UUID string")
}

func TestParseToolName(t *testing.T) {
	tests := []struct {
		name             string
		toolName         string
		expectedService  string
		expectedBareTool string
		expectError      bool
	}{
		{"valid tool name", "service1/-/toola", "service1", "toola", false},
		{"tool name without service", "toola", "", "toola", false},
		{"empty tool name", "", "", "", false},
		{"tool name with multiple separators", "service1/-/tool-a/-/part-b", "service1", "tool-a/-/part-b", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service, bareTool, err := ParseToolName(tt.toolName)
			if (err != nil) != tt.expectError {
				t.Errorf("ParseToolName() error = %v, expectError %v", err, tt.expectError)
				return
			}
			assert.Equal(t, tt.expectedService, service, "Service name should match expected")
			assert.Equal(t, tt.expectedBareTool, bareTool, "Bare tool name should match expected")
		})
	}
}

func TestSanitizeOperationID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"no changes", "valid-id", "valid-id"},
		{"with spaces", "invalid id", "invalid_b858cb_id"},
		{"with special chars", "id-with-$%^", "id-with-$_6dee27_"},
		{"multiple replacements", "a b c", "a_b858cb_b_b858cb_c"},
		{"already sanitized", "a_b858cb_b", "a_b858cb_b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sanitized := SanitizeOperationID(tt.input)
			assert.Equal(t, tt.expected, sanitized)
		})
	}
}
