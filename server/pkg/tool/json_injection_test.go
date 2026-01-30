// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_JSONInjection_Safety(t *testing.T) {
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	// Sentinel Security Check: Verify JSON Injection in Input Transformer is Prevented

	var receivedBody string

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		require.NoError(t, err)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "POST " + server.URL
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	// Configure a tool with a JSON template
	callDef := configv1.HttpCallDefinition_builder{
		Method: configv1.HttpCallDefinition_HTTP_METHOD_POST.Enum(),
		InputTransformer: configv1.InputTransformer_builder{
			Template: lo.ToPtr(`{"username": "{{username}}"}`),
		}.Build(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("username")}.Build(),
			}.Build(),
		},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Malicious input trying to inject a field
	maliciousInput := `", "admin": true, "dummy": "`
	inputs := map[string]interface{}{
		"username": maliciousInput,
	}
	inputsJSON, _ := json.Marshal(inputs)

	req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(inputsJSON)}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	// If injection is successful, we expect the body to be:
	// {"username": "", "admin": true, "dummy": ""}
	// If injection is prevented (e.g., proper escaping), it should be:
	// {"username": "\", \"admin\": true, \"dummy\": \""}

	t.Logf("Received Body: %s", receivedBody)

	// Check if "admin": true exists as a raw key (not part of a string value)
	var bodyMap map[string]interface{}
	err = json.Unmarshal([]byte(receivedBody), &bodyMap)
	require.NoError(t, err, "Body should be valid JSON")

	if val, ok := bodyMap["admin"]; ok {
		// If 'admin' key exists, it must NOT be the injected boolean true.
		// In fact, it shouldn't be there at all if the structure is preserved.
		assert.Fail(t, "VULNERABILITY CONFIRMED: JSON Injection successful! 'admin' field was injected as a separate key.", "Value: %v", val)
	} else {
		t.Log("Vulnerability Mitigated: 'admin' field not found.")
		// The input should be preserved as the value of 'username'
		assert.Equal(t, maliciousInput, bodyMap["username"], "Username should match input (properly escaped)")
	}
}
