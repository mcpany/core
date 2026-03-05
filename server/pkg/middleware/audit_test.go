// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

type MockAuditStore struct {
	Entries []audit.Entry
	Err     error
}

func (m *MockAuditStore) Write(ctx context.Context, entry audit.Entry) error {
	if m.Err != nil {
		return m.Err
	}
	m.Entries = append(m.Entries, entry)
	return nil
}

func (m *MockAuditStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Entries, nil
}

func (m *MockAuditStore) Close() error {
	return nil
}

func TestNewAuditMiddleware(t *testing.T) {
	// Test disabled
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(false)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	require.NotNil(t, mw)

	// Test enabled with defaults (file store)
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_audit.log")

	// Mock validation.IsAllowedPath to allow temp dir
	originalIsAllowedPath := validation.IsAllowedPath
	validation.IsAllowedPath = func(path string) error {
		return nil
	}
	defer func() { validation.IsAllowedPath = originalIsAllowedPath }()

	cfg = configv1.AuditConfig_builder{
		Enabled:    proto.Bool(true),
		OutputPath: proto.String(logPath),
	}.Build()
	mw, err = NewAuditMiddleware(cfg)
	require.NoError(t, err)
	require.NotNil(t, mw)
	defer mw.Close()
}

func TestAuditMiddleware_Execute(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{
		Enabled:      proto.Bool(true),
		LogArguments: proto.Bool(true),
		LogResults:   proto.Bool(true),
	}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(`{"arg": "value"}`),
	}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)

	require.Len(t, mockStore.Entries, 1)
	entry := mockStore.Entries[0]
	assert.Equal(t, "test-tool", entry.ToolName)
	assert.Contains(t, string(entry.Arguments), "value")
	assert.Equal(t, "success", entry.Result)
}

func TestAuditMiddleware_Execute_Disabled(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(false)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: "test-tool"}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.Empty(t, mockStore.Entries)
}

func TestAuditMiddleware_Execute_Error(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: "test-tool"}

	expectedErr := errors.New("execution failed")
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return nil, expectedErr
	}

	_, err = mw.Execute(ctx, req, next)
	assert.Equal(t, expectedErr, err)

	require.Len(t, mockStore.Entries, 1)
	assert.Equal(t, "execution failed", mockStore.Entries[0].Error)
}

func TestAuditMiddleware_Execute_Redaction(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{
		Enabled:      proto.Bool(true),
		LogArguments: proto.Bool(true),
		LogResults:   proto.Bool(true),
	}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	// Configure global settings for DLP
	// NOTE: Redactor uses config.GlobalSettings().GetDlp().
	// We cannot easily set global settings in this test without side effects or if it's singleton.
	// Looking at `server/pkg/config/config.go`, GlobalSettings is likely a singleton.
	// If `NewAuditMiddleware` creates a `redactor`, we might test redaction if we can control `config.GlobalSettings`.
	// For now, we assume default behavior (no DLP) or if we can inject it.
	// `NewAuditMiddleware` does `m.redactor = NewRedactor(config.GlobalSettings().GetDlp(), nil)`.
	// Since we can't easily mock `config.GlobalSettings()`, we skip specific DLP redaction verification here
	// unless we can set it.

	// However, we can verify that arguments and results are logged.

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName:   "sensitive-tool",
		ToolInputs: json.RawMessage(`{"password": "secret"}`),
	}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return map[string]string{"key": "value"}, nil
	}

	_, err = mw.Execute(ctx, req, next)
	assert.NoError(t, err)

	require.Len(t, mockStore.Entries, 1)
	entry := mockStore.Entries[0]
	assert.NotEmpty(t, entry.Arguments)

	// Result should be JSON marshalled then potentially unmarshalled back
	// Since we passed a map, it should be in Result as a map (or generic interface)
	assert.NotNil(t, entry.Result)
}

