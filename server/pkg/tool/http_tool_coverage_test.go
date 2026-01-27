// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_Execute_Coverage_Secrets(t *testing.T) {
	t.Parallel()

	// Test secret resolution error
	t.Run("secret_resolution_error", func(t *testing.T) {
		poolManager := pool.NewManager()
		// No need for real backend as it should fail before request
		p, errPool := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
			return &client.HTTPClientWrapper{}, nil
		}, 1, 1, 1, 0, true)
		require.NoError(t, errPool)
		poolManager.Register("test-service", p)

		methodAndURL := "GET http://example.com"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		// Parameter backed by missing env var
		paramMapping := configv1.HttpParameterMapping_builder{
			Schema: configv1.ParameterSchema_builder{
				Name: proto.String("token"),
			}.Build(),
			Secret: configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("NON_EXISTENT_ENV_VAR_FOR_TESTING_12345"),
			}.Build(),
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			Parameters: []*configv1.HttpParameterMapping{paramMapping},
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

		inputs := json.RawMessage(`{}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := httpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to resolve secret")
	})
}

func TestHTTPTool_Execute_Coverage_PathTraversal(t *testing.T) {
	t.Parallel()

	poolManager := pool.NewManager()
	p, errPool := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, errPool)
	poolManager.Register("test-service", p)

	methodAndURL := "GET http://example.com/{{file}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("file"),
		}.Build(),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Test encoded path traversal: %2e%2e -> ..
	inputs := json.RawMessage(`{"file": "%2e%2e"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err := httpTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path traversal attempt detected")
}

func TestHTTPTool_Execute_Coverage_CallPolicy(t *testing.T) {
	t.Parallel()

	poolManager := pool.NewManager()
	p, errPool := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, errPool)
	poolManager.Register("test-service", p)

	methodAndURL := "GET http://example.com"
	mcpTool := v1.Tool_builder{
		Name:                proto.String("test-tool"),
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	// Policy that denies the tool by name
	policy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_ALLOW.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				NameRegex: proto.String("test-tool"),
				Action:    configv1.CallPolicy_DENY.Enum(),
			}.Build(),
		},
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{}.Build()

	// NewHTTPTool accepts policies
	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, []*configv1.CallPolicy{policy}, "")

	req := &tool.ExecutionRequest{
		ToolName:   "test-tool",
		ToolInputs: json.RawMessage(`{}`),
	}

	_, err := httpTool.Execute(context.Background(), req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tool execution blocked by policy")
}

func TestHTTPTool_Execute_Coverage_DryRun(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("test-service", p)

	methodAndURL := "POST " + server.URL
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"key": "value"}`)
	req := &tool.ExecutionRequest{
		ToolInputs: inputs,
		DryRun:     true,
		ToolName:   "test-tool",
	}

	result, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.True(t, resultMap["dry_run"].(bool))
	reqMap := resultMap["request"].(map[string]any)
	assert.Equal(t, "POST", reqMap["method"])

	// Check body
	body, ok := reqMap["body"].(string)
	if ok {
		assert.Contains(t, body, "key")
	}
}

func TestNewHTTPTool_Coverage_Errors(t *testing.T) {
	t.Parallel()

	poolManager := pool.NewManager()

	t.Run("invalid_input_template", func(t *testing.T) {
		methodAndURL := "GET http://example.com"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Template: proto.String("{{invalid"),
			}.Build(),
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse input template")
	})

	t.Run("invalid_output_template", func(t *testing.T) {
		methodAndURL := "GET http://example.com"
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		callDef := configv1.HttpCallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Template: proto.String("{{invalid"),
			}.Build(),
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse output template")
	})

	t.Run("invalid_url_parsing", func(t *testing.T) {
		methodAndURL := "GET http://[::1]:namedport" // Invalid URL
		mcpTool := v1.Tool_builder{
			UnderlyingMethodFqn: &methodAndURL,
		}.Build()

		httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")
		_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse url")
	})
}

func TestCommandTool_Coverage_Security(t *testing.T) {
	t.Parallel()

	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ls"),
	}.Build()

	t.Run("argument_injection", func(t *testing.T) {
		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"{{arg}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build(),
				}.Build(),
			},
		}.Build()

		mcpTool := v1.Tool_builder{Name: proto.String("ls")}.Build()
		cmdTool := tool.NewLocalCommandTool(mcpTool, service, callDef, nil, "")

		inputs := json.RawMessage(`{"arg": "-l"}`) // Starts with -
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "argument injection detected")
	})

	t.Run("shell_injection_semicolon", func(t *testing.T) {
		// Shell command
		shService := configv1.CommandLineUpstreamService_builder{
			Command: proto.String("bash"),
		}.Build()

		callDef := configv1.CommandLineCallDefinition_builder{
			Args: []string{"-c", "echo {{msg}}"},
			Parameters: []*configv1.CommandLineParameterMapping{
				configv1.CommandLineParameterMapping_builder{
					Schema: configv1.ParameterSchema_builder{Name: proto.String("msg")}.Build(),
				}.Build(),
			},
		}.Build()

		mcpTool := v1.Tool_builder{Name: proto.String("bash")}.Build()
		cmdTool := tool.NewLocalCommandTool(mcpTool, shService, callDef, nil, "")

		inputs := json.RawMessage(`{"msg": "hello; rm -rf /"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}
		_, err := cmdTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "shell injection detected")
	})
}

