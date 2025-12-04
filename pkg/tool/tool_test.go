/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/command"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/testutil"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/util"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestGRPCTool_Execute_NoPool(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	poolManager := pool.NewManager()
	grpcTool := tool.NewGRPCTool(gt, poolManager, "test", nil, &configv1.GrpcCallDefinition{})
	_, err := grpcTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no grpc pool found for service")
}

func TestHTTPTool_Execute_NoPool(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	poolManager := pool.NewManager()
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no http pool found for service")
}

func TestHTTPTool_Execute_SecretResolutionError(t *testing.T) {
	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "GET http://test.com/{{secret}}",
	}
	parameters := []*configv1.HttpParameterMapping{
		{
			Target: &configv1.HttpParameterMapping_Secret{
				Secret: &configv1.Secret{
					Source: &configv1.Secret_EnvironmentVariable{
						EnvironmentVariable: "TEST_SECRET",
					},
				},
			},
			Schema: &v1.JSONSchema{
				Name: "secret",
			},
		},
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{Parameters: parameters}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve secret for parameter")
}

func TestCommandTool_Execute_SecretResolutionError(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	parameters := []*configv1.CommandLineParameterMapping{
		{
			Target: &configv1.CommandLineParameterMapping_Secret{
				Secret: &configv1.Secret{
					Source: &configv1.Secret_EnvironmentVariable{
						EnvironmentVariable: "TEST_SECRET",
					},
				},
			},
			Schema: &v1.JSONSchema{
				Name: "secret",
			},
		},
	}
	commandTool := tool.NewCommandTool(gt, &configv1.CommandLineUpstreamService{}, &configv1.CommandLineCallDefinition{Parameters: parameters})
	_, err := commandTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve secret for parameter")
}

func TestLocalCommandTool_Execute_SecretResolutionError(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	parameters := []*configv1.CommandLineParameterMapping{
		{
			Target: &configv1.CommandLineParameterMapping_Secret{
				Secret: &configv1.Secret{
					Source: &configv1.Secret_EnvironmentVariable{
						EnvironmentVariable: "TEST_SECRET",
					},
				},
			},
			Schema: &v1.JSONSchema{
				Name: "secret",
			},
		},
	}
	localCommandTool := tool.NewLocalCommandTool(gt, &configv1.CommandLineUpstreamService{}, &configv1.CommandLineCallDefinition{Parameters: parameters})
	_, err := localCommandTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve secret for parameter")
}

func TestHTTPTool_Execute_InvalidMethodAndURL(t *testing.T) {
	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "INVALID",
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http tool definition")
}

func TestLocalCommandTool_Execute_NonStringArgs(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	localCommandTool := tool.NewLocalCommandTool(gt, &configv1.CommandLineUpstreamService{}, &configv1.CommandLineCallDefinition{})
	_, err := localCommandTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"args": [123]}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-string value in 'args' array")
}

func TestLocalCommandTool_Execute_ArgsNotAList(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	localCommandTool := tool.NewLocalCommandTool(gt, &configv1.CommandLineUpstreamService{}, &configv1.CommandLineCallDefinition{})
	_, err := localCommandTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"args": "not a list"}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'args' parameter must be an array of strings")
}

func TestCommandTool_Execute_NonStringArgs(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	commandTool := tool.NewCommandTool(gt, &configv1.CommandLineUpstreamService{}, &configv1.CommandLineCallDefinition{})
	_, err := commandTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"args": [123]}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-string value in 'args' array")
}

func TestCommandTool_Execute_ArgsNotAList(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	commandTool := tool.NewCommandTool(gt, &configv1.CommandLineUpstreamService{}, &configv1.CommandLineCallDefinition{})
	_, err := commandTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"args": "not a list"}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "'args' parameter must be an array of strings")
}

func TestHTTPTool_Execute_InputTransformerTemplateError(t *testing.T) {
	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "POST http://test.com",
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: "{{.invalid",
		},
	}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"key": "value"}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create input template")
}

