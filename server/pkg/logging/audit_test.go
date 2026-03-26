// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
<<<<<<< HEAD
	"os"
	"github.com/mcpany/core/server/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
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
<<<<<<< HEAD

func TestAuditHandler_Enabled(t *testing.T) {
	nextHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError})
	cfg := &configv1.AuditConfig{}

	handler := &AuditHandler{
		next:   nextHandler,
		config: cfg,
	}

	assert.False(t, handler.Enabled(context.Background(), slog.LevelInfo))
	assert.True(t, handler.Enabled(context.Background(), slog.LevelError))
}

func TestAuditHandler_WithAttrs_WithGroup(t *testing.T) {
	nextHandler := slog.NewTextHandler(os.Stdout, nil)
	cfg := &configv1.AuditConfig{}

	handler := &AuditHandler{
		next:   nextHandler,
		config: cfg,
	}

	attrs := []slog.Attr{slog.String("foo", "bar")}
	newHandler := handler.WithAttrs(attrs)
	assert.NotNil(t, newHandler)
	assert.IsType(t, &AuditHandler{}, newHandler)

	groupHandler := handler.WithGroup("mygroup")
	assert.NotNil(t, groupHandler)
	assert.IsType(t, &AuditHandler{}, groupHandler)
}

func TestNewAuditHandler_InitializeStore(t *testing.T) {
	nextHandler := slog.NewTextHandler(os.Stdout, nil)

	tempDir := t.TempDir()
	validation.SetAllowedPaths([]string{tempDir})
	t.Cleanup(func() {
		validation.SetAllowedPaths(nil)
	})

	// Use FILE to test initialization
	cfg := &configv1.AuditConfig{}
	cfg.SetEnabled(true)
	cfg.SetStorageType(configv1.AuditConfig_STORAGE_TYPE_FILE)
	cfg.SetOutputPath(tempDir + "/test_audit.log")

	handler := NewAuditHandler(nextHandler, cfg)

	assert.NotNil(t, handler)
	assert.NotNil(t, handler.store)

	// Clean up
	if s, ok := handler.store.(interface{ Close() error }); ok {
		_ = s.Close()
	}
}
=======
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