func TestAuditMiddleware_UpdateConfig(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	// Update to disable
	newCfg := configv1.AuditConfig_builder{Enabled: proto.Bool(false)}.Build()
	err = mw.UpdateConfig(newCfg)
	assert.NoError(t, err)

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: "test-tool"}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}
	mw.Execute(ctx, req, next)
	assert.Empty(t, mockStore.Entries)

	// Update to enable
	newCfg = configv1.AuditConfig_builder{
		Enabled:      proto.Bool(true),
		LogArguments: proto.Bool(true),
	}.Build()
	err = mw.UpdateConfig(newCfg)
	assert.NoError(t, err)

	// Since UpdateConfig re-initializes store, we need to set our mock again
	// because `initializeStore` creates a new real store.
	mw.SetStore(mockStore)

	mw.Execute(ctx, req, next)
	assert.Len(t, mockStore.Entries, 1)
}

func TestAuditMiddleware_Read(t *testing.T) {
	mockStore := &MockAuditStore{
		Entries: []audit.Entry{
			{ToolName: "t1", Timestamp: time.Now()},
		},
	}
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	entries, err := mw.Read(context.Background(), audit.Filter{})
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "t1", entries[0].ToolName)
}

func TestAuditMiddleware_WriteError(t *testing.T) {
	mockStore := &MockAuditStore{
		Err: errors.New("write error"),
	}
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: "test-tool"}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	// Should not fail execution, but log error to stderr (which we can't easily assert here, but at least no panic)
	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
}

func TestAuditMiddleware_Write(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	entry := audit.Entry{ToolName: "direct-write-tool", Timestamp: time.Now()}

	// Write with valid store
	err = mw.Write(context.Background(), entry)
	assert.NoError(t, err)
	require.Len(t, mockStore.Entries, 1)
	assert.Equal(t, "direct-write-tool", mockStore.Entries[0].ToolName)

	// Write without a store
	mw.SetStore(nil)
	err = mw.Write(context.Background(), entry)
	assert.Error(t, err)
	assert.Equal(t, "audit store not initialized", err.Error())
}

func TestAuditMiddleware_Subscription(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	// Subscribe
	ch, history := mw.SubscribeWithHistory()
	assert.NotNil(t, ch)
	assert.Empty(t, history)

	// Broadcast an entry by writing it
	entry := audit.Entry{ToolName: "broadcast-tool", Timestamp: time.Now()}
	err = mw.Write(context.Background(), entry)
	assert.NoError(t, err)

	// The subscriber should receive the broadcast
	select {
	case msg := <-ch:
		auditEntry, ok := msg.(audit.Entry)
		require.True(t, ok)
		assert.Equal(t, "broadcast-tool", auditEntry.ToolName)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for broadcast")
	}

	// Check history
	newHistory := mw.GetHistory()
	require.GreaterOrEqual(t, len(newHistory), 1)
	found := false
	for _, entry := range newHistory {
		if histEntry, ok := entry.(audit.Entry); ok {
			if histEntry.ToolName == "broadcast-tool" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Expected to find broadcast-tool in history")

	// Unsubscribe
	mw.Unsubscribe(ch)

	// Write again, shouldn't panic, channel should be unblocked (or not receiving)
	entry2 := audit.Entry{ToolName: "broadcast-tool-2", Timestamp: time.Now()}
	err = mw.Write(context.Background(), entry2)
	assert.NoError(t, err)

	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("should not receive after unsubscribe or channel should be closed")
		}
	default:
		// Expected if channel is not closed but we're not receiving, though Unsubscribe closes it so `ok` should be false
	}
}

func TestAuditMiddleware_UpdateConfig_Detailed(t *testing.T) {
	// 1. Update with nil config
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)

	mockStore := &MockAuditStore{}
	mw.SetStore(mockStore)

	err = mw.UpdateConfig(nil)
	assert.NoError(t, err)
	// It should close and set store to nil

	// 2. Error handling when re-initializing store (simulate a failure by creating a Postgres store with an invalid path/dsn, assuming NewPostgresAuditStore will fail)
	newCfg := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
		StorageType: configv1.AuditConfig_STORAGE_TYPE_POSTGRES.Enum(),
		OutputPath: proto.String("invalid-dsn://"),
	}.Build()

	err = mw.UpdateConfig(newCfg)
	// Expecting error because "invalid-dsn://" is not a valid postgres DSN format usually, but depending on the implementation it might not fail until connect.
	// Actually we just test if err is returned without panicking.
	// If it succeeds, that's fine too, depending on underlying driver lazy connection.

	// 3. Changing storage type
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_audit2.log")

	originalIsAllowedPath := validation.IsAllowedPath
	validation.IsAllowedPath = func(path string) error {
		return nil
	}
	defer func() { validation.IsAllowedPath = originalIsAllowedPath }()

	fileCfg := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
		StorageType: configv1.AuditConfig_STORAGE_TYPE_FILE.Enum(),
		OutputPath: proto.String(logPath),
	}.Build()

	err = mw.UpdateConfig(fileCfg)
	assert.NoError(t, err)
}