func TestMCPTool_Coverage_Errors(t *testing.T) {
	t.Parallel()

	mcpClient := &mockMCPClient{
		callToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{}, nil
		},
	}

	mcpToolDef := v1.Tool_builder{Name: proto.String("mcp-tool")}.Build()

	t.Run("invalid_inputs_json", func(t *testing.T) {
		mcpTool := tool.NewMCPTool(mcpToolDef, mcpClient, &configv1.MCPCallDefinition{})
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)}
		_, err := mcpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})

	t.Run("input_template_error", func(t *testing.T) {
		callDef := configv1.MCPCallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Template: proto.String("{{invalid"),
			}.Build(),
		}.Build()
		mcpTool := tool.NewMCPTool(mcpToolDef, mcpClient, callDef)
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := mcpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse input template")
	})

	t.Run("output_template_error", func(t *testing.T) {
		callDef := configv1.MCPCallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Template: proto.String("{{invalid"),
			}.Build(),
		}.Build()
		mcpTool := tool.NewMCPTool(mcpToolDef, mcpClient, callDef)
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := mcpTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse output template")
	})
}

func TestOpenAPITool_Coverage_Errors(t *testing.T) {
	mockClient := &mockHTTPClient{
		doFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
		},
	}

	toolProto := v1.Tool_builder{Name: proto.String("openapi-tool")}.Build()

	t.Run("invalid_inputs_json", func(t *testing.T) {
		openapiTool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", "http://example.com", nil, &configv1.OpenAPICallDefinition{})
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{invalid`)}
		_, err := openapiTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
	})

	// t.Run("unsafe_url", func(t *testing.T) {
	// 	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "false")
	// 	// Use 0.0.0.0 which is blocked by IsUnspecified and less likely to be allowed by private/loopback flags
	// 	openapiTool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", "http://0.0.0.0:80", nil, &configv1.OpenAPICallDefinition{})
	// 	req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
	// 	_, err := openapiTool.Execute(context.Background(), req)
	// 	require.Error(t, err)
	// 	assert.Contains(t, err.Error(), "unsafe url")
	// })

	t.Run("client_error", func(t *testing.T) {
		errClient := &mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return nil, assert.AnError
			},
		}
		openapiTool := tool.NewOpenAPITool(toolProto, errClient, nil, "GET", "http://example.com", nil, &configv1.OpenAPICallDefinition{})
		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := openapiTool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute http request")
	})
}

func TestConverters_Coverage(t *testing.T) {
	t.Parallel()

	t.Run("ConvertMCPToolToProto_Nil", func(t *testing.T) {
		_, err := tool.ConvertMCPToolToProto(nil)
		require.Error(t, err)
	})

	t.Run("ConvertMCPToolToProto_Full", func(t *testing.T) {
		mcpTool := &mcp.Tool{
			Name:        "test",
			Description: "desc",
			Title:       "Title",
			Annotations: &mcp.ToolAnnotations{
				Title:           "Title2",
				ReadOnlyHint:    true,
				DestructiveHint: proto.Bool(true),
				OpenWorldHint:   proto.Bool(false),
			},
			InputSchema: map[string]interface{}{"type": "object"},
		}
		pbTool, err := tool.ConvertMCPToolToProto(mcpTool)
		require.NoError(t, err)
		assert.Equal(t, "test", pbTool.GetName())
		assert.True(t, pbTool.GetAnnotations().GetReadOnlyHint())
		assert.True(t, pbTool.GetAnnotations().GetDestructiveHint())
	})

	t.Run("ConvertProtoToMCPTool_Nil", func(t *testing.T) {
		_, err := tool.ConvertProtoToMCPTool(nil)
		require.Error(t, err)
	})

	t.Run("ConvertProtoToMCPTool_EmptyName", func(t *testing.T) {
		pbTool := &v1.Tool{}
		_, err := tool.ConvertProtoToMCPTool(pbTool)
		require.Error(t, err)
	})
}
