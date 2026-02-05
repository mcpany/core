// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"net/http"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTool_Execute_ErrorMessagesWithBody(t *testing.T) {
	// We cannot use t.Parallel() because we are modifying environment variables
	// t.Parallel()

	t.Run("400_bad_request_with_json_body", func(t *testing.T) {
		t.Setenv("MCPANY_DEBUG", "true")
		errorBody := `{"error": "Missing parameter 'q'"}`
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(errorBody))
		})
		httpTool, server := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
		defer server.Close()

		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		// This assertion is expected to FAIL currently, as the body is not included
		assert.Contains(t, err.Error(), errorBody, "Error message should contain the response body")
		assert.Contains(t, err.Error(), "400", "Error message should contain status code")
	})

	t.Run("500_internal_error_with_text_body", func(t *testing.T) {
		t.Setenv("MCPANY_DEBUG", "true")
		errorBody := "Database connection failed"
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(errorBody))
		})
		httpTool, server := setupHTTPToolTest(t, handler, &configv1.HttpCallDefinition{})
		defer server.Close()

		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)

		// The body masking depends on the MCPANY_DEBUG environment variable
		if os.Getenv("MCPANY_DEBUG") == "true" {
			assert.Contains(t, err.Error(), errorBody, "Error message should contain the response body in debug mode")
		} else {
			assert.Contains(t, err.Error(), "[Body hidden for security. Enable debug mode to view.]", "Error message should contain the masked body message when not in debug mode")
		}
		assert.Contains(t, err.Error(), "500", "Error message should contain status code")
	})
}
