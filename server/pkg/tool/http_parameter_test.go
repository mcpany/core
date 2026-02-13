package tool_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/client"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func ptr[T any](v T) *T {
	return &v
}

func TestHTTPTool_ExplicitParameterLocation(t *testing.T) {
	// Enable local IPs for testing
	os.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")
	defer os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	// 1. Setup Mock Server
	receivedHeaders := make(http.Header)
	receivedCookies := make(map[string]*http.Cookie)
	receivedQuery := make(url.Values)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for k, v := range r.Header {
			receivedHeaders[k] = v
		}
		for _, c := range r.Cookies() {
			receivedCookies[c.Name] = c
		}
		receivedQuery = r.URL.Query()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	// 2. Setup Pool Manager
	poolManager := pool.NewManager()

	// Register Mock Pool
	mockPool := &MockHTTPPool{
		client: server.Client(),
	}
	poolManager.Register("test-service", mockPool)

	// 3. Define Tool using Builders
	callDef := configv1.HttpCallDefinition_builder{
		Id:           proto.String("test_call"),
		EndpointPath: proto.String("/test"),
		Method:       ptr(configv1.HttpCallDefinition_HTTP_METHOD_GET),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("headerParam"),
					Type: ptr(configv1.ParameterType_STRING),
				}.Build(),
				Location: ptr(configv1.ParameterLocation_PARAMETER_LOCATION_HEADER),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("cookieParam"),
					Type: ptr(configv1.ParameterType_STRING),
				}.Build(),
				Location: ptr(configv1.ParameterLocation_PARAMETER_LOCATION_COOKIE),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("queryParam"),
					Type: ptr(configv1.ParameterType_STRING),
				}.Build(),
				Location: ptr(configv1.ParameterLocation_PARAMETER_LOCATION_QUERY),
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{
		Name:                proto.String("test_tool"),
		UnderlyingMethodFqn: proto.String("GET " + server.URL + "/test"),
	}.Build()

	// 4. Create HTTPTool
	httpTool := tool.NewHTTPTool(
		toolDef,
		poolManager,
		"test-service",
		nil, // No auth
		callDef,
		nil, // No resilience
		nil, // No policies
		"test_call",
	)

	// 5. Execute
	inputs := map[string]interface{}{
		"headerParam": "header-value",
		"cookieParam": "cookie-value",
		"queryParam":  "query-value",
	}
	inputsBytes, _ := json.Marshal(inputs)

	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: inputsBytes,
	}

	_, err := httpTool.Execute(context.Background(), req)
	require.NoError(t, err)

	// 6. Verify
	assert.Equal(t, "header-value", receivedHeaders.Get("headerParam"))
	assert.NotNil(t, receivedCookies["cookieParam"])
	if receivedCookies["cookieParam"] != nil {
		assert.Equal(t, "cookie-value", receivedCookies["cookieParam"].Value)
	}
	assert.Equal(t, "query-value", receivedQuery.Get("queryParam"))
}

// MockHTTPPool implements pool.Pool[*client.HTTPClientWrapper]
type MockHTTPPool struct {
	client *http.Client
}

func (p *MockHTTPPool) Get(ctx context.Context) (*client.HTTPClientWrapper, error) {
	return &client.HTTPClientWrapper{Client: p.client}, nil
}

func (p *MockHTTPPool) Put(c *client.HTTPClientWrapper) {
}

func (p *MockHTTPPool) Close() error {
    return nil
}

func (p *MockHTTPPool) Len() int {
    return 0
}
