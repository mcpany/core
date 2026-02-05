package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	httppkg "github.com/mcpany/core/server/pkg/upstream/http"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestPathEncodingBug(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Create a test server that checks the RequestURI
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We expect the path to be exactly "/users/user%2F1"
		// If the bug exists, it will likely be "/users/user/1" (double slash, or just decoded)

		expectedURI := "/users/user%2F1"
		if r.RequestURI != expectedURI {
			t.Errorf("Expected RequestURI='%s', got '%s'", expectedURI, r.RequestURI)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := httppkg.NewUpstream(pm)

	// Configure a tool with a path parameter containing encoded slash (%2F)
	configJSON := `{
		"name": "encoding-service",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/users/user%2F1"
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	// Execute the tool
	_, err = registeredTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: []byte("{}"),
	})
	// If status code is 400, Execute might return error or not depending on implementation.
	// Assuming tool returns error on 400.
	require.NoError(t, err)
}

func TestPathEncoding_SlashInParameter_WithTrailingSlash(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Create a test server that checks the path
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We expect the path to contain "foo%2Fbar" (encoded slash)
		// r.URL.Path is decoded by Go's http server, so it will show "foo/bar"
		// r.URL.RawPath should show "foo%2Fbar"

		// However, we can check RequestURI or RawPath
		t.Logf("Received RequestURI: %s", r.RequestURI)
		t.Logf("Received Path: %s", r.URL.Path)
		t.Logf("Received RawPath: %s", r.URL.RawPath)

		// We expect encoded slash.
		// If the client sent /test/foo/bar, then RawPath might be empty or /test/foo/bar
		// If the client sent /test/foo%2Fbar, then RawPath should be /test/foo%2Fbar

		// Note: httptest server might behavior differently regarding RawPath populating.
		// But checking RequestURI is reliable.

		// We expect RequestURI to contain "foo%2Fbar"
		// And we expect it NOT to be just "foo/bar" (unless encoded)

		expectedSnippet := "foo%2Fbar"
		if !strings.Contains(r.RequestURI, expectedSnippet) {
			t.Errorf("Expected RequestURI to contain '%s', got '%s'", expectedSnippet, r.RequestURI)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("{}"))
	}))
	defer server.Close()

	pm := pool.NewManager()
	tm := tool.NewManager(nil)
	upstream := httppkg.NewUpstream(pm)

	// Configure a tool with a path parameter and a TRAILING SLASH in endpoint_path
	configJSON := `{
		"name": "encoding-test-service",
		"http_service": {
			"address": "` + server.URL + `",
			"tools": [{"name": "test-op", "call_id": "test-op-call"}],
			"calls": {
				"test-op-call": {
					"id": "test-op-call",
					"method": "HTTP_METHOD_GET",
					"endpoint_path": "/test/{{val}}/",
					"parameters": [
						{
							"schema": {
								"name": "val",
								"type": "STRING",
								"is_required": true
							}
						}
					]
				}
			}
		}
	}`
	serviceConfig := configv1.UpstreamServiceConfig_builder{}.Build()
	require.NoError(t, protojson.Unmarshal([]byte(configJSON), serviceConfig))

	serviceID, _, _, err := upstream.Register(context.Background(), serviceConfig, tm, nil, nil, false)
	require.NoError(t, err)

	sanitizedToolName, _ := util.SanitizeToolName("test-op")
	toolID := serviceID + "." + sanitizedToolName
	registeredTool, ok := tm.GetTool(toolID)
	require.True(t, ok)

	// Execute the tool with input containing '/'
	_, err = registeredTool.Execute(context.Background(), &tool.ExecutionRequest{
		ToolName:   toolID,
		ToolInputs: []byte(`{"val": "foo/bar"}`),
	})
	require.NoError(t, err)
}
