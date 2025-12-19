// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyPrintExtended(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		contentType string
		expected    string
		isXml       bool
	}{
		{
			name:        "Empty input",
			input:       []byte{},
			contentType: "application/json",
			expected:    "",
		},
		{
			name:        "Binary image",
			input:       make([]byte, 100),
			contentType: "image/png",
			expected:    "[Binary Data: 100 bytes]",
		},
		{
			name:        "Binary audio",
			input:       make([]byte, 50),
			contentType: "audio/mpeg",
			expected:    "[Binary Data: 50 bytes]",
		},
		{
			name:        "Binary video",
			input:       make([]byte, 200),
			contentType: "video/mp4",
			expected:    "[Binary Data: 200 bytes]",
		},
		{
			name:        "Octet stream",
			input:       make([]byte, 10),
			contentType: "application/octet-stream",
			expected:    "[Binary Data: 10 bytes]",
		},
		{
			name:        "Valid JSON",
			input:       []byte(`{"key":"value"}`),
			contentType: "application/json",
			expected:    "{\n  \"key\": \"value\"\n}",
		},
		{
			name:        "Invalid JSON",
			input:       []byte(`{invalid`),
			contentType: "application/json",
			expected:    "{invalid",
		},
		{
			name:        "Valid XML",
			input:       []byte(`<root><child>value</child></root>`),
			contentType: "application/xml",
			isXml:       true,
		},
		{
			name:        "Invalid XML",
			input:       []byte(`<root>invalid`),
			contentType: "application/xml",
			expected:    "<root>invalid",
		},
		{
			name:        "Plain text",
			input:       []byte("hello world"),
			contentType: "text/plain",
			expected:    "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := prettyPrint(tt.input, tt.contentType)
			if tt.isXml {
				assert.Contains(t, res, "\n")
				assert.Contains(t, res, "  ")
				assert.Contains(t, res, "<root>")
			} else {
				assert.Equal(t, tt.expected, res)
			}
		})
	}
}

func TestCacheControlContext(t *testing.T) {
	ctx := context.Background()

	// Initially empty
	cc, ok := GetCacheControl(ctx)
	assert.False(t, ok)
	assert.Nil(t, cc)

	// Set value
	control := &CacheControl{Action: ActionAllow}
	ctx = NewContextWithCacheControl(ctx, control)

	cc, ok = GetCacheControl(ctx)
	assert.True(t, ok)
	assert.Equal(t, control, cc)
	assert.Equal(t, ActionAllow, cc.Action)

	// Modify
	cc.Action = ActionSaveCache
	cc2, _ := GetCacheControl(ctx)
	assert.Equal(t, ActionSaveCache, cc2.Action)
}
