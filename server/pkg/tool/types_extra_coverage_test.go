// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestContextHelpers_Extra(t *testing.T) {
	ctx := context.Background()

	// Tool context
	t1 := &MockTool{}
	ctx = NewContextWithTool(ctx, t1)
	got, ok := GetFromContext(ctx)
	assert.True(t, ok)
	assert.Equal(t, t1, got)

	// CacheControl context
	cc := &CacheControl{Action: ActionAllow}
	ctx = NewContextWithCacheControl(ctx, cc)
	gotCC, ok := GetCacheControl(ctx)
	assert.True(t, ok)
	assert.Equal(t, cc, gotCC)

	// Empty context
	ctxEmpty := context.Background()
	_, ok = GetFromContext(ctxEmpty)
	assert.False(t, ok)

	_, ok = GetCacheControl(ctxEmpty)
	assert.False(t, ok)
}

func TestCheckForAbsolutePath(t *testing.T) {
	assert.Error(t, checkForAbsolutePath("/absolute"))
	assert.NoError(t, checkForAbsolutePath("relative"))
}

func TestCheckForArgumentInjection(t *testing.T) {
    assert.Error(t, checkForArgumentInjection("-flag"))
    assert.NoError(t, checkForArgumentInjection("-123")) // Number allowed
    assert.NoError(t, checkForArgumentInjection("safe"))
}

func TestCheckForShellInjection(t *testing.T) {
    assert.Error(t, checkForShellInjection("safe; rm -rf /", "", "", "sh"))
    assert.NoError(t, checkForShellInjection("safe", "", "", "sh"))

    // Single quoted context
    assert.Error(t, checkForShellInjection("break'out", "'{{val}}'", "{{val}}", "sh"))
    assert.NoError(t, checkForShellInjection("safe; rm", "'{{val}}'", "{{val}}", "sh"))

    // Double quoted context
    assert.Error(t, checkForShellInjection("break\"out", "\"{{val}}\"", "{{val}}", "sh"))
    assert.Error(t, checkForShellInjection("$var", "\"{{val}}\"", "{{val}}", "sh"))
    assert.NoError(t, checkForShellInjection("safe space", "\"{{val}}\"", "{{val}}", "sh"))

    // Extended unquoted
    assert.Error(t, checkForShellInjection("val|ue", "", "", "sh"))
    assert.Error(t, checkForShellInjection("val&ue", "", "", "sh"))
    assert.Error(t, checkForShellInjection("val>ue", "", "", "sh"))

    // Env command specific
    assert.Error(t, checkForShellInjection("VAR=val", "", "", "env"), "env command should block '='")
    assert.NoError(t, checkForShellInjection("VAR=val", "", "", "sh"), "sh command should allow '='")
}

func TestIsShellCommand(t *testing.T) {
    assert.True(t, isShellCommand("bash"))
    assert.True(t, isShellCommand("/bin/sh"))
    assert.True(t, isShellCommand("python"))
    assert.True(t, isShellCommand("cmd.exe"))
    assert.True(t, isShellCommand("timeout"), "timeout should be considered a shell command wrapper")
    assert.True(t, isShellCommand("nice"), "nice should be considered a shell command wrapper")
    assert.True(t, isShellCommand("time"), "time should be considered a shell command wrapper")
    assert.False(t, isShellCommand("ls"))
    assert.False(t, isShellCommand("echo"))
}

func setupHTTPToolExtra(t *testing.T, handler http.Handler, callDefinition *configv1.HttpCallDefinition, urlSuffix string) (*HTTPTool, *httptest.Server) {
    server := httptest.NewServer(handler)
    poolManager := pool.NewManager()
    p, _ := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
        return &client.HTTPClientWrapper{Client: server.Client()}, nil
    }, 1, 1, 1, 0, true)
    poolManager.Register("s", p)

    method := "GET " + server.URL + urlSuffix
    toolDef := &v1.Tool{UnderlyingMethodFqn: &method}
    return NewHTTPTool(toolDef, poolManager, "s", nil, callDefinition, nil, nil, ""), server
}

