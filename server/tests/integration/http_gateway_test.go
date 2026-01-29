package integration

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestHTTPGateway_RegisterService(t *testing.T) {
	server := StartInProcessMCPANYServer(t, "HTTPGatewayTest")
	defer server.CleanupFunc()

	serviceID := "http-gateway-test-service"
	baseURL := "http://example.com"
	operationID := "test-op"
	endpointPath := "/test"

	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	callID := "call-" + operationID
	callDef := configv1.HttpCallDefinition_builder{
		Id:           &callID,
		EndpointPath: &endpointPath,
		Method:       &method,
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   &operationID,
		CallId: &callID,
	}.Build()

	upstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{
		Name: &serviceID,
		HttpService: configv1.HttpUpstreamService_builder{
			Address: &baseURL,
			Tools:   []*configv1.ToolDefinition{toolDef},
			Calls:   map[string]*configv1.HttpCallDefinition{callID: callDef},
		}.Build(),
	}
	config := upstreamServiceConfigBuilder.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	// Marshal request to JSON
	jsonBytes, err := protojson.Marshal(req)
	require.NoError(t, err)

	// Send request to HTTP Gateway
	gatewayURL := server.JSONRPCEndpoint + "/v1/services/register"
	httpReq, err := http.NewRequest(http.MethodPost, gatewayURL, bytes.NewBuffer(jsonBytes))
	require.NoError(t, err)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify registration via ListServices (also via Gateway!)
	listURL := server.JSONRPCEndpoint + "/v1/services"
	listResp, err := client.Get(listURL)
	require.NoError(t, err)
	defer func() { _ = listResp.Body.Close() }()
	require.Equal(t, http.StatusOK, listResp.StatusCode)

	// Unmarshal response
	var listServicesResp apiv1.ListServicesResponse
	// We need to read body into bytes first
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(listResp.Body)
	require.NoError(t, err)

	err = protojson.Unmarshal(buf.Bytes(), &listServicesResp)
	require.NoError(t, err)

	found := false
	for _, s := range listServicesResp.GetServices() {
		if s.GetName() == serviceID {
			found = true
			break
		}
	}
	require.True(t, found, "Registered service not found via ListServices API")
}
