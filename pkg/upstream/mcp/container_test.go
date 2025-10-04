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

package mcp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetContainerImageForCommand(t *testing.T) {
	testCases := []struct {
		name     string
		command  string
		expected string
	}{
		{
			name:     "npx command",
			command:  "npx",
			expected: DefaultNodeImage,
		},
		{
			name:     "npm command",
			command:  "npm",
			expected: DefaultNodeImage,
		},
		{
			name:     "node command",
			command:  "node",
			expected: DefaultNodeImage,
		},
		{
			name:     "python command",
			command:  "python",
			expected: DefaultPythonImage,
		},
		{
			name:     "python3 command",
			command:  "python3",
			expected: DefaultPythonImage,
		},
		{
			name:     "unknown command",
			command:  "bash",
			expected: DefaultAlpineImage,
		},
		{
			name:     "empty command",
			command:  "",
			expected: DefaultAlpineImage,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := GetContainerImageForCommand(tc.command)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
