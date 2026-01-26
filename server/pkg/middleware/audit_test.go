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

type MockStore struct {
	WriteFunc func(ctx context.Context, entry audit.Entry) error
	ReadFunc  func(ctx context.Context, filter audit.Filter) ([]audit.Entry, error)
	CloseFunc func() error
}

func (m *MockStore) Write(ctx context.Context, entry audit.Entry) error {
	if m.WriteFunc != nil {
		return m.WriteFunc(ctx, entry)
	}
	return nil
}

func (m *MockStore) Read(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
	if m.ReadFunc != nil {
		return m.ReadFunc(ctx, filter)
	}
	return nil, nil
}

func (m *MockStore) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

func TestNewAuditMiddleware(t *testing.T) {
	t.Run("Enabled", func(t *testing.T) {
		tmpDir := t.TempDir()
		absTmpDir, err := filepath.Abs(tmpDir)
		require.NoError(t, err)
		validation.SetAllowedPaths([]string{absTmpDir})
		t.Cleanup(func() { validation.SetAllowedPaths(nil) })

		logPath := filepath.Join(absTmpDir, "test_audit.log")

		st := configv1.AuditConfig_STORAGE_TYPE_FILE
		config := configv1.AuditConfig_builder{
			Enabled:     proto.Bool(true),
			StorageType: &st,
			OutputPath:  proto.String(logPath),
		}.Build()
		m, err := NewAuditMiddleware(config)
		require.NoError(t, err)
		assert.NotNil(t, m)
		m.Close()
	})

	t.Run("Disabled", func(t *testing.T) {
		config := configv1.AuditConfig_builder{
			Enabled: proto.Bool(false),
		}.Build()
		m, err := NewAuditMiddleware(config)
		require.NoError(t, err)
		assert.NotNil(t, m)
		m.Close()
	})
}

func TestAuditMiddleware_UpdateConfig(t *testing.T) {
	tmpDir := t.TempDir()
	absTmpDir, err := filepath.Abs(tmpDir)
	require.NoError(t, err)
	validation.SetAllowedPaths([]string{absTmpDir})
	t.Cleanup(func() { validation.SetAllowedPaths(nil) })

	logPath1 := filepath.Join(absTmpDir, "test_audit_1.log")
	logPath2 := filepath.Join(absTmpDir, "test_audit_2.log")

	st := configv1.AuditConfig_STORAGE_TYPE_FILE
	config := configv1.AuditConfig_builder{
		Enabled:     proto.Bool(true),
		StorageType: &st,
		OutputPath:  proto.String(logPath1),
	}.Build()
	m, err := NewAuditMiddleware(config)
	require.NoError(t, err)
	defer m.Close()

	// Mock the store so we can verify Close is called
	closeCalled := false
	mockStore := &MockStore{
		CloseFunc: func() error {
			closeCalled = true
			return nil
		},
	}
	m.SetStore(mockStore)

	// Update config to something different
	newConfig := configv1.AuditConfig_builder{
		Enabled:     proto.Bool(true),
		StorageType: &st,
		OutputPath:  proto.String(logPath2),
	}.Build()

	err = m.UpdateConfig(newConfig)
	require.NoError(t, err)
	assert.True(t, closeCalled, "Old store should be closed")

	// Update with nil config
	err = m.UpdateConfig(nil)
	require.NoError(t, err)
	assert.Nil(t, m.config)
}

