// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"os"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
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

func TestNewAuditHandler_Initialization(t *testing.T) {
	// Create a temp file for audit logs in the current directory
	// We use "audit_test_*.json" which matches common ignore patterns or just clean it up
	tmpfile, err := os.CreateTemp(".", "audit_test_*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	enabled := true
	storageType := configv1.AuditConfig_STORAGE_TYPE_FILE
	outputPath := tmpfile.Name()
	config := &configv1.AuditConfig{
		Enabled:     &enabled,
		StorageType: &storageType,
		OutputPath:  &outputPath,
	}

	h := NewAuditHandler(slog.NewJSONHandler(io.Discard, nil), config)

	assert.NotNil(t, h.store)
	assert.Equal(t, config, h.config)

	// Verify we can write to it via Handle/Export
	ctx := context.Background()
	record := slog.NewRecord(time.Now(), slog.LevelInfo, "init test", 0)
	err = h.Handle(ctx, record)
	assert.NoError(t, err)

	// Close the store (if possible, though AuditHandler doesn't expose Close directly, the store does)
	if store, ok := h.store.(interface{ Close() error }); ok {
		store.Close()
	}
}

func TestAuditHandler_WithMethods(t *testing.T) {
	enabled := false
	config := &configv1.AuditConfig{Enabled: &enabled}
	h := NewAuditHandler(slog.NewJSONHandler(io.Discard, nil), config)

	// Test WithAttrs
	attrs := []slog.Attr{slog.String("key", "value")}
	hAttrs := h.WithAttrs(attrs)

	assert.IsType(t, &AuditHandler{}, hAttrs)
	hAttrsTyped := hAttrs.(*AuditHandler)
	assert.Equal(t, config, hAttrsTyped.config)

	// Test WithGroup
	hGroup := h.WithGroup("group")
	assert.IsType(t, &AuditHandler{}, hGroup)
	hGroupTyped := hGroup.(*AuditHandler)
	assert.Equal(t, config, hGroupTyped.config)
}
