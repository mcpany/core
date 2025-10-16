/*
 * Copyright 2025 Author(s) of MCP-XY
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

	"github.com/stretchr/testify/assert"
)

func TestGenerateServiceKey(t *testing.T) {
	testCases := []struct {
		name          string
		serviceID     string
		serviceType   string
		expected      string
		expectAnError bool
	}{
		{
			name:        "valid service ID without type",
			serviceID:   "my-service",
			serviceType: "",
			expected:    "my-service",
		},
		{
			name:        "valid service ID with type",
			serviceID:   "my-service",
			serviceType: "http",
			expected:    "http-my-service",
		},
		{
			name:          "empty service ID",
			serviceID:     "",
			serviceType:   "http",
			expectAnError: true,
		},
		{
			name:          "invalid service ID characters",
			serviceID:     "my service",
			serviceType:   "http",
			expectAnError: true,
		},
		{
			name:        "service ID with numbers",
			serviceID:   "my-service-123",
			serviceType: "grpc",
			expected:    "grpc-my-service-123",
		},
		{
			name:        "service ID with slashes",
			serviceID:   "my/service",
			serviceType: "http",
			expected:    "http-my/service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := GenerateServiceKey(tc.serviceID, tc.serviceType)
			if tc.expectAnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestGenerateToolID(t *testing.T) {
	testCases := []struct {
		name          string
		serviceKey    string
		toolName      string
		expected      string
		expectAnError bool
	}{
		{
			name:       "valid tool ID",
			serviceKey: "my-service",
			toolName:   "my-tool",
			expected:   "my-service/-/my-tool",
		},
		{
			name:       "empty service key",
			serviceKey: "",
			toolName:   "my-tool",
			expected:   "my-tool",
		},
		{
			name:          "empty tool name",
			serviceKey:    "my-service",
			toolName:      "",
			expectAnError: true,
		},
		{
			name:          "invalid tool name",
			serviceKey:    "my-service",
			toolName:      "my tool",
			expectAnError: true,
		},
		{
			name:       "fully qualified tool name",
			serviceKey: "my-service",
			toolName:   "another-service/-/my-tool",
			expected:   "another-service/-/my-tool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := GenerateToolID(tc.serviceKey, tc.toolName)
			if tc.expectAnError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}