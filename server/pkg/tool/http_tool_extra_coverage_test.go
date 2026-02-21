// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/mcpany/core/server/pkg/tool"
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
