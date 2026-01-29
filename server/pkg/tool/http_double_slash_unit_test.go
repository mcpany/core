package tool_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHTTPTool_Execute_DoubleSlashPreservation(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Handler that asserts the path received
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We expect the path to be //foo
		assert.Equal(t, "//foo", r.URL.Path)
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

	// Construct the FQN with double slash
	// server.URL includes scheme (http://127.0.0.1:port)
	// We append //foo
	// Note: We need to be careful not to create http://127.0.0.1:port//foo if server.URL doesn't end in slash?
	// server.URL does NOT end in slash.
	// So http://host:port//foo
	methodAndURL := "GET " + server.URL + "//foo"

	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")

	req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}

func TestHTTPTool_Execute_DoubleSlashRootPreservation(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// Handler that asserts the path received
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// We expect the path to be //
		assert.Equal(t, "//", r.URL.Path)
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

	// Construct the FQN with double slash at root
	methodAndURL := "GET " + server.URL + "//"

	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, &configv1.HttpCallDefinition{}, nil, nil, "")

	req := &tool.ExecutionRequest{ToolInputs: []byte("{}")}
	_, err = httpTool.Execute(context.Background(), req)
	require.NoError(t, err)
}
