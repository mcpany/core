/**
 * Copyright 2026 Author(s) of MCP Any
 * SPDX-License-Identifier: Apache-2.0
 */

package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockReader struct {
	entries []AuditEntry
	err     error
}

func (m *mockReader) Read(ctx context.Context, filter AuditFilter) ([]AuditEntry, error) {
	return m.entries, m.err
}

func TestExportCSV(t *testing.T) {
	now := time.Now()
	entries := []AuditEntry{
		{
			Timestamp:  now,
			ToolName:   "test-tool",
			UserID:     "user1",
			ProfileID:  "profile1",
			DurationMs: 100,
			Arguments:  json.RawMessage(`{"arg":"val"}`),
			Result:     "success",
		},
	}

	reader := &mockReader{entries: entries}
	var buf bytes.Buffer

	err := ExportCSV(context.Background(), reader, AuditFilter{}, &buf)
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Timestamp,Tool Name,User ID,Profile ID,Duration (ms),Error,Arguments,Result")
	assert.Contains(t, output, now.Format(time.RFC3339))
	assert.Contains(t, output, "test-tool")
	assert.Contains(t, output, "user1")
	assert.Contains(t, output, "100")
	// CSV escapes quotes
	assert.True(t, strings.Contains(output, `"{""arg"":""val""}"`) || strings.Contains(output, `{"arg":"val"}`), "Should contain arguments (escaped or not)")
	assert.Contains(t, output, "success")
}
