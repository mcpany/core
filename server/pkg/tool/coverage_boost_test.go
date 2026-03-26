// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// MockExecutor for testing CommandTool
// MockExecutor for testing CommandTool
// Summary: MockExecutor
	stdout   string
	stderr   string
	exitCode int
}

// Execute ...
// Summary: Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	outR := io.NopCloser(strings.NewReader(m.stdout))
	errR := io.NopCloser(strings.NewReader(m.stderr))
	ch := make(chan int, 1)
	ch <- m.exitCode
	close(ch)
	return outR, errR, ch, nil
}

// ExecuteWithStdIO ...
// Summary: ExecuteWithStdIO
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil, nil, nil, nil
}

// ... setupHTTPTool ...
func setupHTTPTool(t *testing.T, handler http.Handler, callDef *configv1.HttpCallDefinition) (*HTTPTool, *httptest.Server) {
	server := httptest.NewServer(handler)
	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("s", p)

	method := "GET " + server.URL
	toolDef := v1.Tool_builder{
		UnderlyingMethodFqn: proto.String(method),
		Name:                proto.String("test-http"),
	}.Build()
	return NewHTTPTool(toolDef, poolManager, "s", nil, callDef, nil, nil, ""), server
}

// TestHTTPTool_Execute_DryRun ...
// Summary: TestHTTPTool_Execute_DryRun
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Should not be called in dry run")
	})
	callDef := &configv1.HttpCallDefinition{}
	tool, server := setupHTTPTool(t, handler, callDef)
	defer server.Close()

	req := &ExecutionRequest{
		ToolName:   "test-http",
		ToolInputs: []byte(`{}`),
		DryRun:     true,
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap := res.(map[string]any)
	assert.True(t, resMap["dry_run"].(bool))
}

// TestHTTPTool_Execute_InvalidInputJSON ...
// Summary: TestHTTPTool_Execute_InvalidInputJSON
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	callDef := &configv1.HttpCallDefinition{}
	tool, server := setupHTTPTool(t, handler, callDef)
	defer server.Close()

	req := &ExecutionRequest{
		ToolInputs: []byte(`{invalid-json`),
	}
	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to unmarshal tool inputs")
}

// TestCommandTool_DryRun ...
// Summary: TestCommandTool_DryRun
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"hello"},
	}.Build()
	toolDef := v1.Tool_builder{Name: proto.String("test")}.Build()
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "id")

	req := &ExecutionRequest{
		ToolName:   "test",
		ToolInputs: []byte(`{}`),
		DryRun:     true,
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)

	resMap, ok := res.(map[string]any)
	assert.True(t, ok, "result should be a map")
	if !ok {
		return
	}

	val, exists := resMap["dry_run"]
	if !exists {
		t.Errorf("dry_run key missing in response: %+v", resMap)
		return
	}

	assert.True(t, val.(bool))
}

// TestCommandTool_ResolveServiceEnv_Error ...
// Summary: TestCommandTool_ResolveServiceEnv_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Test error when resolving service env secrets
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		Env: map[string]*configv1.SecretValue{
			"SECRET": configv1.SecretValue_builder{
				EnvironmentVariable: proto.String("MISSING_VAR_XYZ_123"),
			}.Build(),
		},
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()
	toolDef := v1.Tool_builder{Name: proto.String("test")}.Build()
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "id")

	req := &ExecutionRequest{
		ToolInputs: []byte(`{}`),
	}

	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve service env")
}

// TestCommandTool_ResolveContainerEnv_Error ...
// Summary: TestCommandTool_ResolveContainerEnv_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Test error when resolving container env secrets
	// We use NewCommandTool because LocalCommandTool does not handle container envs
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		ContainerEnvironment: configv1.ContainerEnvironment_builder{
			Image: proto.String("ubuntu"),
			Env: map[string]*configv1.SecretValue{
				"SECRET": configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MISSING_VAR_XYZ_123"),
				}.Build(),
			},
		}.Build(),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{}.Build()
	toolDef := v1.Tool_builder{Name: proto.String("test")}.Build()
	// Compile policies manually or pass nil (NewCommandTool takes policies slice)
	tool := NewCommandTool(toolDef, service, callDef, nil, "id")

	req := &ExecutionRequest{
		ToolInputs: []byte(`{}`),
	}

	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve container env")
}

