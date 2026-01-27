// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"
)

// Re-using mockToolManager from http_test.go is not possible if it's not exported.
// But since we are in package http, we can access it if it is defined in http_test.go which is also package http.
// The file http_test.go defines `type mockToolManager struct`. It is unexported.
// But `go test` compiles all files in the package together (including _test.go files).
// So I should be able to use `newMockToolManager` and `mockToolManager`.

func TestInputSchemaRequiredCorruption(t *testing.T) {
	pm := pool.NewManager()
	// Use the mock defined in http_test.go
	mockTm := newMockToolManager()
	upstream := NewUpstream(pm)

	// JSON config with input_schema having mixed types in required array.
	// JSON Schema "required" must be strings. But let's say the user puts a number.
	// Or maybe the user puts a string that looks like a number? No, actual number.
	// "required": ["foo", 123]

	configJSON := `{
		"name": "bug-repro-service",
		"http_service": {
			"address": "http://127.0.0.1",
			"tools": [{
				"name": "test-op",
				"call_id": "test-op-call"
			}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test",
					"input_schema": {
						"type": "object",
						"properties": {
							"foo": {"type": "string"},
							"bar": {"type": "number"}
						},
						"required": ["foo", 123]
					},
					"parameters": [
						{
							"schema": {
								"name": "baz",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, mockTm, nil, nil, false)
	require.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName

	// We need to retrieve the tool from the mock manager
	registeredTool, ok := mockTm.GetTool(toolID)
	require.True(t, ok, "Tool should be registered")

	// Get the input schema from the tool definition
	toolProto := registeredTool.Tool()
	inputSchema := toolProto.GetAnnotations().GetInputSchema()
	require.NotNil(t, inputSchema)

	requiredVal := inputSchema.Fields["required"]
	require.NotNil(t, requiredVal)
	listVal := requiredVal.GetListValue()
	require.NotNil(t, listVal)

	// Expected: "foo", "123" (if preserved as string?) or 123 (if preserved as number) and "baz".
	// The bug predicts that 123 will become "" because GetStringValue() returns "" for numbers.

	values := listVal.Values
	var stringValues []string
	for _, v := range values {
		// We want to check what is actually in there.
		// If it's a string value, we get it.
		// If it's a number value, we want to see it as number, but if it was converted to empty string...
		if v.GetStringValue() != "" {
			stringValues = append(stringValues, v.GetStringValue())
		} else {
			// Check if it is actually an empty string or something else
			_, ok := v.Kind.(*structpb.Value_StringValue)
			if ok {
				stringValues = append(stringValues, "<EMPTY_STRING>")
			} else {
				stringValues = append(stringValues, "<NON_STRING>")
			}
		}
	}

	// We expect "baz" to be there.
	// We expect "foo" to be there.
	// The bug: 123 becomes <EMPTY_STRING>.

	t.Logf("Required values: %v", stringValues)

	assert.Contains(t, stringValues, "foo")
	assert.Contains(t, stringValues, "baz")

	// If the bug exists, we will see <EMPTY_STRING>.
	// If the bug is fixed/prevented, we might see 123 preserved (as <NON_STRING> or string "123").
	// But since we are creating a NewList from []any which contains strings (from GetStringValue),
	// it will definitely be converted to StringValue.

	assert.NotContains(t, stringValues, "<EMPTY_STRING>", "Non-string values in 'required' were corrupted to empty strings")
}
