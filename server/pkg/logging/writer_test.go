// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mockWriter is a helper writer that can simulate errors.
type mockWriter struct {
	writeFunc func(p []byte) (n int, err error)
}

func (m *mockWriter) Write(p []byte) (n int, err error) {
	if m.writeFunc != nil {
		return m.writeFunc(p)
	}
	return len(p), nil
}

func TestRedactingWriter_Write(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		writeErr       error
		expectedOutput string
		expectError    bool
	}{
		{
			name:           "Happy Path: Valid JSON with sensitive keys",
			// Verify that known sensitive keys like "password" are redacted.
			input:          []byte(`{"password": "secret", "user": "alice"}`),
			expectedOutput: `{"password": "[REDACTED]", "user": "alice"}`,
		},
		{
			name:           "No Redaction Needed",
			input:          []byte(`{"user": "alice", "role": "admin"}`),
			expectedOutput: `{"user": "alice", "role": "admin"}`,
		},
		{
			name:           "Not JSON: Plain Text",
			input:          []byte(`plain text message`),
			expectedOutput: `plain text message`,
		},
		{
			name:           "Invalid JSON: Malformed",
			input:          []byte(`{"password": "secret"`), // Missing closing brace
			expectedOutput: `{"password": "[REDACTED]"`,
		},
		{
			name:           "Partial JSON: Just a brace",
			input:          []byte(`{`),
			expectedOutput: `{`,
		},
		{
			name:           "Empty Input",
			input:          []byte(``),
			expectedOutput: ``,
		},
		{
			name:        "Underlying Write Error",
			input:       []byte(`{"msg": "hello"}`),
			writeErr:    errors.New("disk full"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			var w *RedactingWriter

			if tt.writeErr != nil {
				// Use mock writer for error case
				mock := &mockWriter{
					writeFunc: func(p []byte) (n int, err error) {
						return 0, tt.writeErr
					},
				}
				w = &RedactingWriter{w: mock}
			} else {
				// Use real buffer for success cases
				w = &RedactingWriter{w: &buf}
			}

			n, err := w.Write(tt.input)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.writeErr, err)
				assert.Equal(t, 0, n)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.input), n, "Write should return length of original input")
				assert.Equal(t, tt.expectedOutput, buf.String())
			}
		})
	}
}