func TestHTTPTool_Execute_InputTransformerRenderError(t *testing.T) {
	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "POST http://test.com",
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: "{{.key}}",
		},
	}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"otherKey": "value"}`),
	})
	assert.Error(t, err)
}

func TestHTTPTool_Execute_OutputTransformerTemplateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{
				Client: server.Client(),
			}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "GET " + server.URL,
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Template: "{{.invalid",
		},
	}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output template")
}

func TestHTTPTool_Execute_OutputTransformerRenderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{
				Client: server.Client(),
			}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "GET " + server.URL,
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Template: "{{.otherKey}}",
		},
	}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.Error(t, err)
}

func TestMCPTool_Execute_InputTransformerTemplateError(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	mcpTool := tool.NewMCPTool(gt, nil, &configv1.MCPCallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: "{{.invalid",
		},
	})
	_, err := mcpTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"key": "value"}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create input template")
}

func TestMCPTool_Execute_InputTransformerRenderError(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	mcpTool := tool.NewMCPTool(gt, nil, &configv1.MCPCallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: "{{.key}}",
		},
	})
	_, err := mcpTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"otherKey": "value"}`),
	})
	assert.Error(t, err)
}

func TestMCPTool_Execute_OutputTransformerTemplateError(t *testing.T) {
	mockClient := &testutil.MockMCPClient{
		CallToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: `{"key": "value"}`,
					},
				},
			}, nil
		},
	}
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	mcpTool := tool.NewMCPTool(gt, mockClient, &configv1.MCPCallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Template: "{{.invalid",
		},
	})
	_, err := mcpTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output template")
}

func TestMCPTool_Execute_OutputTransformerRenderError(t *testing.T) {
	mockClient := &testutil.MockMCPClient{
		CallToolFunc: func(ctx context.Context, params *mcp.CallToolParams) (*mcp.CallToolResult, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{
						Text: `{"key": "value"}`,
					},
				},
			}, nil
		},
	}
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	mcpTool := tool.NewMCPTool(gt, mockClient, &configv1.MCPCallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Template: "{{.otherKey}}",
		},
	})
	_, err := mcpTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	})
	assert.Error(t, err)
}

func TestOpenAPITool_Execute_InputTransformerTemplateError(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	openapiTool := tool.NewOpenAPITool(gt, nil, nil, http.MethodPost, "http://test.com", nil, &configv1.OpenAPICallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: "{{.invalid",
		},
	})
	_, err := openapiTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"key": "value"}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create input template")
}