func TestAuditMiddleware_Execute(t *testing.T) {
	config := configv1.AuditConfig_builder{
		Enabled:      proto.Bool(true),
		LogArguments: proto.Bool(true),
		LogResults:   proto.Bool(true),
	}.Build()
	m, err := NewAuditMiddleware(config)
	require.NoError(t, err)
	defer m.Close()

	var capturedEntry audit.Entry
	mockStore := &MockStore{
		WriteFunc: func(ctx context.Context, entry audit.Entry) error {
			capturedEntry = entry
			return nil
		},
	}
	m.SetStore(mockStore)

	inputs := map[string]interface{}{
		"key": "value",
	}
	inputBytes, _ := json.Marshal(inputs)

	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: json.RawMessage(inputBytes),
	}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		// Simulate a slight delay
		time.Sleep(1 * time.Millisecond)
		return map[string]interface{}{"result": "ok"}, nil
	}

	// Add user and profile to context
	ctx := context.Background()
	ctx = auth.ContextWithUser(ctx, "user1")
	ctx = auth.ContextWithProfileID(ctx, "profile1")

	res, err := m.Execute(ctx, req, next)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"result": "ok"}, res)

	assert.Equal(t, "test_tool", capturedEntry.ToolName)
	assert.Equal(t, "user1", capturedEntry.UserID)
	assert.Equal(t, "profile1", capturedEntry.ProfileID)
	assert.JSONEq(t, `{"key":"value"}`, string(capturedEntry.Arguments))
	assert.Equal(t, map[string]interface{}{"result": "ok"}, capturedEntry.Result)
	assert.Empty(t, capturedEntry.Error)
	assert.True(t, capturedEntry.DurationMs >= 0)
}

func TestAuditMiddleware_Execute_Error(t *testing.T) {
	config := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()
	m, err := NewAuditMiddleware(config)
	require.NoError(t, err)
	defer m.Close()

	var capturedEntry audit.Entry
	mockStore := &MockStore{
		WriteFunc: func(ctx context.Context, entry audit.Entry) error {
			capturedEntry = entry
			return nil
		},
	}
	m.SetStore(mockStore)

	req := &tool.ExecutionRequest{ToolName: "test_tool"}
	expectedErr := errors.New("execution failed")
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return nil, expectedErr
	}

	_, err = m.Execute(context.Background(), req, next)
	assert.Equal(t, expectedErr, err)

	assert.Equal(t, "execution failed", capturedEntry.Error)
}

func TestAuditMiddleware_Read(t *testing.T) {
	config := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()
	m, err := NewAuditMiddleware(config)
	require.NoError(t, err)
	defer m.Close()

	mockStore := &MockStore{
		ReadFunc: func(ctx context.Context, filter audit.Filter) ([]audit.Entry, error) {
			return []audit.Entry{{ToolName: "found"}}, nil
		},
	}
	m.SetStore(mockStore)

	entries, err := m.Read(context.Background(), audit.Filter{})
	require.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "found", entries[0].ToolName)
}

func TestAuditMiddleware_Disabled(t *testing.T) {
	config := configv1.AuditConfig_builder{
		Enabled: proto.Bool(false),
	}.Build()
	m, err := NewAuditMiddleware(config)
	require.NoError(t, err)
	defer m.Close()

	mockStore := &MockStore{
		WriteFunc: func(ctx context.Context, entry audit.Entry) error {
			t.Fatal("Should not write when disabled")
			return nil
		},
	}
	m.SetStore(mockStore)

	req := &tool.ExecutionRequest{ToolName: "test_tool"}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}

	_, err = m.Execute(context.Background(), req, next)
	require.NoError(t, err)
}

func TestAuditMiddleware_Read_Uninitialized(t *testing.T) {
	// Create middleware but set store to nil
	m := &AuditMiddleware{}

	_, err := m.Read(context.Background(), audit.Filter{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audit store not initialized")
}

func TestAuditMiddleware_WriteLog_Error(t *testing.T) {
	config := configv1.AuditConfig_builder{
		Enabled: proto.Bool(true),
	}.Build()
	m, err := NewAuditMiddleware(config)
	require.NoError(t, err)
	defer m.Close()

	mockStore := &MockStore{
		WriteFunc: func(ctx context.Context, entry audit.Entry) error {
			return errors.New("write failed")
		},
	}
	m.SetStore(mockStore)

	req := &tool.ExecutionRequest{ToolName: "test_tool"}
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "ok", nil
	}

	// This just ensures we hit the error path, though we don't verify the log output
	_, err = m.Execute(context.Background(), req, next)
	require.NoError(t, err)
}
