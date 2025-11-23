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
//
// SPDX-License-Identifier: Apache-2.0


package util

import (
	"regexp"
	"strings"
	"testing"
)

func TestSanitizeID(t *testing.T) {
	testCases := []struct {
		name                     string
		ids                      []string
		alwaysAppendHash         bool
		maxSanitizedPrefixLength int
		hashLength               int
		expected                 string
		expectError              bool
	}{
		{
			name:                     "single id, no hash",
			ids:                      []string{"test"},
			alwaysAppendHash:         false,
			maxSanitizedPrefixLength: 10,
			hashLength:               8,
			expected:                 "test",
			expectError:              false,
		},
		{
			name:                     "single id, with hash",
			ids:                      []string{"test"},
			alwaysAppendHash:         true,
			maxSanitizedPrefixLength: 10,
			hashLength:               8,
			expected:                 "test_9f86d081",
			expectError:              false,
		},
		{
			name:                     "multiple ids, no hash",
			ids:                      []string{"test", "service"},
			alwaysAppendHash:         false,
			maxSanitizedPrefixLength: 10,
			hashLength:               8,
			expected:                 "test.service",
			expectError:              false,
		},
		{
			name:                     "multiple ids, with hash",
			ids:                      []string{"test", "service"},
			alwaysAppendHash:         true,
			maxSanitizedPrefixLength: 10,
			hashLength:               8,
			expected:                 "test_9f86d081.service_9df6b026",
			expectError:              false,
		},
		{
			name:                     "long id, with hash",
			ids:                      []string{strings.Repeat("a", 20)},
			alwaysAppendHash:         false,
			maxSanitizedPrefixLength: 10,
			hashLength:               8,
			expected:                 "aaaaaaaaaa_42492da0",
			expectError:              false,
		},
		{
			name:        "empty id",
			ids:         []string{""},
			expectError: true,
		},
		{
			name:                     "id with only non-word characters",
			ids:                      []string{"@@@"},
			alwaysAppendHash:         false,
			maxSanitizedPrefixLength: 10,
			hashLength:               8,
			expected:                 "id_2ec847d8",
			expectError:              false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := SanitizeID(tc.ids, tc.alwaysAppendHash, tc.maxSanitizedPrefixLength, tc.hashLength)

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
			}
		})
	}
}

func TestSanitizeServiceName(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid service name",
			input:       "test_service",
			expected:    "test_service",
			expectError: false,
		},
		{
			name:        "service name with special characters",
			input:       "test-service-1.0",
			expected:    "test-service-10_57f3fff2",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := SanitizeServiceName(tc.input)

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
			}
		})
	}
}

func TestSanitizeToolName(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    string
		expectError bool
	}{
		{
			name:        "valid tool name",
			input:       "test_tool",
			expected:    "test_tool",
			expectError: false,
		},
		{
			name:        "tool name with special characters",
			input:       "test-tool-1.0",
			expected:    "test-tool-10_8c924588",
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := SanitizeToolName(tc.input)

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
			name:     "no disallowed characters",
			input:    "get-user-by-id",
			expected: "get-user-by-id",
		},
		{
			name:     "with disallowed characters",
			input:    "get user by id",
			expected: "get_36a9e7_user_36a9e7_by_36a9e7_id",
		},
		{
			name:     "with multiple disallowed characters",
			input:    "get user by id (new)",
			expected: "get_36a9e7_user_36a9e7_by_36a9e7_id_36a9e7_(new)",
		},
		{
			name:     "with consecutive disallowed characters",
			input:    "get  user",
			expected: "get_6c179f_user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := SanitizeOperationID(tc.input)
			if actual != tc.expected {
				t.Errorf("Expected %q, but got %q", tc.expected, actual)
			}
		})
	}
}

func TestGetDockerCommand(t *testing.T) {
	t.Run("without sudo", func(t *testing.T) {
		t.Setenv("USE_SUDO_FOR_DOCKER", "false")
		cmd, args := GetDockerCommand()
		if cmd != "docker" {
			t.Errorf("Expected command to be 'docker', but got %q", cmd)
		}
		if len(args) != 0 {
			t.Errorf("Expected no arguments, but got %v", args)
		}
	})

	t.Run("with sudo", func(t *testing.T) {
		t.Setenv("USE_SUDO_FOR_DOCKER", "true")
		cmd, args := GetDockerCommand()
		if cmd != "sudo" {
			t.Errorf("Expected command to be 'sudo', but got %q", cmd)
		}
		if len(args) != 1 || args[0] != "docker" {
			t.Errorf("Expected arguments to be ['docker'], but got %v", args)
		}
	})
}

func TestGenerateUUID(t *testing.T) {
	uuid := GenerateUUID()
	match, err := regexp.MatchString(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`, uuid)
	if err != nil {
		t.Fatalf("Error matching UUID regex: %v", err)
	}
	if !match {
		t.Errorf("Generated UUID %q does not match the expected format", uuid)
	}
}

func TestParseToolName(t *testing.T) {
	testCases := []struct {
		name               string
		toolName           string
		expectedService    string
		expectedBareTool   string
		expectError        bool
	}{
		{
			name:             "valid tool name",
			toolName:         "service.tool",
			expectedService:  "service",
			expectedBareTool: "tool",
			expectError:      false,
		},
		{
			name:             "no service name",
			toolName:         "tool",
			expectedService:  "",
			expectedBareTool: "tool",
			expectError:      false,
		},
		{
			name:             "empty tool name",
			toolName:         "",
			expectedService:  "",
			expectedBareTool: "",
			expectError:      false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, bareTool, err := ParseToolName(tc.toolName)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if service != tc.expectedService {
					t.Errorf("Expected service %q, but got %q", tc.expectedService, service)
				}
				if bareTool != tc.expectedBareTool {
					t.Errorf("Expected bare tool %q, but got %q", tc.expectedBareTool, bareTool)
				}
			}
		})
	}
}