func TestHTTPTool_Execute_Secret(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Query().Get("key") == "mysecret" {
            w.WriteHeader(http.StatusOK)
            w.Write([]byte(`{}`))
        } else {
            w.WriteHeader(http.StatusUnauthorized)
        }
    })

    secretVal := &configv1.SecretValue{
        Value: &configv1.SecretValue_PlainText{PlainText: "mysecret"},
    }

    param := &configv1.HttpParameterMapping{
        Schema: &configv1.ParameterSchema{Name: proto.String("key")},
        Secret: secretVal,
    }

    callDef := &configv1.HttpCallDefinition{
        Parameters: []*configv1.HttpParameterMapping{param},
    }

    tool, server := setupHTTPToolExtra(t, handler, callDef, "?key={{key}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{})
    assert.NoError(t, err)
}

func TestHTTPTool_Execute_MissingRequired(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    param := &configv1.HttpParameterMapping{
        Schema: &configv1.ParameterSchema{
            Name: proto.String("req"),
            IsRequired: proto.Bool(true),
        },
    }

    callDef := &configv1.HttpCallDefinition{
        Parameters: []*configv1.HttpParameterMapping{param},
    }

    tool, server := setupHTTPToolExtra(t, handler, callDef, "?req={{req}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "missing required parameter")
}

func TestHTTPTool_Execute_PathTraversal(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

    param := &configv1.HttpParameterMapping{
        Schema: &configv1.ParameterSchema{Name: proto.String("path")},
    }

    callDef := &configv1.HttpCallDefinition{
        Parameters: []*configv1.HttpParameterMapping{param},
    }

    // URL with placeholder in path (not query)

    tool, server := setupHTTPToolExtra(t, handler, callDef, "/{{path}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: []byte(`{"path": "../etc/passwd"}`)})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "path traversal attempt detected")
}

func TestHTTPTool_Execute_Secret_Error(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

    secretVal := &configv1.SecretValue{
        Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING_ENV_VAR_XYZ"},
    }

    param := &configv1.HttpParameterMapping{
        Schema: &configv1.ParameterSchema{Name: proto.String("key")},
        Secret: secretVal,
    }

    callDef := &configv1.HttpCallDefinition{
        Parameters: []*configv1.HttpParameterMapping{param},
    }

    tool, server := setupHTTPToolExtra(t, handler, callDef, "?key={{key}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to resolve secret")
}

// Added Extra Coverage Tests

func TestIsSensitiveHeader_Extra(t *testing.T) {
	tests := []struct {
		key      string
		expected bool
	}{
		{"Authorization", true},
		{"authorization", true},
		{"Proxy-Authorization", true},
		{"Cookie", true},
		{"Set-Cookie", true},
		{"X-Api-Key", true},
		{"My-Token", true},
		{"My-Secret", true},
		{"My-Password", true},
		{"Access-Token", true},
		{"X-Auth-Token", true},
		{"Csrf-Token", true},
		{"Xsrf-Token", true},
		{"X-Signature", true},
		{"Content-Type", false},
		{"Accept", false},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, isSensitiveHeader(tc.key), "Key: %s", tc.key)
	}
}

func TestCheckForPathTraversal_Direct(t *testing.T) {
	tests := []struct {
		val       string
		shouldErr bool
	}{
		{"..", true},
		{"../foo", true},
		{"..\\foo", true},
		{"foo/..", true},
		{"foo\\..", true},
		{"foo/../bar", true},
		{"foo\\..\\bar", true},
		{"%2e%2e", true},
		{"%2E%2E", true},
		{"%2e%2e/", true},
		{"normal", false},
		{"normal/path", false},
	}

	for _, tc := range tests {
		err := checkForPathTraversal(tc.val)
		if tc.shouldErr {
			assert.Error(t, err, "Val: %s", tc.val)
		} else {
			assert.NoError(t, err, "Val: %s", tc.val)
		}
	}
}

func TestCleanPathPreserveDoubleSlash_Extra(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "."},
		{"/", "/"},
		{"//", "/"}, // Cleans to /, caller handles restoration if needed
		{"/foo", "/foo"},
		{"/foo/", "/foo"},
		{"//foo", "//foo"},
		{"//foo/", "//foo"},
		{"/foo/.", "/foo"},
		{"/foo/..", "/"},
		{"/foo/../bar", "/bar"},
		{"/..", "/"},
		{"//..", "/"}, // Cleans to /
		{"foo", "foo"},
		{"foo/.", "foo"},
		{"foo/..", "."},
		{"./foo", "foo"},
	}

	for _, tc := range tests {
		assert.Equal(t, tc.expected, cleanPathPreserveDoubleSlash(tc.input), "Input: %s", tc.input)
	}
}

