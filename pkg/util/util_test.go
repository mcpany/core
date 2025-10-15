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

func TestGenerateToolID(t *testing.T) {
	// Test cases for GenerateToolID
}

func TestGenerateServiceKey(t *testing.T) {
	// Test cases for GenerateServiceKey
}

func TestParseToolName(t *testing.T) {
	// Test cases for ParseToolName
}

func TestSanitizeOperationID(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no sanitization needed",
			input:    "valid-id_123.foo-bar~baz",
			expected: "valid-id_123.foo-bar~baz",
		},
		{
			name:     "single invalid character",
			input:    "invalid:id",
			expected: "invalid_05a79f_id",
		},
		{
			name:     "multiple different invalid sequences",
			input:    "a!b@c#d$e%f^g&h*i(j)k",
			expected: "a_0ab831_b_9a7821_c_d08f88_d_3cdf29_e_4345cb_f_5e6f80_g_7c4d33_h_df5824_i_28ed3a_j_e7064f_k",
		},
		{
			name:     "multiple same invalid characters",
			input:    "a:b:c",
			expected: "a_05a79f_b_05a79f_c",
		},
		{
			name:     "leading and trailing invalid characters",
			input:    "!abc!",
			expected: "_0ab831_abc_0ab831_",
		},
		{
			name:     "only invalid characters",
			input:    "!@#",
			expected: "_e2bb10_",
		},
		{
			name:     "mixed valid and invalid",
			input:    "foo!bar@baz",
			expected: "foo_0ab831_bar_9a7821_baz",
		},
		{
			name:     "long string with multiple invalid sequences",
			input:    "this_is_a_long_string_with_several_invalid_sequences-like*this-and&that",
			expected: "this_is_a_long_string_with_several_invalid_sequences-like_df5824_this-and_7c4d33_that",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sanitized := SanitizeOperationID(tc.input)
			assert.Equal(t, tc.expected, sanitized)
		})
	}
}

func TestGetDockerCommand(t *testing.T) {
	// Test cases for GetDockerCommand
}