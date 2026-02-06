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

func TestCheckForDangerousSchemes(t *testing.T) {
	// Host execution (isDocker=false)
	assert.Error(t, checkForDangerousSchemes("file:///etc/passwd", false))
	assert.Error(t, checkForDangerousSchemes("FILE:///etc/passwd", false))
	assert.Error(t, checkForDangerousSchemes("file:foo", false))
	assert.Error(t, checkForDangerousSchemes("gopher://127.0.0.1:6379/_SLAVEOF...", false))
	assert.Error(t, checkForDangerousSchemes("dict://127.0.0.1:2628/quit", false))
	assert.Error(t, checkForDangerousSchemes("ldap://127.0.0.1", false))
	assert.Error(t, checkForDangerousSchemes("tftp://127.0.0.1", false))
	assert.Error(t, checkForDangerousSchemes("expect://id", false))
	assert.NoError(t, checkForDangerousSchemes("http://google.com", false))
	assert.NoError(t, checkForDangerousSchemes("https://google.com", false))
	assert.NoError(t, checkForDangerousSchemes("/absolute/path", false)) // Handled by checkForLocalFileAccess

	// Docker execution (isDocker=true)
	// file: scheme is allowed in Docker as it might be needed for internal container operations
	// and container isolation provides some protection.
	assert.NoError(t, checkForDangerousSchemes("file:///etc/passwd", true))
	assert.Error(t, checkForDangerousSchemes("gopher://127.0.0.1:6379/_SLAVEOF...", true))
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
    toolDef := v1.Tool_builder{UnderlyingMethodFqn: proto.String(method)}.Build()
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

    secretVal := configv1.SecretValue_builder{
        PlainText: proto.String("mysecret"),
    }.Build()

    param := configv1.HttpParameterMapping_builder{
        Schema: configv1.ParameterSchema_builder{Name: proto.String("key")}.Build(),
        Secret: secretVal,
    }.Build()

    callDef := configv1.HttpCallDefinition_builder{
        Parameters: []*configv1.HttpParameterMapping{param},
    }.Build()

    tool, server := setupHTTPToolExtra(t, handler, callDef, "?key={{key}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{})
    assert.NoError(t, err)
}

func TestHTTPTool_Execute_MissingRequired(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    })

    param := configv1.HttpParameterMapping_builder{
        Schema: configv1.ParameterSchema_builder{
            Name: proto.String("req"),
            IsRequired: proto.Bool(true),
        }.Build(),
    }.Build()

    callDef := configv1.HttpCallDefinition_builder{
        Parameters: []*configv1.HttpParameterMapping{param},
    }.Build()

    tool, server := setupHTTPToolExtra(t, handler, callDef, "?req={{req}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "missing required parameter")
}

func TestHTTPTool_Execute_PathTraversal(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

    param := configv1.HttpParameterMapping_builder{
        Schema: configv1.ParameterSchema_builder{Name: proto.String("path")}.Build(),
    }.Build()

    callDef := configv1.HttpCallDefinition_builder{
        Parameters: []*configv1.HttpParameterMapping{param},
    }.Build()

    // URL with placeholder in path (not query)

    tool, server := setupHTTPToolExtra(t, handler, callDef, "/{{path}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{ToolInputs: []byte(`{"path": "../etc/passwd"}`)})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "path traversal attempt detected")
}

func TestHTTPTool_Execute_Secret_Error(t *testing.T) {
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

    secretVal := configv1.SecretValue_builder{
        EnvironmentVariable: proto.String("MISSING_ENV_VAR_XYZ"),
    }.Build()

    param := configv1.HttpParameterMapping_builder{
        Schema: configv1.ParameterSchema_builder{Name: proto.String("key")}.Build(),
        Secret: secretVal,
    }.Build()

    callDef := configv1.HttpCallDefinition_builder{
        Parameters: []*configv1.HttpParameterMapping{param},
    }.Build()

    tool, server := setupHTTPToolExtra(t, handler, callDef, "?key={{key}}")
    defer server.Close()

    _, err := tool.Execute(context.Background(), &ExecutionRequest{})
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "failed to resolve secret")
}