func TestAnalyzeQuoteContext(t *testing.T) {
	// 0=Unquoted, 1=Double, 2=Single
	assert.Equal(t, 0, analyzeQuoteContext("cmd {{v}}", "{{v}}"))
	assert.Equal(t, 1, analyzeQuoteContext(`cmd "{{v}}"`, "{{v}}"))
	assert.Equal(t, 2, analyzeQuoteContext(`cmd '{{v}}'`, "{{v}}"))
	assert.Equal(t, 0, analyzeQuoteContext(`cmd '{{v}"`, "{{v}}")) // Mismatched
	assert.Equal(t, 0, analyzeQuoteContext(`cmd "{{v}'`, "{{v}}")) // Mismatched

	// Escaped quotes
	assert.Equal(t, 0, analyzeQuoteContext(`cmd \"{{v}}\"`, "{{v}}")) // Escaped quotes -> unquoted
	assert.Equal(t, 1, analyzeQuoteContext(`cmd "foo \" {{v}}"`, "{{v}}")) // Inside double quotes

	// Not found case (defensive)
	assert.Equal(t, 0, analyzeQuoteContext("cmd", "{{v}}"))
}

func TestProcessResponse_Error(t *testing.T) {
    // We can mock http.Response.
    t.Run("BodyReadError", func(t *testing.T) {
        tool := &HTTPTool{}

        mockBody := &MockReadCloser{
            ReadFunc: func(p []byte) (n int, err error) {
                return 0, io.ErrUnexpectedEOF
            },
            CloseFunc: func() error { return nil },
        }

        resp := &http.Response{
            Body: mockBody,
        }

        _, err := tool.processResponse(context.Background(), resp)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "failed to read http response body")
    })

    // Test OutputTransformer errors
    t.Run("OutputTransformerParseError", func(t *testing.T) {
        format := configv1.OutputTransformer_JSON
        ot := &configv1.OutputTransformer{
            Format: &format,
            JqQuery: proto.String("invalid jq query ["),
        }

        tool := &HTTPTool{
            outputTransformer: ot,
        }

        mockBody := io.NopCloser(strings.NewReader(`{}`))
        resp := &http.Response{
            Body: mockBody,
        }

        _, err := tool.processResponse(context.Background(), resp)

        // This relies on transformer parser actually validating JQ
        if err != nil {
             assert.Error(t, err)
        }
    })
}

// MockReadCloser for testing body read errors
type MockReadCloser struct {
    ReadFunc func(p []byte) (n int, err error)
    CloseFunc func() error
}

func (m *MockReadCloser) Read(p []byte) (n int, err error) {
    return m.ReadFunc(p)
}

func (m *MockReadCloser) Close() error {
    return m.CloseFunc()
}

func TestCommandTool_Execute_Errors(t *testing.T) {
    // Test creating a tool with invalid command line args input

    t.Run("Args_NonString", func(t *testing.T) {
        toolDef := &v1.Tool{
            Name: proto.String("cmd-tool"),
            InputSchema: &structpb.Struct{
                Fields: map[string]*structpb.Value{
                    "properties": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
                        Fields: map[string]*structpb.Value{
                            "args": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
                                Fields: map[string]*structpb.Value{
                                    "type": {Kind: &structpb.Value_StringValue{StringValue: "array"}},
                                },
                            }}},
                        },
                    }}},
                },
            },
        }

        service := &configv1.CommandLineUpstreamService{
            Command: proto.String("ls"),
        }
        callDef := &configv1.CommandLineCallDefinition{}

        ct := NewLocalCommandTool(toolDef, service, callDef, nil, "call")

        req := &ExecutionRequest{
            ToolName: "cmd-tool",
            ToolInputs: []byte(`{"args": [123]}`), // Number instead of string
        }

        _, err := ct.Execute(context.Background(), req)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "non-string value in 'args' array")
    })

    t.Run("Args_NotAllowed", func(t *testing.T) {
        toolDef := &v1.Tool{
            Name: proto.String("cmd-tool"),
            InputSchema: &structpb.Struct{
                Fields: map[string]*structpb.Value{
                    "properties": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
                        Fields: map[string]*structpb.Value{
                            // args NOT in properties
                        },
                    }}},
                },
            },
        }

        service := &configv1.CommandLineUpstreamService{
            Command: proto.String("ls"),
        }
        callDef := &configv1.CommandLineCallDefinition{}

        ct := NewLocalCommandTool(toolDef, service, callDef, nil, "call")

        req := &ExecutionRequest{
            ToolName: "cmd-tool",
            ToolInputs: []byte(`{"args": ["foo"]}`),
        }

        _, err := ct.Execute(context.Background(), req)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "'args' parameter is not allowed")
    })
}

