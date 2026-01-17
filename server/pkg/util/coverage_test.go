// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"math"
	"os"
	"testing"
)

func TestToString_Coverage(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"int", int(123), "123"},
		{"int8", int8(123), "123"},
		{"int16", int16(123), "123"},
		{"int32", int32(123), "123"},
		{"int64", int64(123), "123"},
		{"uint", uint(123), "123"},
		{"uint8", uint8(123), "123"},
		{"uint16", uint16(123), "123"},
		{"uint32", uint32(123), "123"},
		{"uint64", uint64(123), "123"},
		{"float32 exact", float32(123.0), "123"},
		{"float32 decimal", float32(123.456), "123.456"},
		{"float64 exact", float64(123.0), "123"},
		{"float64 decimal", float64(123.456), "123.456"},
		{"json.Number", json.Number("123.456"), "123.456"},
		{"string", "hello", "hello"},
		{"nil", nil, "<nil>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToString(tt.input); got != tt.expected {
				t.Errorf("ToString(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToString_Float64_Boundaries(t *testing.T) {
	// MaxInt64 is 9223372036854775807
	// float64(MaxInt64) is 9223372036854775808 (2^63) - which overflows int64

	// Test value just below MaxInt64 that IS representable as int64
	// 2^63 - 2048 = 9223372036854773760. This is exactly representable in float64.
	valSafe := float64(9223372036854773760)
	strSafe := ToString(valSafe)
	if strSafe != "9223372036854773760" {
		t.Errorf("ToString(2^63 - 2048) = %s, want 9223372036854773760", strSafe)
	}

	// Test value equal to MaxInt64 (rounded up to 2^63)
	// This should NOT use int64 formatting, but float formatting (scientific)
	valMax := float64(math.MaxInt64)
	strMax := ToString(valMax)
	// format 'g' with -1 precision for 9.223372036854776e+18
	expected := "9.223372036854776e+18"
	if strMax != expected {
		t.Errorf("ToString(float64(MaxInt64)) = %s, want %s", strMax, expected)
	}
}

func TestIsNil_Coverage(t *testing.T) {
	var i interface{}
	if !IsNil(i) {
		t.Error("IsNil(nil interface) should be true")
	}

	var p *int
	if !IsNil(p) {
		t.Error("IsNil(nil pointer) should be true")
	}

	var m map[string]string
	if !IsNil(m) {
		t.Error("IsNil(nil map) should be true")
	}

	var s []int
	if !IsNil(s) {
		t.Error("IsNil(nil slice) should be true")
	}

	var ch chan int
	if !IsNil(ch) {
		t.Error("IsNil(nil chan) should be true")
	}

	// Not nil
	x := 5
	if IsNil(x) {
		t.Error("IsNil(5) should be false")
	}

	if IsNil(&x) {
		t.Error("IsNil(&x) should be false")
	}
}

func TestSanitizeID_Coverage(t *testing.T) {
	// Empty IDs
	_, err := SanitizeID([]string{}, false, 10, 8)
	if err != nil {
		t.Errorf("SanitizeID(empty) returned error: %v", err)
	}

	_, err = SanitizeID([]string{"valid", ""}, false, 10, 8)
	if err == nil {
		t.Error("SanitizeID(valid, empty) should return error")
	}

	// Single ID optimization
	res, err := SanitizeID([]string{"valid-id"}, false, 50, 8)
	if err != nil || res != "valid-id" {
		t.Errorf("SanitizeID(valid-id) = %s, %v; want valid-id", res, err)
	}

	// Dirty chars force hash
	res, err = SanitizeID([]string{"invalid@char"}, false, 50, 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(res) < 12 {
		t.Errorf("SanitizeID(invalid@char) result too short: %s", res)
	}

	// Max sanitized length
	longID := "this-is-a-very-long-id-that-should-be-truncated-before-hashing"
	res, err = SanitizeID([]string{longID}, false, 10, 8)
	if err != nil {
		t.Fatal(err)
	}
	// It should take first 10 chars: "this-is-a-"
	// Then append hash.
	// "this-is-a-_" + hash
	if len(res) != 10 + 1 + 8 {
		t.Errorf("SanitizeID(longID) length = %d, want %d", len(res), 19)
	}
}

func TestReplaceURLPath_Coverage(t *testing.T) {
	path := "/api/{{version}}/resource/{{id}}"
	params := map[string]interface{}{
		"version": "v1",
		"id":      "123",
	}
	res := ReplaceURLPath(path, params, nil)
	expected := "/api/v1/resource/123"
	if res != expected {
		t.Errorf("ReplaceURLPath = %s, want %s", res, expected)
	}

	// Missing key
	path = "/api/{{missing}}"
	res = ReplaceURLPath(path, params, nil)
	if res != "/api/{{missing}}" {
		t.Errorf("ReplaceURLPath(missing) = %s, want /api/{{missing}}", res)
	}

	// Escaping
	path = "/search/{{query}}"
	params = map[string]interface{}{
		"query": "foo bar",
	}
	res = ReplaceURLPath(path, params, nil)
	// PathEscape converts space to %20
	if res != "/search/foo%20bar" {
		t.Errorf("ReplaceURLPath(space) = %s, want /search/foo%%20bar", res)
	}

	// No escaping
	noEscape := map[string]bool{"query": true}
	res = ReplaceURLPath(path, params, noEscape)
	if res != "/search/foo bar" {
		t.Errorf("ReplaceURLPath(noEscape) = %s, want /search/foo bar", res)
	}
}

func TestReplaceURLQuery_Coverage(t *testing.T) {
	query := "q={{query}}&page={{page}}"
	params := map[string]interface{}{
		"query": "foo bar",
		"page":  1,
	}
	res := ReplaceURLQuery(query, params, nil)
	// QueryEscape converts space to +
	if res != "q=foo+bar&page=1" {
		t.Errorf("ReplaceURLQuery = %s, want q=foo+bar&page=1", res)
	}
}

func TestExtractIP_Coverage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1.2.3.4", "1.2.3.4"},
		{"1.2.3.4:80", "1.2.3.4"},
		{"[::1]", "::1"},
		{"[::1]:80", "::1"},
		{"fe80::1%eth0", "fe80::1"},
		{"[fe80::1%eth0]:80", "fe80::1"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := ExtractIP(tt.input); got != tt.expected {
			t.Errorf("ExtractIP(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestRedactMap_Coverage(t *testing.T) {
	input := map[string]interface{}{
		"normal": "value",
		"password": "secret",
		"nested": map[string]interface{}{
			"api_key": "12345",
			"public": "visible",
		},
		"list": []interface{}{
			"item1",
			map[string]interface{}{"token": "s3cr3t"},
			[]interface{}{"nested-list"},
		},
	}

	output := RedactMap(input)

	if output["normal"] != "value" {
		t.Error("RedactMap modified normal key")
	}
	if output["password"] != redactedPlaceholder {
		t.Errorf("RedactMap failed to redact password: %v", output["password"])
	}

	nested := output["nested"].(map[string]interface{})
	if nested["api_key"] != redactedPlaceholder {
		t.Errorf("RedactMap failed to redact nested api_key")
	}
	if nested["public"] != "visible" {
		t.Error("RedactMap modified public key")
	}

	list := output["list"].([]interface{})
	if list[0] != "item1" {
		t.Error("RedactMap modified list item")
	}
	listMap := list[1].(map[string]interface{})
	if listMap["token"] != redactedPlaceholder {
		t.Errorf("RedactMap failed to redact list map token")
	}
}

func TestIsSensitiveKey_Coverage(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"password", true},
		{"PASSWORD", true},
		{"PassWord", true},
		{"my_api_key", true}, // contains api_key
		{"token", true},
		{"access_token", true}, // contains token
		{"auth", true},
		{"author", false}, // boundary check
		{"AUTH", true},
		{"AUTHOR", false}, // boundary check
		{"AuthToken", true},
		{"not_sensitive", false},
		{"a", false},
		{"ap", false},
		{"api_ke", false},
	}

	for _, tt := range tests {
		if got := IsSensitiveKey(tt.key); got != tt.expected {
			t.Errorf("IsSensitiveKey(%q) = %v, want %v", tt.key, got, tt.expected)
		}
	}
}

func TestSanitizeOperationID_Coverage(t *testing.T) {
	// Clean path
	if SanitizeOperationID("valid-id") != "valid-id" {
		t.Error("clean path failed")
	}

	// Dirty path
	res := SanitizeOperationID("invalid space")
	if len(res) <= len("invalid") {
		t.Errorf("SanitizeOperationID failed: %s", res)
	}

	// Consecutive dirty chars
	res = SanitizeOperationID("a@@b")
	if len(res) <= 3 {
		t.Errorf("SanitizeOperationID failed: %s", res)
	}
}

func TestGetDockerCommand_Coverage(t *testing.T) {
	// Default
	cmd, args := GetDockerCommand()
	if cmd != "docker" || len(args) != 0 {
		t.Error("default docker command incorrect")
	}

	// Sudo
	os.Setenv("USE_SUDO_FOR_DOCKER", "true")
	defer os.Unsetenv("USE_SUDO_FOR_DOCKER")
	cmd, args = GetDockerCommand()
	if cmd != "sudo" || len(args) != 1 || args[0] != "docker" {
		t.Error("sudo docker command incorrect")
	}
}

func TestParseToolName_Coverage(t *testing.T) {
	svc, tool, err := ParseToolName("service.tool")
	if err != nil || svc != "service" || tool != "tool" {
		t.Error("ParseToolName failed")
	}

	svc, tool, err = ParseToolName("tool-only")
	if err != nil || svc != "" || tool != "tool-only" {
		t.Error("ParseToolName failed for tool only")
	}
}

func TestRedactJSONFast_Coverage(t *testing.T) {
	// 1. Nested objects and arrays
	input := []byte(`{"a": {"b": [1, 2, {"c": "d"}]}, "password": "secret", "arr": [true, false, null]}`)
	expected := `{"a": {"b": [1, 2, {"c": "d"}]}, "password": "[REDACTED]", "arr": [true, false, null]}`

	got := RedactJSON(input)
	if string(got) != expected {
		t.Errorf("RedactJSON(nested) = %s, want %s", got, expected)
	}

	// 2. Escaped keys
	// "p\u0061ssword": "secret" -> password
	input = []byte(`{"p\u0061ssword": "secret"}`)
	expected = `{"p\u0061ssword": "[REDACTED]"}`
	got = RedactJSON(input)
	if string(got) != expected {
		t.Errorf("RedactJSON(escaped) = %s, want %s", got, expected)
	}

	// 3. Weird escapes
	// "to\u006ben": "secret" (token)
	input = []byte(`{"to\u006ben": "secret"}`)
	expected = `{"to\u006ben": "[REDACTED]"}`
	got = RedactJSON(input)
	if string(got) != expected {
		t.Errorf("RedactJSON(escaped-token) = %s, want %s", got, expected)
	}

	// 4. Malformed JSON (missing brace)
	input = []byte(`{"password": "secret"`)
	expected = `{"password": "[REDACTED]"`
	got = RedactJSON(input)
	if string(got) != expected {
		t.Errorf("RedactJSON(malformed) = %s, want %s", got, expected)
	}

	// 5. Literal values skipping
	input = []byte(`{"key": true, "password": "secret"}`)
	expected = `{"key": true, "password": "[REDACTED]"}`
	got = RedactJSON(input)
	if string(got) != expected {
		t.Errorf("RedactJSON(literal) = %s, want %s", got, expected)
	}

	// 6. Number skipping
	input = []byte(`{"k": 123.456, "password": "secret"}`)
	expected = `{"k": 123.456, "password": "[REDACTED]"}`
	got = RedactJSON(input)
	if string(got) != expected {
		t.Errorf("RedactJSON(number) = %s, want %s", got, expected)
	}
}

func TestRedactJSONFast_LongKey(t *testing.T) {
	// Temporarily lower limit to trigger large key path
	oldLimit := maxUnescapeLimit
	maxUnescapeLimit = 10
	defer func() { maxUnescapeLimit = oldLimit }()

	// Key longer than 10 bytes: "very_long_password_key"
	// It contains "password".
	input := []byte(`{"very_long_password_key": "secret"}`)
	expected := `{"very_long_password_key": "[REDACTED]"}`
	got := RedactJSON(input)
	if string(got) != expected {
		t.Errorf("RedactJSON(long key) = %s, want %s", got, expected)
	}
}

// Note: Need to import "context" and "google.golang.org/grpc"?
// WrappedServerStream embeds grpc.ServerStream which is an interface.
// If I use &WrappedServerStream{}, I don't need to implement methods if I don't call them,
// except Context().
// But I need to import "context". "google.golang.org/grpc" is needed if I reference grpc.ServerStream?
// Yes.

// I need to add imports to the top of the file.
// Since I cannot edit the file easily to add imports, I will create a NEW file for these tests.
