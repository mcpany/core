<<<<<<< HEAD
=======
// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
package tool

import (
	"context"
<<<<<<< HEAD
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"github.com/mcpany/core/server/pkg/pool"
)

func TestCheckUnquotedKeywords_ExtraCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		val      string
		keywords []string
		wantErr  bool
	}{
		{
			name:     "unquoted keyword",
			val:      "hello system world",
			keywords: []string{"system"},
			wantErr:  true,
		},
		{
			name:     "escaped keyword",
			val:      "hello \\system world",
			keywords: []string{"system"},
			wantErr:  false,
		},
		{
			name:     "single quoted keyword",
			val:      "hello 'system' world",
			keywords: []string{"system"},
			wantErr:  false,
		},
		{
			name:     "double quoted keyword",
			val:      "hello \"system\" world",
			keywords: []string{"system"},
			wantErr:  false,
		},
		{
			name:     "backtick quoted keyword",
			val:      "hello `system` world",
			keywords: []string{"system"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkUnquotedKeywords(tt.val, tt.keywords)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckAwkInjection_ExtraCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		val     string
		base    string
		wantErr bool
	}{
		{
			name:    "not awk",
			val:     "|",
			base:    "echo",
			wantErr: false,
		},
		{
			name:    "awk pipe",
			val:     "print | something",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "gawk redirect in",
			val:     "< file",
			base:    "gawk",
			wantErr: true,
		},
		{
			name:    "nawk redirect out",
			val:     "> file",
			base:    "nawk",
			wantErr: true,
		},
		{
			name:    "mawk getline",
			val:     "getline",
			base:    "mawk",
			wantErr: true,
		},
		{
			name:    "awk indirect func",
			val:     "@fun",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "awk include",
			val:     "@include",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "awk load",
			val:     "@load",
			base:    "awk",
			wantErr: true,
		},
		{
			name:    "awk append",
			val:     ">>",
			base:    "awk",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkAwkInjection(tt.val, tt.base)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckNodePerlPhpInjection_ExtraCoverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		val        string
		base       string
		quoteLevel int
		wantErr    bool
	}{
		{
			name:       "not node/perl/php",
			val:        "${injection}",
			base:       "python",
			quoteLevel: 1, // Double quotes
			wantErr:    false,
		},
		{
			name:       "node unquoted",
			val:        "${injection}",
			base:       "node",
			quoteLevel: 0,
			wantErr:    false,
		},
		{
			name:       "node backticks",
			val:        "${injection}",
			base:       "node",
			quoteLevel: 3, // backticks
			wantErr:    true,
		},
		{
			name:       "bun backticks",
			val:        "${injection}",
			base:       "bun",
			quoteLevel: 3,
			wantErr:    true,
		},
		{
			name:       "deno backticks",
			val:        "${injection}",
			base:       "deno",
			quoteLevel: 3,
			wantErr:    true,
		},
		{
			name:       "perl double quotes",
			val:        "${injection}",
			base:       "perl",
			quoteLevel: 1, // double quotes
			wantErr:    true,
		},
		{
			name:       "php double quotes",
			val:        "${injection}",
			base:       "php",
			quoteLevel: 1, // double quotes
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := checkNodePerlPhpInjection(tt.val, tt.base, tt.quoteLevel)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGRPCTool_Coverage(t *testing.T) {
	t.Parallel()

	toolProto := v1.Tool_builder{
		Name: proto.String("test-grpc-tool"),
	}.Build()

	methodDesc := new(MockMethodDescriptor)
	msgDesc := new(MockMessageDescriptor)
	msgDesc.On("FullName").Return(protoreflect.FullName("TestMessage"))
	methodDesc.On("Input").Return(msgDesc)
	methodDesc.On("FullName").Return(protoreflect.FullName("TestService.TestMethod"))

	poolManager := pool.NewManager()
	callDef := &configv1.GrpcCallDefinition{}
	resilienceCfg := &configv1.ResilienceConfig{}

	grpcTool := NewGRPCTool(toolProto, poolManager, "test-service", methodDesc, callDef, resilienceCfg)

	// Test IsStreaming
	assert.False(t, grpcTool.IsStreaming())

	// Test StreamExecute
	ch, err := grpcTool.StreamExecute(context.Background(), &ExecutionRequest{})
	assert.NoError(t, err)
	// it should execute and return an error (due to pool manager error, etc.)
	select {
	case res := <-ch:
		errRes, ok := res.(error)
		assert.True(t, ok)
		assert.Error(t, errRes)
	case <-time.After(time.Second):
		t.Fatal("StreamExecute did not return anything")
	}

	// Test MCPTool (it should at least initialize and return a non-nil object or gracefully handle if tool is incomplete)
	mcpTool := grpcTool.MCPTool()
	assert.NotNil(t, mcpTool)
=======
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

func TestCheckForLocalFileAccess(t *testing.T) {
	assert.Error(t, checkForLocalFileAccess("/absolute"))
	assert.Error(t, checkForLocalFileAccess("file:///etc/passwd"))
	assert.Error(t, checkForLocalFileAccess("FILE:///etc/passwd"))
	assert.Error(t, checkForLocalFileAccess("file:foo"))
	assert.NoError(t, checkForLocalFileAccess("relative"))
}

func TestCheckForArgumentInjection(t *testing.T) {
    assert.Error(t, checkForArgumentInjection("-flag"))
    assert.NoError(t, checkForArgumentInjection("-123")) // Number allowed
    assert.NoError(t, checkForArgumentInjection("safe"))
}

func TestCheckForShellInjection(t *testing.T) {
    assert.Error(t, checkForShellInjection("safe; rm -rf /", "", "", "sh", true))
    assert.NoError(t, checkForShellInjection("safe", "", "", "sh", true))

    // Single quoted context
    assert.Error(t, checkForShellInjection("break'out", "'{{val}}'", "{{val}}", "sh", true))
    assert.NoError(t, checkForShellInjection("safe; rm", "'{{val}}'", "{{val}}", "sh", true))

    // Double quoted context
    assert.Error(t, checkForShellInjection("break\"out", "\"{{val}}\"", "{{val}}", "sh", true))
    assert.Error(t, checkForShellInjection("$var", "\"{{val}}\"", "{{val}}", "sh", true))
    assert.NoError(t, checkForShellInjection("safe space", "\"{{val}}\"", "{{val}}", "sh", true))

    // Extended unquoted
    assert.Error(t, checkForShellInjection("val|ue", "", "", "sh", true))
    assert.Error(t, checkForShellInjection("val&ue", "", "", "sh", true))
    assert.Error(t, checkForShellInjection("val>ue", "", "", "sh", true))

    // Space check
    assert.Error(t, checkForShellInjection("safe space", "", "", "sh", true), "shell should block space in unquoted context")
    assert.NoError(t, checkForShellInjection("safe space", "", "", "python", false), "interpreter should allow space in unquoted context")

    // Env command specific
    assert.Error(t, checkForShellInjection("VAR=val", "", "", "env", true), "env command should block '='")
    assert.NoError(t, checkForShellInjection("VAR=val", "", "", "sh", true), "sh command should allow '='")
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
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))
}