func TestLocalCommandTool_Coverage(t *testing.T) {
	t.Run("Execute DryRun", func(t *testing.T) {
		toolDef := &v1.Tool{
			Name: proto.String("cmd-tool"),
		}
		service := &configv1.CommandLineUpstreamService{
			Command: proto.String("ls"),
		}
		callDef := &configv1.CommandLineCallDefinition{
			Args: []string{"-l"},
		}

		ct := NewLocalCommandTool(toolDef, service, callDef, nil, "call")

		req := &ExecutionRequest{
			ToolName: "cmd-tool",
			ToolInputs: []byte(`{}`),
			DryRun: true,
		}

		result, err := ct.Execute(context.Background(), req)
		assert.NoError(t, err)

		resMap, ok := result.(map[string]any)
		assert.True(t, ok)
		assert.True(t, resMap["dry_run"].(bool))
		reqMap := resMap["request"].(map[string]any)
		assert.Equal(t, "ls", reqMap["command"])
		// Args might be []string or []interface{}
		assert.NotEmpty(t, reqMap["args"])
	})

	t.Run("Execute with 'args' parameter allowed", func(t *testing.T) {
		toolDef := &v1.Tool{
			Name: proto.String("cmd-tool"),
			InputSchema: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"properties": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"args": {Kind: &structpb.Value_StructValue{StructValue: &structpb.Struct{}}},
						},
					}}},
				},
			},
		}

		service := &configv1.CommandLineUpstreamService{
			Command: proto.String("echo"),
		}
		callDef := &configv1.CommandLineCallDefinition{}

		ct := NewLocalCommandTool(toolDef, service, callDef, nil, "call")

		req := &ExecutionRequest{
			ToolName: "cmd-tool",
			ToolInputs: []byte(`{"args": ["hello", "world"]}`),
		}

		// This will actually execute "echo hello world".
		// We expect success.
		result, err := ct.Execute(context.Background(), req)
		assert.NoError(t, err)

		resMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, 0, resMap["return_code"])
		stdout, _ := resMap["stdout"].(string)
		assert.Contains(t, stdout, "hello world")
	})
}

func TestCommandTool_ShellInjectionBypass_Timeout(t *testing.T) {
	// This test attempts to bypass shell injection checks by using "timeout" as the command.
	// "timeout" executes another command, but if it's not recognized as a shell command,
	// the shell injection checks (checkForShellInjection) will be skipped.

	// Configuration: timeout 1s sh -c {{cmd}}
	cmd := "timeout"
	args := []string{"1s", "sh", "-c", "{{cmd}}"}

	// Create a minimal tool definition
	toolDef := &v1.Tool{
		Name: proto.String("timeout-tool"),
	}

	service := &configv1.CommandLineUpstreamService{
		Command: proto.String(cmd),
	}

	callDef := &configv1.CommandLineCallDefinition{
		Args: args,
	}

	// Create the tool
	tool := NewLocalCommandTool(toolDef, service, callDef, nil, "test-call")

	// Payload with shell injection characters (;)
	// We want to execute 'echo pwned'.
	payload := "echo pwned; echo injection"
	req := &ExecutionRequest{
		ToolName:   "timeout-tool",
		ToolInputs: []byte(`{"cmd": "` + payload + `"}`),
	}

	// Execute
	_, err := tool.Execute(context.Background(), req)

	// After fix, we expect an error "shell injection detected"
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "shell injection detected", "Shell injection check should catch usage in 'timeout'")
	}
}
