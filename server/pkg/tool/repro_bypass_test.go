package tool_test

import (
	"context"
	"encoding/json"
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
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_PathTraversal_TripleEncodingBypass(t *testing.T) {
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

	// This test attempts to bypass path traversal checks using triple URL encoding.
	// %2e -> .
	// %252e -> %2e
	// %25252e -> %252e
	//
	// Payload: ../
	// Triple encoded: %25252e%25252e%25252f

	payload := "%25252e%25252e%25252f"

	pathHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	})
	server := httptest.NewServer(pathHandler)
	defer server.Close()

	poolManager := pool.NewManager()
	p, err := pool.New(func(_ context.Context) (*client.HTTPClientWrapper, error) {
		return &client.HTTPClientWrapper{Client: server.Client()}, nil
	}, 1, 1, 1, 0, true)
	require.NoError(t, err)
	poolManager.Register("test-service", p)

	methodAndURL := "GET " + server.URL + "/data/{{param}}"
	mcpTool := v1.Tool_builder{
		UnderlyingMethodFqn: &methodAndURL,
	}.Build()

	// Enable DisableEscape to allow raw injection
	paramMapping := configv1.HttpParameterMapping_builder{
		Schema: configv1.ParameterSchema_builder{
			Name: proto.String("param"),
		}.Build(),
		DisableEscape: proto.Bool(true),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Parameters: []*configv1.HttpParameterMapping{paramMapping},
	}.Build()

	httpTool := tool.NewHTTPTool(mcpTool, poolManager, "test-service", nil, callDef, nil, nil, "")

	inputs := json.RawMessage(`{"param": "` + payload + `"}`)
	req := &tool.ExecutionRequest{ToolInputs: inputs}

	// We expect this to fail if the bug is fixed.
	// If the bug exists, this will succeed (err == nil).
	_, err = httpTool.Execute(context.Background(), req)

	if err == nil {
		t.Fatal("VULNERABILITY DETECTED: Triple encoded path traversal was not detected!")
	} else {
		assert.Contains(t, err.Error(), "path traversal attempt detected")
	}
}
