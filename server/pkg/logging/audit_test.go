// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/audit"
	"github.com/stretchr/testify/assert"
)

// TestAuditHandler_Handle ...
// Summary: TestAuditHandler_Handle
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	var buf bytes.Buffer
	baseHandler := slog.NewTextHandler(&buf, nil)
	auditHandler := NewAuditHandler(baseHandler, nil)

	logger := slog.New(auditHandler)
	logger.Info("test audit message")

	// Verify that the message was passed to the base handler
	if !strings.Contains(buf.String(), "test audit message") {
		t.Errorf("Expected log message to be forwarded, got: %s", buf.String())
	}
}

type mockStore struct {
	entries []audit.Entry
}

// Write ...
// Summary: Write
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m.entries = append(m.entries, entry)
	return nil
}
// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// TestAuditHandler_Export ...
// Summary: TestAuditHandler_Export
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	mock := &mockStore{}
	h := &AuditHandler{
		next:  slog.NewJSONHandler(io.Discard, nil),
		store: mock,
	}

	logger := slog.New(h)

	logger.Info("test message", slog.String("foo", "bar"))

	if len(mock.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(mock.entries))
	}

	entry := mock.entries[0]
	assert.Equal(t, "log:test message", entry.ToolName)
	assert.Contains(t, string(entry.Arguments), "foo")
	assert.Contains(t, string(entry.Arguments), "bar")
}
