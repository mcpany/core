// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/validation"
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
	// Mock validation
	origIsAllowedPath := validation.IsAllowedPath
	validation.IsAllowedPath = func(path string) error { return nil }
	defer func() { validation.IsAllowedPath = origIsAllowedPath }()

	tmpFile, err := os.CreateTemp("", "audit_test.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	config := &configv1.AuditConfig{
		Enabled:     ptr(true),
		StorageType: ptr(configv1.AuditConfig_STORAGE_TYPE_FILE),
		OutputPath:  ptr(tmpFile.Name()),
	}

	h := NewAuditHandler(slog.NewTextHandler(io.Discard, nil), config)
	assert.NotNil(t, h.store)

	// Verify we can write to the store via Export (which checks for nil store)
	ctx := context.Background()
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test init", 0)
	err = h.Export(ctx, r)
	assert.NoError(t, err)

	// Verify file content
	content, err := os.ReadFile(tmpFile.Name())
	assert.NoError(t, err)
	assert.Contains(t, string(content), "log:test init")
}

func TestAuditHandler_WithMethods(t *testing.T) {
	mock := &mockStore{}
	h := &AuditHandler{
		next:  slog.NewJSONHandler(io.Discard, nil),
		store: mock,
	}

	// Test WithAttrs
	h2 := h.WithAttrs([]slog.Attr{slog.String("foo", "bar")})
	assert.IsType(t, &AuditHandler{}, h2)
	assert.Equal(t, h.store, h2.(*AuditHandler).store)

	// Test WithGroup
	h3 := h.WithGroup("mygroup")
	assert.IsType(t, &AuditHandler{}, h3)
	assert.Equal(t, h.store, h3.(*AuditHandler).store)
}

func ptr[T any](v T) *T {
	return &v
}
