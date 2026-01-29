package util //nolint:revive,nolintlint // Package name 'util' is common in this codebase

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"
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
		{
			name:                     "max sanitized length zero",
			ids:                      []string{"abc"},
			alwaysAppendHash:         false,
			maxSanitizedPrefixLength: 0,
			hashLength:               8,
			expected:                 "id_ba7816bf",
			expectError:              false,
		},
		{
			name:                     "hash length zero defaults to 8",
			ids:                      []string{"abc"},
			alwaysAppendHash:         true,
			maxSanitizedPrefixLength: 10,
			hashLength:               0,
			expected:                 "abc_ba7816bf",
			expectError:              false,
		},
		{
			name:                     "hash length negative defaults to 8",
			ids:                      []string{"abc"},
			alwaysAppendHash:         true,
			maxSanitizedPrefixLength: 10,
			hashLength:               -1,
			expected:                 "abc_ba7816bf",
			expectError:              false,
		},
		{
			name:                     "hash length too large caps at 64",
			ids:                      []string{"abc"},
			alwaysAppendHash:         true,
			maxSanitizedPrefixLength: 10,
			hashLength:               100,
			expected:                 "abc_ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
			expectError:              false,
		},
		{
			name:        "empty ids list",
			ids:         []string{},
			expectError: false,
			expected:    "",
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

type StringerStruct struct {
	val string
}

func (s StringerStruct) String() string {
	return s.val
}

func TestToString(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string", "hello", "hello"},
		{"json.Number", json.Number("123.45"), "123.45"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"error", fmt.Errorf("some error"), "some error"},
		{"int", 123, "123"},
		{"int8", int8(123), "123"},
		{"int16", int16(123), "123"},
		{"int32", int32(123), "123"},
		{"int64", int64(123), "123"},
		{"uint", uint(123), "123"},
		{"uint8", uint8(123), "123"},
		{"uint16", uint16(123), "123"},
		{"uint32", uint32(123), "123"},
		{"uint64", uint64(123), "123"},
		{"float32", float32(123.456), "123.456"},
		{"float64", float64(123.456), "123.456"},
		{"float64 MaxInt64", float64(math.MaxInt64), "9.223372036854776e+18"},
		{"fmt.Stringer", StringerStruct{"stringer"}, "stringer"},
		{"default (struct)", struct{ A int }{A: 1}, "{1}"},
		{"default (nil)", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ToString(tt.input))
		})
	}

	// Ptr coverage
	i := 123
	ptr := &i
	if len(ToString(ptr)) == 0 {
		t.Error("ToString(ptr) is empty")
	}

}

// TestIsNil removed to avoid duplication with isnil_test.go

