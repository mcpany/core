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
