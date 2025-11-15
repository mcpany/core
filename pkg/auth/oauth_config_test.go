// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewOAuthConfig(t *testing.T) {
	testCases := []struct {
		name        string
		config      *OAuth2Config
		expectError bool
		expected    *OAuth2Config
	}{
		{
			name:        "Empty config",
			config:      &OAuth2Config{},
			expectError: false,
			expected:    &OAuth2Config{},
		},
		{
			name: "Valid config",
			config: &OAuth2Config{
				IssuerURL: "test-issuer",
				Audience:  "test-audience",
			},
			expectError: false,
			expected: &OAuth2Config{
				IssuerURL: "test-issuer",
				Audience:  "test-audience",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.config)
		})
	}
}
