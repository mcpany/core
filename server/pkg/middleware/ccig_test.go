// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCCIGMiddleware_Execute_Pass(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ccig_test_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte("valid cache data")
	err = os.WriteFile(tmpFile.Name(), content, 0644)
	require.NoError(t, err)

	hash := fmt.Sprintf("%x", sha256.Sum256(content))

	config := CCIGConfig{
		Enabled:     true,
		TargetTools: []string{"cache:restore"},
		AttestedHashes: map[string]string{
			tmpFile.Name(): hash,
		},
		ArgumentName: "path",
	}

	middleware := NewCCIGMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "cache:restore",
		Arguments: map[string]interface{}{
			"path": tmpFile.Name() + "/../" + filepath.Base(tmpFile.Name()), // Test normalization bypass
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestCCIGMiddleware_Execute_FailHashMismatch(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "ccig_test_*")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	content := []byte("poisoned cache data")
	err = os.WriteFile(tmpFile.Name(), content, 0644)
	require.NoError(t, err)

	config := CCIGConfig{
		Enabled:     true,
		TargetTools: []string{"cache:restore"},
		AttestedHashes: map[string]string{
			tmpFile.Name(): "expected_hash_that_does_not_match",
		},
		ArgumentName: "path",
	}

	middleware := NewCCIGMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "cache:restore",
		Arguments: map[string]interface{}{
			"path": tmpFile.Name(),
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Supply Chain Poisoning suspected")
	assert.Nil(t, res)
	assert.False(t, nextCalled)
}

func TestCCIGMiddleware_Execute_FileMissing(t *testing.T) {
	config := CCIGConfig{
		Enabled:     true,
		TargetTools: []string{"cache:restore"},
		AttestedHashes: map[string]string{
			"/tmp/does_not_exist_ccig": "some_hash",
		},
		ArgumentName: "path",
	}

	middleware := NewCCIGMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "cache:restore",
		Arguments: map[string]interface{}{
			"path": "/tmp/does_not_exist_ccig",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unable to read file")
	assert.Nil(t, res)
	assert.False(t, nextCalled)
}

func TestCCIGMiddleware_Execute_NotTargetTool(t *testing.T) {
	config := CCIGConfig{
		Enabled:     true,
		TargetTools: []string{"cache:restore"},
		AttestedHashes: map[string]string{
			"/some/path": "hash",
		},
		ArgumentName: "path",
	}

	middleware := NewCCIGMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "other_tool",
		Arguments: map[string]interface{}{
			"path": "/some/path",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestCCIGMiddleware_Execute_Disabled(t *testing.T) {
	config := CCIGConfig{
		Enabled:     false,
		TargetTools: []string{"cache:restore"},
		AttestedHashes: map[string]string{
			"/some/path": "hash",
		},
		ArgumentName: "path",
	}

	middleware := NewCCIGMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "cache:restore",
		Arguments: map[string]interface{}{
			"path": "/some/path",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}

func TestCCIGMiddleware_Execute_MissingArgument(t *testing.T) {
	config := CCIGConfig{
		Enabled:     true,
		TargetTools: []string{"cache:restore"},
		AttestedHashes: map[string]string{
			"/some/path": "hash",
		},
		ArgumentName: "filepath",
	}

	middleware := NewCCIGMiddleware(config)

	req := &tool.ExecutionRequest{
		ToolName: "cache:restore",
		Arguments: map[string]interface{}{
			"wrong_arg": "/some/path",
		},
	}

	nextCalled := false
	next := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		nextCalled = true
		return "success", nil
	}

	res, err := middleware.Execute(context.Background(), req, next)
	assert.NoError(t, err)
	assert.Equal(t, "success", res)
	assert.True(t, nextCalled)
}
