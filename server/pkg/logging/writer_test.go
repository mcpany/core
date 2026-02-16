// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedactingWriter(t *testing.T) {
	t.Run("RedactsSensitiveJSON", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := &RedactingWriter{w: buf}

		// The input JSON contains keys that should be redacted.
		// We use spaces in input to verify if they are preserved or normalized.
		// Based on previous run, it seems they are preserved or re-formatted with spaces.
		input := []byte(`{"message": "hello", "api_key": "secret", "password": "password123"}`)
		n, err := writer.Write(input)
		require.NoError(t, err)
		assert.Equal(t, len(input), n)

		// Check output
		output := buf.String()
		// We check for the key and the redacted value.
		// Since spacing might vary, we can be a bit more flexible or just check for existence of substring.
		assert.Contains(t, output, `"message": "hello"`)
		assert.Contains(t, output, `"api_key": "[REDACTED]"`)
		assert.Contains(t, output, `"password": "[REDACTED]"`)

		assert.NotContains(t, output, "secret")
		assert.NotContains(t, output, "password123")
	})

	t.Run("PassesThroughNonSensitiveJSON", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := &RedactingWriter{w: buf}

		input := []byte(`{"message": "hello", "user_id": "123"}`)
		n, err := writer.Write(input)
		require.NoError(t, err)
		assert.Equal(t, len(input), n)

		output := buf.String()
		assert.Contains(t, output, `"message": "hello"`)
		assert.Contains(t, output, `"user_id": "123"`)
	})

	t.Run("PassesThroughNonJSON", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := &RedactingWriter{w: buf}

		input := []byte(`some random text log with key=value`)
		n, err := writer.Write(input)
		require.NoError(t, err)
		assert.Equal(t, len(input), n)
		assert.Equal(t, string(input), buf.String())
	})

	t.Run("HandlesEmptyInput", func(t *testing.T) {
		buf := &bytes.Buffer{}
		writer := &RedactingWriter{w: buf}

		input := []byte("")
		n, err := writer.Write(input)
		require.NoError(t, err)
		assert.Equal(t, 0, n)
		assert.Equal(t, "", buf.String())
	})

	t.Run("HandlesWriteError", func(t *testing.T) {
		errWriter := &errorWriter{err: errors.New("write failed")}
		writer := &RedactingWriter{w: errWriter}

		input := []byte(`{"msg":"test"}`)
		n, err := writer.Write(input)
		assert.Error(t, err)
		assert.Equal(t, 0, n) // Should return 0 on error
		assert.Equal(t, "write failed", err.Error())
	})
}

type errorWriter struct {
	err error
}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, w.err
}
