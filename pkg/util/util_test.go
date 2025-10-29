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
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateID(t *testing.T) {
	testCases := []struct {
		name          string
		parts         []string
		expectedRegex string
		expectError   bool
		errorMessage  string
	}{
		{
			name:          "single valid part",
			parts:         []string{"valid-part"},
			expectedRegex: `^valid-part$`,
		},
		{
			name:          "multiple valid parts",
			parts:         []string{"serviceA", "toolB"},
			expectedRegex: `^serviceA\.toolB$`,
		},
		{
			name:          "single part with invalid characters",
			parts:         []string{"invalid part!"},
			expectedRegex: `^invalidpart_[a-f0-9]{8}$`,
		},
		{
			name:          "multiple parts with invalid characters",
			parts:         []string{"service A!", "tool B?"},
			expectedRegex: `^serviceA_[a-f0-9]{8}\.toolB_[a-f0-9]{8}$`,
		},
		{
			name:          "mixed valid and invalid parts",
			parts:         []string{"valid-service", "invalid tool!"},
			expectedRegex: `^valid-service\.invalidtool_[a-f0-9]{8}$`,
		},
		{
			name:          "part exceeding max length",
			parts:         []string{strings.Repeat("a", 70)},
			expectedRegex: `^aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_[a-f0-9]{8}$`,
		},
		{
			name:          "empty parts",
			parts:         []string{},
			expectError:   true,
			errorMessage:  "at least one part must be provided",
		},
		{
			name:          "empty string in parts",
			parts:         []string{"valid", ""},
			expectError:   true,
			errorMessage:  "name parts cannot be empty",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := GenerateID(tc.parts...)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.errorMessage, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Regexp(t, regexp.MustCompile(tc.expectedRegex), actual)
			}
		})
	}
}

func TestParseToolName(t *testing.T) {
	testCases := []struct {
		name               string
		toolName           string
		expectedService    string
		expectedBareTool   string
		expectError        bool
		expectedErrMessage string
	}{
		{
			name:             "Valid tool name",
			toolName:         "service1.tool1",
			expectedService:  "service1",
			expectedBareTool: "tool1",
			expectError:      false,
		},
		{
			name:             "Tool name with no service",
			toolName:         "tool1",
			expectedService:  "",
			expectedBareTool: "tool1",
			expectError:      false,
		},
		{
			name:             "Tool name with multiple separators",
			toolName:         "service1.tool1.extra",
			expectedService:  "service1",
			expectedBareTool: "tool1.extra",
			expectError:      false,
		},
		{
			name:             "Empty tool name",
			toolName:         "",
			expectedService:  "",
			expectedBareTool: "",
			expectError:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, bareTool, err := ParseToolName(tc.toolName)

			assert.Equal(t, tc.expectedService, service)
			assert.Equal(t, tc.expectedBareTool, bareTool)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErrMessage, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSanitizeOperationID(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No sanitization needed",
			input:    "valid-operation-id",
			expected: "valid-operation-id",
		},
		{
			name:     "Spaces replaced with hash",
			input:    "operation with spaces",
			expected: fmt.Sprintf("operation_%s_with_%s_spaces", hashString(" "), hashString(" ")),
		},
		{
			name:     "Multiple invalid characters replaced",
			input:    "op!@#$id",
			expected: fmt.Sprintf("op_%s_id", hashString("!@#$")),
		},
		{
			name:     "Starts and ends with invalid characters",
			input:    "$$op-id##",
			expected: fmt.Sprintf("_%s_op-id_%s_", hashString("$$"), hashString("##")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := SanitizeOperationID(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

// Helper function to hash a string for testing SanitizeOperationID
func hashString(s string) string {
	return SanitizeOperationID(s)[1:7]
}