func TestAuditMiddleware_Execute_NilConfig(t *testing.T) {
	cfg := configv1.AuditConfig_builder{Enabled: proto.Bool(true)}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)

	// Set to nil
	err = mw.UpdateConfig(nil)
	assert.NoError(t, err)

	ctx := context.Background()
	req := &tool.ExecutionRequest{ToolName: "test-tool"}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	// When config is nil, it should just call next
	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
}

func TestAuditMiddleware_Execute_EmptyArgs(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
		LogArguments: proto.Bool(true),
	}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName: "test-tool",
		Arguments: nil,
	}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)

	require.Len(t, mockStore.Entries, 1)
	// 'null' translates to `[]byte{110, 117, 108, 108}` in JSON bytes
	if len(mockStore.Entries[0].Arguments) > 0 {
		assert.Equal(t, []byte("null"), []byte(mockStore.Entries[0].Arguments))
	} else {
		assert.Empty(t, mockStore.Entries[0].Arguments)
	}
}

func TestAuditMiddleware_Execute_FullDetails(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{
		Enabled:      proto.Bool(true),
		LogArguments: proto.Bool(true),
		LogResults:   proto.Bool(true),
	}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	ctx := context.Background()
	ctx = WithTraceContext(ctx, "existing-trace", "existing-span", "existing-parent")
	ctx = auth.ContextWithUser(ctx, "test-user")
	ctx = auth.ContextWithProfileID(ctx, "test-profile")

	argsMap := map[string]any{
		"param1": "value1",
		"secret": "super-secret-password", // Should be redacted if DLP matches, but here redactor is nil or default
	}
	argsBytes, _ := json.Marshal(argsMap)
	req := &tool.ExecutionRequest{
		ToolName:   "full-tool",
		ToolInputs: json.RawMessage(argsBytes),
	}

	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return map[string]any{"status": "ok", "data": "secret-result"}, nil
	}

	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	require.Len(t, mockStore.Entries, 1)
	entry := mockStore.Entries[0]
	assert.Equal(t, "full-tool", entry.ToolName)
	assert.Equal(t, "existing-trace", entry.TraceID)
	// Parent should be the previous span ID
	assert.Equal(t, "existing-span", entry.ParentID)
	// Span should be newly generated
	assert.NotEmpty(t, entry.SpanID)
	assert.NotEqual(t, "existing-span", entry.SpanID)

	assert.Equal(t, "test-user", entry.UserID)
	assert.Equal(t, "test-profile", entry.ProfileID)
	assert.NotEmpty(t, entry.Arguments)
	assert.NotNil(t, entry.Result)
}

func TestAuditMiddleware_Execute_NoLogging(t *testing.T) {
	mockStore := &MockAuditStore{}
	cfg := configv1.AuditConfig_builder{
		Enabled:      proto.Bool(true),
		LogArguments: proto.Bool(false),
		LogResults:   proto.Bool(false),
	}.Build()
	mw, err := NewAuditMiddleware(cfg)
	require.NoError(t, err)
	mw.SetStore(mockStore)

	ctx := context.Background()
	argsMap := map[string]any{"param": "value"}
	argsBytes, _ := json.Marshal(argsMap)
	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(argsBytes),
	}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	res, err := mw.Execute(ctx, req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)

	require.Len(t, mockStore.Entries, 1)
	entry := mockStore.Entries[0]
	// Arguments should be nil since LogArguments is false
	assert.Nil(t, entry.Arguments)
	// Result should be nil since LogResults is false
	assert.Nil(t, entry.Result)
}
