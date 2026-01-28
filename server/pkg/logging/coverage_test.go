// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestAuditHandler_Coverage_Extra(t *testing.T) {
	// Use a local file to pass validation
	tmpFile := filepath.Join(t.TempDir(), "audit_test_coverage.log")

	// Test NewAuditHandler with enabled config
	st := configv1.AuditConfig_STORAGE_TYPE_FILE
	cfg := configv1.AuditConfig_builder{
		Enabled:    proto.Bool(true),
		StorageType: &st,
		OutputPath:  proto.String(tmpFile),
	}.Build()

	baseHandler := slog.NewTextHandler(io.Discard, nil)
	h := NewAuditHandler(baseHandler, cfg)
	// It might still fail if validation is strict or deps missing, so we check h.store is either nil or not, but we ensured the code path ran.
	// But to test WithAttrs on a full handler, we'd like it to be initialized.
	// If initialization failed, h.store is nil. WithAttrs handles that (copying nil interface).

	// Let's assert h is returned.
	assert.NotNil(t, h)

	// Test WithAttrs
	hAttrs := h.WithAttrs([]slog.Attr{slog.String("key", "val")})
	assert.NotNil(t, hAttrs)
	assert.IsType(t, &AuditHandler{}, hAttrs)

	// Test WithGroup
	hGroup := h.WithGroup("group")
	assert.NotNil(t, hGroup)
	assert.IsType(t, &AuditHandler{}, hGroup)

	// Test Enabled delegation
	assert.True(t, h.Enabled(context.Background(), slog.LevelInfo))
}

func TestAuditHandler_InitializeStore_Coverage(t *testing.T) {
	// Test various storage types initialization
	types := []configv1.AuditConfig_StorageType{
		configv1.AuditConfig_STORAGE_TYPE_POSTGRES,
		configv1.AuditConfig_STORAGE_TYPE_SQLITE,
		configv1.AuditConfig_STORAGE_TYPE_WEBHOOK,
		configv1.AuditConfig_STORAGE_TYPE_SPLUNK,
		configv1.AuditConfig_STORAGE_TYPE_DATADOG,
		configv1.AuditConfig_STORAGE_TYPE_UNSPECIFIED,
	}

	for _, st := range types {
		val := st // Copy to take address
		// Use a safe temp path for all types to avoid platform-specific file issues (like /dev/null on Windows)
		safePath := filepath.Join(t.TempDir(), "audit.log")
		cfg := configv1.AuditConfig_builder{
			Enabled:    proto.Bool(true),
			StorageType: &val,
			OutputPath:  proto.String(safePath),
		}.Build()
		NewAuditHandler(slog.NewTextHandler(io.Discard, nil), cfg)
	}
}

// errorWriter implements io.Writer and always returns error
type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, errors.New("write error")
}

func TestRedactingWriter_Error(t *testing.T) {
	w := &RedactingWriter{w: &errorWriter{}}
	n, err := w.Write([]byte("{}"))
	assert.Error(t, err)
	assert.Equal(t, 0, n)
}