func TestReplaceURLPath_Security(t *testing.T) {
	tests := []struct {
		name     string
		urlPath  string
		params   map[string]interface{}
		expected string
	}{
		{
			name:     "prevent path traversal",
			urlPath:  "/files/{{name}}",
			params:   map[string]interface{}{"name": "../secret"},
			expected: "/files/..%2Fsecret",
		},
		{
			name:     "prevent path injection",
			urlPath:  "/files/{{name}}",
			params:   map[string]interface{}{"name": "dir/file"},
			expected: "/files/dir%2Ffile",
		},
		{
			name:     "prevent query injection",
			urlPath:  "/files/{{name}}",
			params:   map[string]interface{}{"name": "file?admin=true"},
			expected: "/files/file%3Fadmin=true",
		},
		{
			name:     "nil value",
			urlPath:  "/{{key}}",
			params:   map[string]interface{}{"key": nil},
			expected: "/%3Cnil%3E",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceURLPath(tt.urlPath, tt.params, nil)
			assert.Equal(t, tt.expected, result)
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
		name             string
		toolName         string
		expectedService  string
		expectedBareTool string
		expectError      bool
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

func TestReplaceURLPath(t *testing.T) {
	tests := []struct {
		name           string
		urlPath        string
		params         map[string]interface{}
		noEscapeParams map[string]bool
		expected       string
	}{
		{
			name:     "replace single param",
			urlPath:  "/api/v1/users/{{userID}}",
			params:   map[string]interface{}{"userID": "123"},
			expected: "/api/v1/users/123",
		},
		{
			name:     "replace multiple params",
			urlPath:  "/api/{{version}}/items/{{itemID}}",
			params:   map[string]interface{}{"version": "v2", "itemID": "456"},
			expected: "/api/v2/items/456",
		},
		{
			name:     "replace with integer param",
			urlPath:  "/items/{{id}}",
			params:   map[string]interface{}{"id": 789},
			expected: "/items/789",
		},
		{
			name:     "no placeholders",
			urlPath:  "/static/path",
			params:   map[string]interface{}{"id": "123"},
			expected: "/static/path",
		},
		{
			name:     "missing param value",
			urlPath:  "/users/{{userID}}",
			params:   map[string]interface{}{"other": "value"},
			expected: "/users/{{userID}}",
		},
		{
			name:     "unclosed placeholder",
			urlPath:  "/users/{{userID",
			params:   map[string]interface{}{"userID": "123"},
			expected: "/users/{{userID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceURLPath(tt.urlPath, tt.params, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestReplaceURLQuery(t *testing.T) {
	tests := []struct {
		name           string
		urlQuery       string
		params         map[string]interface{}
		noEscapeParams map[string]bool
		expected       string
	}{
		{
			name:     "replace single param",
			urlQuery: "q={{query}}",
			params:   map[string]interface{}{"query": "hello world"},
			expected: "q=hello+world", // QueryEscape encodes space as +
		},
		{
			name:     "replace with special chars",
			urlQuery: "q={{query}}",
			params:   map[string]interface{}{"query": "a&b"},
			expected: "q=a%26b", // QueryEscape encodes &
		},
		{
			name:           "no escape param",
			urlQuery:       "q={{query}}",
			params:         map[string]interface{}{"query": "a&b"},
			noEscapeParams: map[string]bool{"query": true},
			expected:       "q=a&b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ReplaceURLQuery(tt.urlQuery, tt.params, tt.noEscapeParams)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRandomFloat64(t *testing.T) {
	val := RandomFloat64()
	if val < 0.0 || val >= 1.0 {
		t.Errorf("RandomFloat64() = %v, want [0.0, 1.0)", val)
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "valid filename",
			input:    "safe_file.txt",
			expected: "safe_file.txt",
		},
		{
			name:     "traversal attempt",
			input:    "../../etc/passwd",
			expected: "passwd",
		},
		{
			name:     "weird characters",
			input:    "my$file!.txt",
			expected: "my_file_.txt",
		},
		{
			name:     "null bytes",
			input:    "file\x00name.txt",
			expected: "filename.txt",
		},
		{
			name:     "dots and dashes",
			input:    ".config-file.yaml",
			expected: ".config-file.yaml",
		},
		{
			name:     "empty result",
			input:    "$$$",
			expected: "___",
		},
		{
			name:     "reserved names",
			input:    ".",
			expected: "unnamed_file",
		},
		{
			name:     "reserved names 2",
			input:    "..",
			expected: "unnamed_file",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "unnamed_file",
		},
		{
			name:     "long filename",
			input:    strings.Repeat("a", 300),
			expected: strings.Repeat("a", 255),
		},
		{
			name:     "unicode characters preserved",
			input:    "测试.txt",
			expected: "测试.txt",
		},
		{
			name: "unicode truncation",
			// "a" * 254 + "あ" (3 bytes).
			// Total 257 bytes. Truncate to 255.
			// "あ" is E3 81 82.
			// [254] = E3.
			// Should remove E3.
			// Result length 254.
			input:    strings.Repeat("a", 254) + "あ",
			expected: strings.Repeat("a", 254),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, SanitizeFilename(tt.input))
		})
	}
}
