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

func TestSanitizeID(t *testing.T) {
	testCases := []struct {
		name                     string
		ids                      []string
		alwaysAppendHash         bool
		maxSanitizedPrefixLength int
		hashLength               int
		expected                 string
		expectedError            string
	}{
		{
			name:     "single id, no hash",
			ids:      []string{"my-service"},
			expected: "my-service",
		},
		{
			name:             "single id, always hash",
			ids:              []string{"my-service"},
			alwaysAppendHash: true,
			hashLength:       8,
			expected:         "my-service_e43fea3e",
		},
		{
			name:       "long id, with hash",
			ids:        []string{"a-very-long-name-that-exceeds-the-maximum-length-of-53-characters"},
			hashLength: 8,
			expected:   "a-very-long-name-that-exceeds-the-maximum-length-of-_d7a922f9",
		},
		{
			name:       "id with special characters, with hash",
			ids:        []string{"my-service!@#$%^&*()"},
			hashLength: 8,
			expected:   "my-service_f699d6c2",
		},
		{
			name:          "empty id",
			ids:           []string{""},
			expectedError: "id cannot be empty",
		},
		{
			name:     "multiple ids",
			ids:      []string{"my-service", "my-tool"},
			expected: "my-service.my-tool",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			maxLen := maxSanitizedPrefixLength
			if tc.maxSanitizedPrefixLength > 0 {
				maxLen = tc.maxSanitizedPrefixLength
			}
			hashLen := hashLength
			if tc.hashLength > 0 {
				hashLen = tc.hashLength
			}
			actual, err := sanitizeID(tc.ids, tc.alwaysAppendHash, maxLen, hashLen)

			if tc.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestSanitizeServiceName(t *testing.T) {
	name := "my-service"
	expected := "my-service"
	actual, err := SanitizeServiceName(name)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestSanitizeToolName(t *testing.T) {
	serviceName := "my-service"
	toolName := "my-tool"
	expected := "my-service.my-tool"
	actual, err := SanitizeToolName(serviceName, toolName)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
