package tool_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/client"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockHTTPClient for testing
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

func TestOpenAPITool_Execute(t *testing.T) {
	t.Run("successful execution with path and query params", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/users/123", r.URL.Path)
			assert.Equal(t, "test", r.URL.Query().Get("q"))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"123", "name":"test"}`))
		}))
		defer server.Close()

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		toolProto.SetName("getUser")
		parameterDefs := map[string]string{
			"userId": "path",
			"q":      "query",
		}
		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, parameterDefs, "GET", server.URL+"/users/{{userId}}", nil, &configv1.OpenAPICallDefinition{})

		inputs := json.RawMessage(`{"userId": "123", "q": "test"}`)
		req := &tool.ExecutionRequest{ToolInputs: inputs}

		result, err := openAPITool.Execute(context.Background(), req)
		require.NoError(t, err)

		expected := map[string]any{"id": "123", "name": "test"}
		assert.Equal(t, expected, result)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, nil, &configv1.OpenAPICallDefinition{})

		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err := openAPITool.Execute(context.Background(), req)
		assert.Error(t, err)
	})

	t.Run("with authentication", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "my-secret-key", r.Header.Get("X-API-Key"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		mockClient := &mockHTTPClient{
			doFunc: server.Client().Do,
		}

		toolProto := &v1.Tool{}
		apiKeyAuth := &configv1.UpstreamAPIKeyAuth{}
		apiKeyAuth.SetHeaderName("X-API-Key")
		secret := &configv1.SecretValue{}
		secret.SetPlainText("my-secret-key")
		apiKeyAuth.SetApiKey(secret)
		authn := &configv1.UpstreamAuthentication{}
		authn.SetApiKey(apiKeyAuth)
		authenticator, err := auth.NewUpstreamAuthenticator(authn)
		require.NoError(t, err)

		openAPITool := tool.NewOpenAPITool(toolProto, mockClient, nil, "GET", server.URL, authenticator, &configv1.OpenAPICallDefinition{})

		req := &tool.ExecutionRequest{ToolInputs: json.RawMessage(`{}`)}
		_, err = openAPITool.Execute(context.Background(), req)
		assert.NoError(t, err)
	})
}