// TestCommandTool_ResolveParameterSecret_Error ...
// Summary: TestCommandTool_ResolveParameterSecret_Error
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Parameters: []*configv1.CommandLineParameterMapping{
			configv1.CommandLineParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{Name: proto.String("secret")}.Build(),
				Secret: configv1.SecretValue_builder{
					EnvironmentVariable: proto.String("MISSING_VAR_XYZ_123"),
				}.Build(),
			}.Build(),
		},
	}.Build()
	toolDef := v1.Tool_builder{Name: proto.String("test")}.Build()
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "id")

	req := &ExecutionRequest{
		ToolInputs: []byte(`{}`),
	}

	_, err := tool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to resolve secret")
}

// TestCommandTool_Success ...
// Summary: TestCommandTool_Success
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	service := configv1.CommandLineUpstreamService_builder{
		Command: proto.String("echo"),
		ContainerEnvironment: configv1.ContainerEnvironment_builder{
			Image: proto.String("ubuntu"),
		}.Build(),
	}.Build()
	callDef := configv1.CommandLineCallDefinition_builder{
		Args: []string{"hello"},
	}.Build()
	toolDef := v1.Tool_builder{Name: proto.String("test")}.Build()
	tool := NewCommandTool(toolDef, service, callDef, nil, "id")

	// Inject mock executor
	ct := tool.(*CommandTool)
	ct.executorFactory = func(env *configv1.ContainerEnvironment) command.Executor {
		return &MockExecutor{
			stdout:   "hello\n",
			stderr:   "",
			exitCode: 0,
		}
	}

	req := &ExecutionRequest{
		ToolInputs: []byte(`{}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap := res.(map[string]any)
	assert.Equal(t, "hello\n", resMap["stdout"])
	// json.Number handling
	rc, ok := resMap["return_code"].(json.Number)
	if ok {
		f, _ := rc.Float64()
		assert.Equal(t, float64(0), f)
	} else {
		// It might be int if not unmarshaled from JSON?
		// CommandTool returns map[string]interface{} constructed in code.
		// "return_code": exitCode (int)
		assert.Equal(t, 0, resMap["return_code"])
	}
}

// TestHTTPTool_OutputTransformer ...
// Summary: TestHTTPTool_OutputTransformer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"foo": "bar"}`))
	})

	// Test output transformer template
	callDef := configv1.HttpCallDefinition_builder{
		OutputTransformer: configv1.OutputTransformer_builder{
			Template: proto.String("Foo is {{foo}}"),
			Format:   configv1.OutputTransformer_JSON.Enum(),
			ExtractionRules: map[string]string{
				"foo": "{.foo}",
			},
		}.Build(),
	}.Build()
	tool, server := setupHTTPTool(t, handler, callDef)
	defer server.Close()

	req := &ExecutionRequest{
		ToolName:   "test-http",
		ToolInputs: []byte(`{}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap := res.(map[string]any)
	assert.Equal(t, "Foo is bar", resMap["result"])
}

// TestHTTPTool_OutputTransformer_Raw ...
// Summary: TestHTTPTool_OutputTransformer_Raw
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`raw data`))
	})

	callDef := configv1.HttpCallDefinition_builder{
		OutputTransformer: configv1.OutputTransformer_builder{
			Format: configv1.OutputTransformer_RAW_BYTES.Enum(),
		}.Build(),
	}.Build()
	tool, server := setupHTTPTool(t, handler, callDef)
	defer server.Close()

	req := &ExecutionRequest{
		ToolName:   "test-http",
		ToolInputs: []byte(`{}`),
	}

	res, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)
	resMap := res.(map[string]any)
	assert.Equal(t, []byte("raw data"), resMap["raw"])
}

// TestHTTPTool_InputTransformer_Template ...
// Summary: TestHTTPTool_InputTransformer_Template
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// We need to use POST to trigger body preparation
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		assert.Equal(t, "Hello World", buf.String())
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	poolManager.Register("s", p)

	method := "POST " + server.URL

	callDef := configv1.HttpCallDefinition_builder{
		InputTransformer: configv1.InputTransformer_builder{
			Template: proto.String("Hello {{name}}"),
		}.Build(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("name"),
				}.Build(),
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{
		UnderlyingMethodFqn: proto.String(method),
		Name:                proto.String("test-post"),
	}.Build()
	tool := NewHTTPTool(toolDef, poolManager, "s", nil, callDef, nil, nil, "")

	req := &ExecutionRequest{
		ToolName:   "test-post",
		ToolInputs: []byte(`{"name": "World"}`),
	}

	_, err := tool.Execute(context.Background(), req)
	assert.NoError(t, err)
}
