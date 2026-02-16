// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockWriter struct {
	written []byte
	err     error
}

func (m *MockWriter) Write(p []byte) (n int, err error) {
	if m.err != nil {
		return 0, m.err
	}
	m.written = append(m.written, p...)
	return len(p), nil
}

// Ensure MockWriter implements io.Writer
var _ io.Writer = (*MockWriter)(nil)

func TestRedactingWriter_Write(t *testing.T) {
	t.Run("Redaction", func(t *testing.T) {
		mock := &MockWriter{}
		w := &RedactingWriter{w: mock}

		input := []byte(`{"api_key": "secret123", "public": "value"}`)
		n, err := w.Write(input)
		require.NoError(t, err)
		assert.Equal(t, len(input), n)

		output := string(mock.written)
		assert.Contains(t, output, `"api_key": "[REDACTED]"`)
		assert.Contains(t, output, `"public": "value"`)
	})

	t.Run("NonJSON", func(t *testing.T) {
		mock := &MockWriter{}
		w := &RedactingWriter{w: mock}

		input := []byte("plain text log message")
		n, err := w.Write(input)
		require.NoError(t, err)
		assert.Equal(t, len(input), n)

		assert.Equal(t, "plain text log message", string(mock.written))
	})

	t.Run("WriteError", func(t *testing.T) {
		mock := &MockWriter{err: errors.New("write failed")}
		w := &RedactingWriter{w: mock}

		input := []byte(`{"key": "value"}`)
		n, err := w.Write(input)
		assert.Error(t, err)
		assert.Equal(t, "write failed", err.Error())
		assert.Equal(t, 0, n)
	})

	t.Run("ReturnValue", func(t *testing.T) {
		mock := &MockWriter{}
		w := &RedactingWriter{w: mock}

		// Input that will be redacted (so output length != input length)
		input := []byte(`{"password": "secret"}`)

		// Expected output: {"password":"[REDACTED]"}
		// len(`{"password": "secret"}`) = 22
		// len(`{"password":"[REDACTED]"}`) = 25

		n, err := w.Write(input)
		require.NoError(t, err)

		// The writer should return the number of bytes from input, not output
		assert.Equal(t, len(input), n)

		// Verify output length is different
		assert.NotEqual(t, len(input), len(mock.written))
	})
}
