// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/validation"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTool_Execute_429(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	httpTool, server := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
	defer server.Close()

	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Too Many Requests")
}

func TestHTTPTool_Execute_UnsafeURL_Sequential(t *testing.T) {
	// Restore IsSafeURL just for this test (mocked in TestMain of tool package)
	// We mock it to fail for any URL to simulate unsafe URL.
	// We don't need real logic, we just want to verify Execute handles the error.
	originalIsSafeURL := validation.IsSafeURL
	validation.IsSafeURL = func(urlStr string) error {
		return fmt.Errorf("unsafe url: %s", urlStr)
	}
	defer func() { validation.IsSafeURL = originalIsSafeURL }()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	httpTool, server := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
	defer server.Close()

	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsafe url")
}
