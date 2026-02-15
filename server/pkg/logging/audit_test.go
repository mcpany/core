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

func TestAuditHandler_Handle(t *testing.T) {
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

func (m *mockStore) Write(ctx context.Context, entry audit.Entry) error {
	m.entries = append(m.entries, entry)
	return nil
}
func (m *mockStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	return nil, nil
}
func (m *mockStore) Close() error { return nil }

func TestAuditHandler_Export(t *testing.T) {
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
