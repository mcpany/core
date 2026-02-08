// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock HTTP Client for tool package tests
type mockHTTPClient struct {
	client.HTTPClient
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return nil, errors.New("not implemented")
}

func TestLocalCommandTool_SecurityChecks(t *testing.T) {
	// Test CommandTool security checks directly via Execute or helper methods if exposed (they are not exported)
	// We use Execute to trigger them.

	svcConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ls"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"{{arg}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("arg")}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{Name: proto.String("cmd_tool")}.Build()

	cmdTool := NewLocalCommandTool(toolDef, svcConfig, callDef, nil, "call1")

	ctx := context.Background()

	tests := []struct {
		name      string
		inputs    string
		wantErr   string
	}{
		{
			name:    "Path Traversal ..",
			inputs:  `{"arg": ".."}`,
			wantErr: "path traversal attempt detected",
		},
		{
			name:    "Path Traversal ../",
			inputs:  `{"arg": "../foo"}`,
			wantErr: "path traversal attempt detected",
		},
		{
			name:    "Path Traversal /..",
			inputs:  `{"arg": "foo/.."}`,
			wantErr: "path traversal attempt detected",
		},
		{
			name:    "Path Traversal encoded",
			inputs:  `{"arg": "%2e%2e"}`,
			wantErr: "path traversal attempt detected",
		},
		{
			name:    "Absolute Path",
			inputs:  `{"arg": "/etc/passwd"}`,
			wantErr: "absolute path detected",
		},
		{
			name:    "Argument Injection",
			inputs:  `{"arg": "-rf"}`,
			wantErr: "argument injection detected",
		},
		{
			name:    "Argument Injection Negative Number Allowed",
			inputs:  `{"arg": "-1"}`,
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName:   "cmd_tool",
				ToolInputs: []byte(tt.inputs),
				DryRun:     true, // Use DryRun to avoid actual execution but trigger checks
			}
			_, err := cmdTool.Execute(ctx, req)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLocalCommandTool_ShellInjection(t *testing.T) {
	// Test shell injection checks when command is a shell
	svcConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("bash"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"-c", "{{script}}"},
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("script")}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{Name: proto.String("shell_tool")}.Build()

	cmdTool := NewLocalCommandTool(toolDef, svcConfig, callDef, nil, "call1")

	ctx := context.Background()

	tests := []struct {
		name      string
		inputs    string
		wantErr   string
	}{
		{
			name:    "Shell Injection ;",
			inputs:  `{"script": "echo hi; rm -rf /"}`,
			wantErr: "security risk: template substitution is not allowed", // Strict block
		},
		{
			name:    "Shell Injection backtick",
			inputs:  `{"script": "echo ` + "`whoami`" + `"}`,
			wantErr: "security risk: template substitution is not allowed", // Strict block
		},
		{
			name:    "Safe Shell Command",
			inputs:  `{"script": "whoami"}`,
			wantErr: "security risk: template substitution is not allowed", // Even safe is blocked in -c
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ExecutionRequest{
				ToolName:   "shell_tool",
				ToolInputs: []byte(tt.inputs),
				DryRun:     true,
			}
			_, err := cmdTool.Execute(ctx, req)
			if tt.wantErr != "" {
				require.Error(t, err)
				// Relax check to allow both old and new errors if needed, but here we expect stricter
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLocalCommandTool_ArgsParameter(t *testing.T) {
    // Test the "args" parameter special handling
	svcConfig := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("ls"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build() // No defined args, rely on input "args"

	// Tool definition must allow args in input schema
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
	    "type": "object",
	    "properties": map[string]interface{}{
	        "args": map[string]interface{}{
	            "type": "array",
	            "items": map[string]interface{}{
	                "type": "string",
	            },
	        },
	    },
	})

	toolDef := v1.Tool_builder{
	    Name: proto.String("ls_tool"),
	    InputSchema: inputSchema, // Direct assignment of *structpb.Struct
	}.Build()

	cmdTool := NewLocalCommandTool(toolDef, svcConfig, callDef, nil, "call1")
	ctx := context.Background()

	req := &ExecutionRequest{
	    ToolName: "ls_tool",
	    ToolInputs: []byte(`{"args": ["-l", "/tmp"]}`),
	    DryRun: true,
	}

	// This should fail because "-l" triggers argument injection check
	_, err := cmdTool.Execute(ctx, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "argument injection detected")

	// Valid args
	req.ToolInputs = []byte(`{"args": ["foo", "bar"]}`)
	_, err = cmdTool.Execute(ctx, req)
	assert.NoError(t, err)
}

func TestLocalCommandTool_DockerEnv(t *testing.T) {
    // Test that checking for absolute path is skipped for Docker
    svcConfig := configv1.CommandLineUpstreamService_builder{
        Command: proto.String("ls"),
        ContainerEnvironment: configv1.ContainerEnvironment_builder{
            Image: proto.String("ubuntu"),
        }.Build(),
    }.Build()
    callDef := configv1.CommandLineCallDefinition_builder{
        Args: []string{"{{path}}"},
        Parameters: []*configv1.CommandLineParameterMapping{
            configv1.CommandLineParameterMapping_builder{Schema: configv1.ParameterSchema_builder{Name: proto.String("path")}.Build()}.Build(),
        },
    }.Build()
    toolDef := v1.Tool_builder{Name: proto.String("docker_tool")}.Build()

    cmdTool := NewLocalCommandTool(toolDef, svcConfig, callDef, nil, "call1")
    ctx := context.Background()

    // Absolute path should be allowed in Docker
    req := &ExecutionRequest{
        ToolName: "docker_tool",
        ToolInputs: []byte(`{"path": "/etc/passwd"}`),
        DryRun: true,
    }
    _, err := cmdTool.Execute(ctx, req)
    assert.NoError(t, err)
}

func helperSetupHTTPTool(t *testing.T, toolDef *v1.Tool, callDef *configv1.HttpCallDefinition) *HTTPTool {
	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: http.DefaultClient}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("svc", p)

	return NewHTTPTool(toolDef, poolManager, "svc", nil, callDef, nil, nil, "call1")
}

func TestHTTPTool_RootDoubleSlash(t *testing.T) {
	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("path")}.Build(),
				DisableEscape: proto.Bool(true),
			}.Build(),
		},
	}.Build()
	// Note: We use a template that allows us to construct //
	toolDef := v1.Tool_builder{Name: proto.String("http_tool"), UnderlyingMethodFqn: proto.String("GET http://example.com/{{path}}")}.Build()

	httpTool := helperSetupHTTPTool(t, toolDef, callDef)
	ctx := context.Background()

	req := &ExecutionRequest{
		ToolName: "http_tool",
		ToolInputs: []byte(`{"path": "/"}`), // Becomes http://example.com//
		DryRun: true,
	}

	res, err := httpTool.Execute(ctx, req)
	assert.NoError(t, err)

	resMap := res.(map[string]any)["request"].(map[string]any)
	urlStr := resMap["url"].(string)

	// http://example.com// -> cleaned to http://example.com//
	assert.Equal(t, "http://example.com//", urlStr)
}

