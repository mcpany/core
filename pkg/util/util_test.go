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
	"os"
	"strings"
	"testing"
)

func TestGenerateID(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid id",
			input:       "test_service",
			expected:    "test_service_96b74ebc",
			expectError: false,
		},
		{
			name:        "id with special characters",
			input:       "test-service-1.0",
			expected:    "test-service-10_57f3fff2",
			expectError: false,
		},
		{
			name:        "long id",
			input:       strings.Repeat("a", 100),
			expected:    fmt.Sprintf("%s_28165978", strings.Repeat("a", 53)),
			expectError: false,
		},
		{
			name:        "empty id",
			input:       "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := GenerateID(tc.input)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if actual != tc.expected {
					t.Errorf("Expected %q, but got %q", tc.expected, actual)
				}
				if len(actual) > maxGeneratedIDLength {
					t.Errorf("Generated ID exceeds max length of %d", maxGeneratedIDLength)
				}
			}
		})
	}
}

func TestGenerateServiceKey(t *testing.T) {
	input := "test_service"
	expected, _ := GenerateID(input)
	actual, err := GenerateServiceKey(input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if actual != expected {
		t.Errorf("Expected %q, but got %q", expected, actual)
	}
}

func TestGetDockerCommand(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		os.Unsetenv("USE_SUDO_FOR_DOCKER")
		cmd, args := GetDockerCommand()
		if cmd != "docker" {
			t.Errorf("Expected cmd to be 'docker', but got %q", cmd)
		}
		if len(args) != 0 {
			t.Errorf("Expected args to be empty, but got %v", args)
		}
	})

	t.Run("with sudo", func(t *testing.T) {
		os.Setenv("USE_SUDO_FOR_DOCKER", "true")
		defer os.Unsetenv("USE_SUDO_FOR_DOCKER")
		cmd, args := GetDockerCommand()
		if cmd != "sudo" {
			t.Errorf("Expected cmd to be 'sudo', but got %q", cmd)
		}
		if len(args) != 1 || args[0] != "docker" {
			t.Errorf("Expected args to be ['docker'], but got %v", args)
		}
	})
}

func TestGenerateToolName(t *testing.T) {
	input := "test_tool"
	expected, _ := GenerateID(input)
	actual, err := GenerateToolName(input)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if actual != expected {
		t.Errorf("Expected %q, but got %q", expected, actual)
	}
}
