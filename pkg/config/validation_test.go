/*
 * Copyright 2024 Author(s) of MCP Any
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

package config

import (
	"testing"
)

func TestValidateListenAddress(t *testing.T) {
	testCases := []struct {
		name        string
		address     string
		expectError bool
	}{
		{
			name:        "valid address with port",
			address:     "localhost:8080",
			expectError: false,
		},
		{
			name:        "invalid address without port",
			address:     "localhost",
			expectError: true,
		},
		{
			name:        "invalid address with extra colon",
			address:     "localhost:8080:8080",
			expectError: true,
		},
		{
			name:        "empty address",
			address:     "",
			expectError: true,
		},
		{
			name:        "valid IPv6 address",
			address:     "[::1]:8080",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateListenAddress(tc.address)
			if (err != nil) != tc.expectError {
				t.Errorf("ValidateListenAddress(%q) = %v, want error: %v", tc.address, err, tc.expectError)
			}
		})
	}
}
