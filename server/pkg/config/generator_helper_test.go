// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeHTTPMethod(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GET", "HTTP_METHOD_GET"},
		{"post", "HTTP_METHOD_POST"},
		{"Put", "HTTP_METHOD_PUT"},
		{"DELETE", "HTTP_METHOD_DELETE"},
		{"PATCH", "HTTP_METHOD_PATCH"},
		// These are currently buggy, they return HTTP_METHOD_... but should probably not be normalized
		// if they are not supported by the proto enum, OR they should just return as is (which is also invalid, but at least we don't pretend we support them).
		// However, if the intention of this function is to "normalize input to valid enum string", it should only do so for valid enums.
		{"HEAD", "HEAD"},
		{"OPTIONS", "OPTIONS"},
		{"CONNECT", "CONNECT"},
		{"TRACE", "TRACE"},
		{"UNKNOWN", "UNKNOWN"},
		{"HTTP_METHOD_GET", "HTTP_METHOD_GET"}, // Should fall through default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeHTTPMethod(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
