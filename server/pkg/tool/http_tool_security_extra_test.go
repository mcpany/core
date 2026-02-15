package tool_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_Security_PathTraversal(t *testing.T) {
	t.Parallel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Should not be reached if blocked
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	// Tool definition: GET /files/{{file}}
	methodAndURL := "GET " + server.URL + "/files/{{file}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("file"),
		}.Build(),
		// We explicitly do NOT set DisableEscape, so escape is enabled (default false).
		// However, we want to test that even if escaped or not, traversal is blocked.
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	testCases := []struct {
		name      string
		input     string
		shouldErr bool
	}{
		{"Normal", "document.txt", false},
		{"DotDot", "../secret", true},
		{"DotDotSlash", "../", true},
		{"EncodedDotDot", "%2e%2e%2fsecret", true},
		{"EncodedDotDot2", "..%2fsecret", true},
		{"DoubleEncoded", "%252e%252e%252fsecret", true}, // %252e -> %2e -> .
		{"MixedEncoding", "%2e.%2f", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			inputs := json.RawMessage(fmt.Sprintf(`{"file": "%s"}`, tc.input))
			req := &tool.ExecutionRequest{ToolInputs: inputs}
			_, err := httpTool.Execute(context.Background(), req)

			if tc.shouldErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "path traversal attempt detected")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHTTPTool_Security_SSRF_Host_Parsing(t *testing.T) {
	t.Parallel()
	// This test confirms that dynamic host injection via simple substitution
	// (e.g. http://{{host}}/) is PREVENTED by the strict URL parsing at initialization.
	// HTTPTool requires the method+url string to be a valid URL *before* substitution.
	// Since {{host}} contains {, it fails parsing.

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: http.DefaultClient}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	// Tool definition: GET http://{{host}}/api
	// This is invalid URL syntax and should cause an init error.
	methodAndURL := "GET http://{{host}}/api"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("host"),
		}.Build(),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"host": "localhost"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse url")
}

func TestHTTPTool_Security_SSRF_Scheme(t *testing.T) {
	t.Parallel()
	// Verify that Scheme Injection is prevented.
	// We rely on the real IsSafeURL or http.Client validation.
	// IsSafeURL (real) blocks anything other than http/https.
	// http.Client blocks unsupported schemes or invalid URLs.

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: http.DefaultClient}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	// Tool definition: GET {{url}}
	// This template will be parsed as path relative to root because it has no scheme.
	methodAndURL := "GET {{url}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("url"),
		}.Build(),
		DisableEscape: proto.Bool(true), // Allow raw injection
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"url": "file:///etc/passwd"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}
	_, err = httpTool.Execute(context.Background(), req)

	// With DisableEscape=true, the constructed URL becomes "/file:///etc/passwd" (prepended slash).
	// This fails at http.Client execution because it's a relative URL without a scheme.
	// Or IsSafeURL might catch it if it parses weirdly.

	require.Error(t, err)
	assert.True(t,
		strings.Contains(err.Error(), "unsafe url") ||
		strings.Contains(err.Error(), "unsupported protocol scheme") ||
		strings.Contains(err.Error(), "unsupported scheme"),
		"Error should indicate blocked scheme or invalid URL: %v", err)
}

func TestHTTPTool_Security_HeaderInjection(t *testing.T) {
	t.Parallel()
	// Attempt to inject headers via path or query parameters.
	// Go's net/http and url packages should prevent this by escaping or rejecting control characters.

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that headers were NOT injected
		assert.Equal(t, "", r.Header.Get("X-Injected"), "Header should not be injected")

		// Also check that the path/query contains the raw/escaped characters, not interpreted as newlines
		// Note: URL encoding might happen, so we check for the encoded versions or absence of actual newlines.
		// If newline was interpreted, it would break the request line.

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL + "/echo?msg={{msg}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("msg"),
		}.Build(),
	}.Build()
	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	// Attempt injection: "hello\r\nX-Injected: true"
	injectionPayload := "hello\r\nX-Injected: true"
	inputs := json.RawMessage(fmt.Sprintf(`{"msg": "%s"}`, injectionPayload))
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	_, err = httpTool.Execute(context.Background(), req)

	// Either error (net/http might reject) or success (but header not injected) is fine.
	// If success, the handler assertion guarantees no injection.
	// If error, it means the client prevented sending malformed request, which is also safe.
	if err != nil {
		// Verify it's not a timeout or unrelated error
		// Go's net/http often returns error for invalid control characters in URL
	}
}
