// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSummarizeMapResult(t *testing.T) {
	fakeB64 := base64.StdEncoding.EncodeToString([]byte("fake"))

	tests := []struct {
		name     string
		input    map[string]any
		checks   func(t *testing.T, attrs []slog.Attr)
	}{
		{
			name: "Valid Text",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "hello world",
					},
				},
				"isError": false,
			},
			checks: func(t *testing.T, attrs []slog.Attr) {
				// isError is false, so it's added. content is present, so it's added.
				// Wait, summarizeMapResult logic: if isError, ok := m["isError"].(bool); ok { ... }
				// If isError is present, it is added.

				// Let's verify content
				foundContent := false
				for _, a := range attrs {
					if a.Key == "content" {
						foundContent = true
						assert.Contains(t, fmt.Sprint(a.Value.Any()), "Text(len=11): \"hello world\"")
					}
				}
				assert.True(t, foundContent)
			},
		},
		{
			name: "Valid Image",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type":     "image",
						"data":     fakeB64,
						"mimeType": "image/png",
					},
				},
			},
			checks: func(t *testing.T, attrs []slog.Attr) {
				for _, a := range attrs {
					if a.Key == "content" {
						assert.Contains(t, fmt.Sprint(a.Value.Any()), "Image(mime=image/png, size=4 bytes)")
					}
				}
			},
		},
		{
			name: "Valid Resource with Blob",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "resource",
						"resource": map[string]any{
							"uri":  "file:///test.bin",
							"blob": fakeB64,
						},
					},
				},
			},
			checks: func(t *testing.T, attrs []slog.Attr) {
				for _, a := range attrs {
					if a.Key == "content" {
						s := fmt.Sprint(a.Value.Any())
						assert.Contains(t, s, "Resource(uri=file:///test.bin)")
						assert.Contains(t, s, "blob=4 bytes")
					}
				}
			},
		},
		{
			name: "Truncated Text",
			input: map[string]any{
				"content": []any{
					map[string]any{
						"type": "text",
						"text": strings.Repeat("a", 600),
					},
				},
			},
			checks: func(t *testing.T, attrs []slog.Attr) {
				for _, a := range attrs {
					if a.Key == "content" {
						s := fmt.Sprint(a.Value.Any())
						assert.Contains(t, s, "chars truncated")
						assert.Contains(t, s, "Text(len=600)")
					}
				}
			},
		},
		{
			name: "Error Only",
			input: map[string]any{
				"isError": true,
			},
			checks: func(t *testing.T, attrs []slog.Attr) {
				foundError := false
				for _, a := range attrs {
					if a.Key == "isError" {
						foundError = true
						assert.True(t, a.Value.Bool())
					}
				}
				assert.True(t, foundError)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, ok := summarizeMapResult(tt.input)
			assert.True(t, ok, "should be valid")
			assert.Equal(t, slog.KindGroup, val.Kind())
			tt.checks(t, val.Group())
		})
	}
}

func TestSummarizeMapResult_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input map[string]any
	}{
		{
			name: "Invalid Content Item",
			input: map[string]any{
				"content": []any{"not a map"},
			},
		},
		{
			name: "Content Not List",
			input: map[string]any{
				"content": "not a list",
			},
		},
		{
			name: "IsError Not Bool",
			input: map[string]any{
				"isError": "maybe",
			},
		},
		{
			name: "Missing Type",
			input: map[string]any{
				"content": []any{
					map[string]any{"text": "foo"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, ok := summarizeMapResult(tt.input)
			assert.False(t, ok, "should be invalid")
		})
	}
}

func TestLazyLogResult_MapOptimization(t *testing.T) {
	m := map[string]any{
		"content": []any{
			map[string]any{"type": "text", "text": "optimized"},
		},
	}
	val := LazyLogResult{Value: m}.LogValue()
	assert.Equal(t, slog.KindGroup, val.Kind())

	attrs := val.Group()
	found := false
	for _, a := range attrs {
		if a.Key == "content" {
			found = true
			assert.Contains(t, fmt.Sprint(a.Value.Any()), "Text(len=9)")
		}
	}
	assert.True(t, found, "should contain content summary")

	// Test fallback for sensitive/invalid data
	m2 := map[string]any{"secret": "sensitive"} // No content/isError -> fallback
	val2 := LazyLogResult{Value: m2}.LogValue()
	assert.Equal(t, slog.KindString, val2.Kind()) // Redacted JSON string

	m3 := map[string]any{"content": "sensitive"} // Invalid content -> fallback
	val3 := LazyLogResult{Value: m3}.LogValue()
	assert.Equal(t, slog.KindString, val3.Kind()) // Redacted JSON string
}
