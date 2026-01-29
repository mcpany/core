package middleware

import (
	"log/slog"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestNewRedactor(t *testing.T) {
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())
	assert.NotNil(t, r)
	assert.Len(t, r.customPatterns, 1)

	// Test with invalid regex
	cfg.SetCustomPatterns([]string{`[`})
	r = NewRedactor(cfg, slog.Default())
	assert.NotNil(t, r)
	assert.Len(t, r.customPatterns, 0)

	// Test with nil config
	r = NewRedactor(nil, slog.Default())
	assert.Nil(t, r)

	// Test with disabled config
	disabled := false
	cfg.SetEnabled(disabled)
	r = NewRedactor(cfg, slog.Default())
	assert.Nil(t, r)
}

func TestRedactString(t *testing.T) {
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No redaction needed",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Email redaction",
			input:    "Contact me at user@example.com",
			expected: "Contact me at ***REDACTED***",
		},
		{
			name:     "Credit card redaction",
			input:    "My card is 1234-5678-9012-3456",
			expected: "My card is ***REDACTED***",
		},
		{
			name:     "SSN redaction",
			input:    "My SSN is 123-45-6789",
			expected: "My SSN is ***REDACTED***",
		},
		{
			name:     "Custom pattern redaction",
			input:    "This is secret-123",
			expected: "This is ***REDACTED***",
		},
		{
			name:     "Multiple redactions",
			input:    "Email: user@example.com, Secret: secret-999",
			expected: "Email: ***REDACTED***, Secret: ***REDACTED***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, r.RedactString(tt.input))
		})
	}
}

func TestRedactString_NilRedactor(t *testing.T) {
	var r *Redactor
	input := "user@example.com"
	assert.Equal(t, input, r.RedactString(input))
}

func TestRedactStruct(t *testing.T) {
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	input := map[string]interface{}{
		"name":  "John Doe",
		"email": "john@example.com",
		"nested": map[string]interface{}{
			"secret": "secret-123",
			"list": []interface{}{
				"123-45-6789",
				"safe",
			},
		},
	}

	expected := map[string]interface{}{
		"name":  "John Doe",
		"email": "***REDACTED***",
		"nested": map[string]interface{}{
			"secret": "***REDACTED***",
			"list": []interface{}{
				"***REDACTED***",
				"safe",
			},
		},
	}

	r.RedactStruct(input)
	assert.Equal(t, expected, input)
}

func TestRedactStruct_NilRedactor(t *testing.T) {
	var r *Redactor
	input := map[string]interface{}{
		"email": "user@example.com",
	}
	r.RedactStruct(input)
	assert.Equal(t, "user@example.com", input["email"])
}

func TestRedactValue(t *testing.T) {
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "String",
			input:    "user@example.com",
			expected: "***REDACTED***",
		},
		{
			name: "Map",
			input: map[string]interface{}{
				"key": "secret-123",
			},
			expected: map[string]interface{}{
				"key": "***REDACTED***",
			},
		},
		{
			name: "Slice",
			input: []interface{}{
				"secret-123",
				"safe",
			},
			expected: []interface{}{
				"***REDACTED***",
				"safe",
			},
		},
		{
			name:     "StructPb Value",
			input:    &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "test"}},
			expected: &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: "test"}},
		},
		{
			name:     "Other type",
			input:    123,
			expected: 123,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, r.RedactValue(tt.input))
		})
	}
}

func TestRedactValue_NilRedactor(t *testing.T) {
	var r *Redactor
	input := "user@example.com"
	assert.Equal(t, input, r.RedactValue(input))
}

func TestRedactJSON(t *testing.T) {
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled:        &enabled,
		CustomPatterns: []string{`secret-\d+`},
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	input := []byte(`{
		"email": "user@example.com",
		"details": {
			"secret": "secret-123"
		}
	}`)

	// Note: WalkJSONStrings might reformat whitespace, but here we are checking for redaction.
	// Actually WalkJSONStrings does not reformat, it replaces strings in place if length matches or reconstructs if not.
	// RedactString returns "***REDACTED***", which is likely different length than "user@example.com".

	output, err := r.RedactJSON(input)
	assert.NoError(t, err)

	outputStr := string(output)
	assert.Contains(t, outputStr, "***REDACTED***")
	assert.NotContains(t, outputStr, "user@example.com")
	assert.NotContains(t, outputStr, "secret-123")
}

func TestRedactJSON_NilRedactor(t *testing.T) {
	var r *Redactor
	input := []byte(`{"email":"user@example.com"}`)
	output, err := r.RedactJSON(input)
	assert.NoError(t, err)
	assert.Equal(t, input, output)
}

func TestRedactJSON_Empty(t *testing.T) {
	enabled := true
	cfg := configv1.DLPConfig_builder{
		Enabled: &enabled,
	}.Build()
	r := NewRedactor(cfg, slog.Default())

	output, err := r.RedactJSON(nil)
	assert.NoError(t, err)
	assert.Nil(t, output)

	output, err = r.RedactJSON([]byte{})
	assert.NoError(t, err)
	assert.Empty(t, output)
}
