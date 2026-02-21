// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type failWriter struct {
	err error
}

func (w *failWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}

func TestRedactingWriter_Write(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		expectedOutput string
		writerErr      error
		expectedN      int
		expectErr      bool
	}{
		{
			name:           "Happy Path: Valid JSON with sensitive key",
			input:          []byte(`{"password": "secret", "user": "alice"}`),
			expectedOutput: `{"password": "[REDACTED]", "user": "alice"}`,
			expectedN:      39, // Length of original input
			expectErr:      false,
		},
		{
			name:           "No Redaction Needed: Valid JSON without sensitive keys",
			input:          []byte(`{"user": "alice", "role": "admin"}`),
			expectedOutput: `{"user": "alice", "role": "admin"}`,
			expectedN:      34,
			expectErr:      false,
		},
		{
			name:           "Invalid JSON: Should pass through unchanged",
			input:          []byte(`{invalid-json`),
			expectedOutput: `{invalid-json`,
			expectedN:      13,
			expectErr:      false,
		},
		{
			name:           "Not JSON: Plain text should pass through",
			input:          []byte(`Just a plain text log message`),
			expectedOutput: `Just a plain text log message`,
			expectedN:      29,
			expectErr:      false,
		},
		{
			name:           "Partial JSON: Should redact if key is identified",
			input:          []byte(`{"password": "secre`), // Incomplete
			expectedOutput: `{"password": "[REDACTED]"`,
			expectedN:      19,
			expectErr:      false,
		},
		{
			name:           "Underlying Write Error",
			input:          []byte(`{"msg": "hello"}`),
			writerErr:      errors.New("disk full"),
			expectedOutput: "",
			expectedN:      0,
			expectErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var w *RedactingWriter

			if tt.writerErr != nil {
				w = &RedactingWriter{w: &failWriter{err: tt.writerErr}}
			} else {
				w = &RedactingWriter{w: &buf}
			}

			n, err := w.Write(tt.input)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.writerErr != nil {
					assert.Equal(t, tt.writerErr, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, buf.String())
			}

			assert.Equal(t, tt.expectedN, n)
		})
	}
}
