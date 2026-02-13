// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_Execute_ExplicitParameterLocations(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Header
		assert.Equal(t, "header-value", r.Header.Get("X-Test-Header"))

		// Check Cookie
		cookie, err := r.Cookie("Test-Cookie")
		if err != nil {
			t.Errorf("Cookie 'Test-Cookie' not found: %v", err)
		} else {
			assert.Equal(t, "cookie-value", cookie.Value)
		}

		// Check Query (Explicitly mapped)
		assert.Equal(t, "query-value", r.URL.Query().Get("explicit_query"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})

	callDef := configv1.HttpCallDefinition_builder{
		Method: configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("header_param")}.Build(),
				Location: configv1.ParameterLocation_HEADER.Enum(),
				TargetName: proto.String("X-Test-Header"),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("cookie_param")}.Build(),
				Location: configv1.ParameterLocation_COOKIE.Enum(),
				TargetName: proto.String("Test-Cookie"),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("query_param")}.Build(),
				Location: configv1.ParameterLocation_QUERY.Enum(),
				TargetName: proto.String("explicit_query"),
			}.Build(),
		},
	}.Build()

	httpTool, server := setupHTTPToolTest(t, handler, callDef)
	defer server.Close()

	inputs := json.RawMessage(`{
		"header_param": "header-value",
		"cookie_param": "cookie-value",
		"query_param": "query-value"
	}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}