func TestOpenAPITool_Execute_InputTransformerRenderError(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	openapiTool := tool.NewOpenAPITool(gt, nil, nil, http.MethodPost, "http://test.com", nil, &configv1.OpenAPICallDefinition{
		InputTransformer: &configv1.InputTransformer{
			Template: "{{.key}}",
		},
	})
	_, err := openapiTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"otherKey": "value"}`),
	})
	assert.Error(t, err)
}

func TestOpenAPITool_Execute_OutputTransformerTemplateError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	openapiTool := tool.NewOpenAPITool(gt, server.Client(), nil, http.MethodGet, server.URL, nil, &configv1.OpenAPICallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Template: "{{.invalid",
		},
	})
	_, err := openapiTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create output template")
}

func TestOpenAPITool_Execute_OutputTransformerRenderError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"key": "value"}`))
	}))
	defer server.Close()

	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	openapiTool := tool.NewOpenAPITool(gt, server.Client(), nil, http.MethodGet, server.URL, nil, &configv1.OpenAPICallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Template: "{{.otherKey}}",
		},
	})
	_, err := openapiTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{}`),
	})
	assert.Error(t, err)
}

func TestContextWithTool(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	grpcTool := tool.NewGRPCTool(gt, nil, "test", nil, &configv1.GrpcCallDefinition{})
	ctx := tool.NewContextWithTool(context.Background(), grpcTool)
	retrievedTool, ok := tool.GetFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, grpcTool, retrievedTool)
}

func TestGetFromContext_NoTool(t *testing.T) {
	_, ok := tool.GetFromContext(context.Background())
	assert.False(t, ok)
}

func TestHTTPTool_Execute_RawBytesOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("raw data"))
	}))
	defer server.Close()

	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{
				Client: server.Client(),
			}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "GET " + server.URL,
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{
		OutputTransformer: &configv1.OutputTransformer{
			Format: configv1.OutputTransformer_RAW_BYTES,
		},
	}, nil)
	result, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.NoError(t, err)
	assert.Equal(t, map[string]any{"raw": []byte("raw data")}, result)
}

func TestHTTPTool_Execute_DeleteRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		assert.Equal(t, "value", r.URL.Query().Get("key"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	httpPool, _ := pool.New[*testutil.MockHttpClientWrapper](
		func(ctx context.Context) (*testutil.MockHttpClientWrapper, error) {
			return &testutil.MockHttpClientWrapper{
				Client: server.Client(),
			}, nil
		}, 1, 1, 0, false)
	poolManager := pool.NewManager()
	poolManager.Register("test", httpPool)

	gt := &v1.Tool{
		Name:                "test",
		ServiceId:           "test",
		UnderlyingMethodFqn: "DELETE " + server.URL,
	}
	httpTool := tool.NewHTTPTool(gt, poolManager, "test", nil, &configv1.HttpCallDefinition{}, nil)
	_, err := httpTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolInputs: json.RawMessage(`{"key": "value"}`),
	})
	assert.NoError(t, err)
}

func TestLocalCommandTool_Execute_Timeout(t *testing.T) {
	gt := &v1.Tool{
		Name:      "test",
		ServiceId: "test",
	}
	localCommandTool := tool.NewLocalCommandTool(gt, &configv1.CommandLineUpstreamService{
		Command: "sleep",
		Timeout: durationpb.New(100 * time.Millisecond),
	}, &configv1.CommandLineCallDefinition{
		Args: []string{"1"},
	})
	result, err := localCommandTool.Execute(context.Background(), &tool.ExecutionRequest{})
	assert.NoError(t, err)
	resultMap, ok := result.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "timeout", resultMap["status"])
}

func TestToolManager_AddTool(t *testing.T) {
	tm := tool.NewToolManager(nil)
	err := tm.AddTool(&testutil.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      "test",
				ServiceId: "test",
			}
		},
	})
	assert.NoError(t, err)
	assert.Len(t, tm.ListTools(), 1)
}

func TestToolManager_AddTool_EmptyServiceID(t *testing.T) {
	tm := tool.NewToolManager(nil)
	err := tm.AddTool(&testutil.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name: "test",
			}
		},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tool service ID cannot be empty")
}

func TestToolManager_AddTool_SanitizeError(t *testing.T) {
	tm := tool.NewToolManager(nil)
	err := tm.AddTool(&testutil.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      "test tool",
				ServiceId: "test",
			}
		},
	})
	assert.Error(t, err)
}

func TestToolManager_ExecuteTool_NotFound(t *testing.T) {
	tm := tool.NewToolManager(nil)
	_, err := tm.ExecuteTool(context.Background(), &tool.ExecutionRequest{
		ToolName: "test.test",
	})
	assert.Error(t, err)
	assert.Equal(t, tool.ErrToolNotFound, err)
}

func TestToolManager_ClearToolsForService(t *testing.T) {
	tm := tool.NewToolManager(nil)
	err := tm.AddTool(&testutil.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      "test",
				ServiceId: "test",
			}
		},
	})
	assert.NoError(t, err)
	err = tm.AddTool(&testutil.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      "test2",
				ServiceId: "test",
			}
		},
	})
	assert.NoError(t, err)
	err = tm.AddTool(&testutil.MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Name:      "test",
				ServiceId: "other",
			}
		},
	})
	assert.NoError(t, err)
	assert.Len(t, tm.ListTools(), 3)
	tm.ClearToolsForService("test")
	assert.Len(t, tm.ListTools(), 1)
}

func TestConvertMCPToolToProto_Nil(t *testing.T) {
	_, err := tool.ConvertMCPToolToProto(nil)
	assert.Error(t, err)
}

func TestConvertMCPToolToProto_DefaultDisplayName(t *testing.T) {
	mcpTool := &mcp.Tool{
		Name: "test",
	}
	pbTool, err := tool.ConvertMCPToolToProto(mcpTool)
	assert.NoError(t, err)
	assert.Equal(t, "test", pbTool.GetDisplayName())
}

func TestConvertProtoToMCPTool_Nil(t *testing.T) {
	_, err := tool.ConvertProtoToMCPTool(nil)
	assert.Error(t, err)
}

func TestConvertProtoToMCPTool_EmptyName(t *testing.T) {
	_, err := tool.ConvertProtoToMCPTool(&v1.Tool{})
	assert.Error(t, err)
}
