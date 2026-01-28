// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
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
    // Unquoted - always strict regardless of mode
    assert.Error(t, checkForShellInjection("safe; rm -rf /", "", "", "sh", false))
    assert.NoError(t, checkForShellInjection("safe", "", "", "sh", false))

    // Single quoted context - Strict Mode (Interpreter)
    assert.Error(t, checkForShellInjection("break'out", "'{{val}}'", "{{val}}", "sh", true))
    assert.Error(t, checkForShellInjection("safe; rm", "'{{val}}'", "{{val}}", "sh", true), "strict mode blocks ; even in quotes")

    // Single quoted context - Standard Mode (Sensitive Command like curl)
    assert.Error(t, checkForShellInjection("break'out", "'{{val}}'", "{{val}}", "curl", false))
    assert.NoError(t, checkForShellInjection("safe; rm", "'{{val}}'", "{{val}}", "curl", false), "standard mode allows ; in quotes")

    // Double quoted context - Strict Mode
    assert.Error(t, checkForShellInjection("break\"out", "\"{{val}}\"", "{{val}}", "sh", true))
    assert.Error(t, checkForShellInjection("$var", "\"{{val}}\"", "{{val}}", "sh", true))
    assert.NoError(t, checkForShellInjection("safe space", "\"{{val}}\"", "{{val}}", "sh", true), "space is allowed even in strict mode")

    // Double quoted context - Standard Mode
    assert.Error(t, checkForShellInjection("break\"out", "\"{{val}}\"", "{{val}}", "curl", false))
    assert.Error(t, checkForShellInjection("$var", "\"{{val}}\"", "{{val}}", "curl", false))
    assert.NoError(t, checkForShellInjection("safe; space", "\"{{val}}\"", "{{val}}", "curl", false))

    // Extended unquoted
    assert.Error(t, checkForShellInjection("val|ue", "", "", "sh", true))
    assert.Error(t, checkForShellInjection("val&ue", "", "", "sh", true))
    assert.Error(t, checkForShellInjection("val>ue", "", "", "sh", true))

    // Env command specific
    assert.Error(t, checkForShellInjection("VAR=val", "", "", "env", true), "env command should block '='")
    assert.NoError(t, checkForShellInjection("VAR=val", "", "", "sh", true), "sh command should allow '='")
}

func TestIsInterpreter(t *testing.T) {
    assert.True(t, isInterpreter("bash"))
    assert.True(t, isInterpreter("/bin/sh"))
    // python is no longer considered strict interpreter for shell injection purposes
    // it falls into sensitive/standard validation
    assert.False(t, isInterpreter("python"))
    assert.False(t, isInterpreter("curl"))
    assert.False(t, isInterpreter("echo"))
}

func TestIsSensitiveCommand(t *testing.T) {
    assert.True(t, isSensitiveCommand("curl"))
    assert.True(t, isSensitiveCommand("git"))
    assert.True(t, isSensitiveCommand("python")) // python is sensitive
    assert.False(t, isSensitiveCommand("bash")) // bash is strict interpreter
    assert.False(t, isSensitiveCommand("echo"))
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
