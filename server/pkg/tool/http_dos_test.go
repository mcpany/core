// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"crypto/rand"
	"net/http"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTool_Execute_LargeResponse(t *testing.T) {
	// Not parallel because we modify env vars
	// t.Parallel()

	// Enable loopback for testing
	os.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	defer os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")

	// 15MB response
	const responseSize = 15 * 1024 * 1024

	// Create a handler that sends a large response
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)

		// Write random data to avoid compression or optimization
		// We write in chunks to be nice to memory
		chunk := make([]byte, 1024*1024) // 1MB chunk
		for i := 0; i < 15; i++ {
			if _, err := rand.Read(chunk); err != nil {
				return
			}
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
	})

	format := configv1.OutputTransformer_RAW_BYTES
	callDef := configv1.HttpCallDefinition_builder{
		OutputTransformer: configv1.OutputTransformer_builder{
			Format: &format,
		}.Build(),
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	req := &tool.ExecutionRequest{}

	// Execute the tool
	_, err := httpTool.Execute(context.Background(), req)

	// Expect error due to response size limit
	require.Error(t, err, "Should fail due to response size limit")
	assert.Contains(t, err.Error(), "response body exceeds maximum size", "Error should mention size limit")
	t.Logf("Got expected error: %v", err)
}