func TestOpenAPITool_Coverage(t *testing.T) {
	t.Parallel()
	t.Run("POST with Input Template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.Equal(t, `{"name": "test"}`, string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		toolProto := v1.Tool_builder{}.Build()
		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		callDef := configv1.OpenAPICallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Template: proto.String(`{"name": "{{name}}"}`),
			}.Build(),
		}.Build()

		openAPITool := NewOpenAPITool(toolProto, mockClient, nil, "POST", server.URL, nil, callDef)

		inputs := json.RawMessage(`{"name": "test"}`)
		req := &ExecutionRequest{ToolName: "testTool", ToolInputs: inputs}

		_, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)
	})

	t.Run("Output Transformer Template", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"data": "success"}`))
		}))
		defer server.Close()

		toolProto := v1.Tool_builder{}.Build()
		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		callDef := configv1.OpenAPICallDefinition_builder{
			OutputTransformer: configv1.OutputTransformer_builder{
				Format:   configv1.OutputTransformer_JSON.Enum(),
				Template: proto.String(`Result: {{data}}`),
				ExtractionRules: map[string]string{
					"data": "{.data}",
				},
			}.Build(),
		}.Build()

		openAPITool := NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, nil, callDef)

		inputs := json.RawMessage(`{}`)
		req := &ExecutionRequest{ToolName: "testTool", ToolInputs: inputs}

		result, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)

		// Result should be map[string]any{"result": "Result: success"}
		resMap, ok := result.(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "Result: success", resMap["result"])
	})

	t.Run("Init Error", func(t *testing.T) {
		// NewOpenAPITool with invalid template to trigger initError
		callDef := configv1.OpenAPICallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Template: proto.String("{{unclosed"),
			}.Build(),
		}.Build()
		toolProto := v1.Tool_builder{}.Build()
		openAPITool := NewOpenAPITool(toolProto, nil, nil, "POST", "http://example.com", nil, callDef)

		req := &ExecutionRequest{ToolName: "test"}
		_, err := openAPITool.Execute(context.Background(), req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse input template")
	})

	t.Run("Input Transformer via Webhook", func(t *testing.T) {
		webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/cloudevents+json")
			responseEvent := `{
                "specversion": "1.0",
                "type": "com.mcpany.tool.transform_input.response",
                "source": "webhook-test",
                "id": "123",
                "data": {"transformed": "input"}
            }`
			w.Write([]byte(responseEvent))
		}))
		defer webhookServer.Close()

		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			assert.JSONEq(t, `{"transformed": "input"}`, string(body))
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer targetServer.Close()

		toolProto := v1.Tool_builder{}.Build()
		mockClient := &mockHTTPClient{
			doFunc: targetServer.Client().Do,
		}

		callDef := configv1.OpenAPICallDefinition_builder{
			InputTransformer: configv1.InputTransformer_builder{
				Webhook: configv1.WebhookConfig_builder{
					Url: webhookServer.URL,
				}.Build(),
			}.Build(),
		}.Build()

		openAPITool := NewOpenAPITool(toolProto, mockClient, nil, "POST", targetServer.URL, nil, callDef)

		inputs := json.RawMessage(`{}`)
		req := &ExecutionRequest{ToolName: "testTool", ToolInputs: inputs}

		_, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)
	})
}

// Mock Executor for JSON protocol
type mockExecutorForCoverage struct {
    stdout string
    stderr string
    err    error
}

func (m *mockExecutorForCoverage) Execute(ctx context.Context, name string, args []string, dir string, env []string) (io.ReadCloser, io.ReadCloser, <-chan int, error) {
    return nil, nil, nil, errors.New("not implemented")
}

func (m *mockExecutorForCoverage) ExecuteWithStdIO(ctx context.Context, name string, args []string, dir string, env []string) (io.WriteCloser, io.ReadCloser, io.ReadCloser, <-chan int, error) {
    // Return pipes
    prOut, pwOut := io.Pipe()
    go func() {
        pwOut.Write([]byte(m.stdout))
        pwOut.Close()
    }()

    prErr, pwErr := io.Pipe()
    go func() {
        pwErr.Write([]byte(m.stderr))
        pwErr.Close()
    }()

    // Stdin
    prIn, pwIn := io.Pipe()
    go func() {
        io.Copy(io.Discard, prIn)
        prIn.Close()
    }()

    exitChan := make(chan int, 1)
    exitChan <- 0
    close(exitChan)

    return pwIn, prOut, prErr, exitChan, m.err
}

func TestLocalCommandTool_JSONProtocol(t *testing.T) {
    svcConfig := configv1.CommandLineUpstreamService_builder{
        Command: proto.String("my-json-tool"),
        CommunicationProtocol: configv1.CommandLineUpstreamService_COMMUNICATION_PROTOCOL_JSON.Enum(),
    }.Build()
    callDef := configv1.CommandLineCallDefinition_builder{}.Build()
    toolDef := v1.Tool_builder{Name: proto.String("json_tool")}.Build()

    factory := func(env *configv1.ContainerEnvironment) command.Executor {
        return &mockExecutorForCoverage{
            stdout: `{"result": "success"}`,
            stderr: "",
        }
    }

    ct := &CommandTool{
        tool: toolDef,
        service: svcConfig,
        callDefinition: callDef,
        executorFactory: factory,
        callID: "call1",
    }

    req := &ExecutionRequest{
        ToolName: "json_tool",
        ToolInputs: []byte(`{"arg": "val"}`),
    }

    res, err := ct.Execute(context.Background(), req)
    require.NoError(t, err)

    resMap, ok := res.(map[string]any)
    require.True(t, ok)
    assert.Equal(t, "success", resMap["result"])
}
