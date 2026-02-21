// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWriter is a mock implementation of io.Writer
type MockWriter struct {
	mock.Mock
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	args := m.Called(p)
	return args.Int(0), args.Error(1)
}

func TestRedactingWriter_Write(t *testing.T) {
	tests := []struct {
		name         string
		input        []byte
		expectedOut  []byte
		mockError    error
		expectError  bool
		expectOutput bool // If false, we expect output to be the original input (pass-through)
	}{
		{
			name:        "Happy Path: Valid JSON with sensitive keys",
			input:       []byte(`{"user": "alice", "password": "secret"}`),
			expectedOut: []byte(`{"user": "alice", "password": "[REDACTED]"}`),
		},
		{
			name:        "Happy Path: Valid JSON with no sensitive keys",
			input:       []byte(`{"user": "alice", "role": "admin"}`),
			expectedOut: []byte(`{"user": "alice", "role": "admin"}`),
		},
		{
			name:        "Invalid JSON: Malformed string",
			input:       []byte(`{invalid-json`),
			expectedOut: []byte(`{invalid-json`),
		},
		{
			name:        "Not JSON: Plain text",
			input:       []byte(`This is a plain text log message`),
			expectedOut: []byte(`This is a plain text log message`),
		},
		{
			name:        "Partial JSON: Incomplete object",
			input:       []byte(`{"user": "alice", `),
			expectedOut: []byte(`{"user": "alice", `),
		},
		{
			name:        "Complex JSON: Nested object",
			input:       []byte(`{"config": {"api_key": "12345"}}`),
			expectedOut: []byte(`{"config": {"api_key": "[REDACTED]"}}`),
		},
		{
			name:        "Underlying Write Error",
			input:       []byte(`{"message": "hello"}`),
			expectedOut: []byte(`{"message": "hello"}`), // Output is passed through
			mockError:   errors.New("write failed"),
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWriter := new(MockWriter)
			rw := &RedactingWriter{w: mockWriter}

			// Determine expected write call
			var expectedWrite []byte
			if tt.expectedOut != nil {
				expectedWrite = tt.expectedOut
			} else {
				expectedWrite = tt.input
			}

			// Mock behavior
			if tt.expectError {
				mockWriter.On("Write", mock.Anything).Return(0, tt.mockError)
			} else {
				mockWriter.On("Write", mock.MatchedBy(func(p []byte) bool {
					// Compare byte slices ignoring whitespace differences if needed,
					// but RedactJSON typically returns compact JSON or original bytes.
					// For strict equality:
					return bytes.Equal(p, expectedWrite)
				})).Return(len(expectedWrite), nil)
			}

			n, err := rw.Write(tt.input)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.mockError, err)
				assert.Equal(t, 0, n)
			} else {
				require.NoError(t, err)
				assert.Equal(t, len(tt.input), n)
			}

			mockWriter.AssertExpectations(t)
		})
	}
}

// TestRedactingWriter_Write_Integration uses a real bytes.Buffer instead of a mock
// to verify the actual data written without mocking.
func TestRedactingWriter_Write_Integration(t *testing.T) {
	var buf bytes.Buffer
	rw := &RedactingWriter{w: &buf}

	input := []byte(`{"password": "secret"}`)
	n, err := rw.Write(input)

	require.NoError(t, err)
	assert.Equal(t, len(input), n) // Must pretend to write full length

	expected := `{"password": "[REDACTED]"}`
	assert.Equal(t, expected, buf.String())
}
