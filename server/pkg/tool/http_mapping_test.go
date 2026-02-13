package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_ExplicitMapping(t *testing.T) {
	// Setup
	callID := "test_call"
	serviceID := "test_service"
	poolManager := pool.NewManager()

	// Config with explicit Header, Cookie, and Query mapping
	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	headerLoc := configv1.ParameterLocation_HEADER
	cookieLoc := configv1.ParameterLocation_COOKIE
	queryLoc := configv1.ParameterLocation_QUERY
	stringType := configv1.ParameterType_STRING

	callDef := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		Method:       &method,
		EndpointPath: proto.String("/test"),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name:       proto.String("headerParam"),
					Type:       &stringType,
					IsRequired: proto.Bool(true),
				}.Build(),
				Location: &headerLoc,
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name:       proto.String("cookieParam"),
					Type:       &stringType,
					IsRequired: proto.Bool(true),
				}.Build(),
				Location: &cookieLoc,
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name:       proto.String("queryParam"),
					Type:       &stringType,
					IsRequired: proto.Bool(true),
				}.Build(),
				Location: &queryLoc,
			}.Build(),
		},
	}.Build()

	toolDef := v1.Tool_builder{
		Name:                proto.String("test_tool"),
		UnderlyingMethodFqn: proto.String("GET http://example.com/test"),
	}.Build()

	// Create tool
	httpTool := NewHTTPTool(toolDef, poolManager, serviceID, nil, callDef, nil, nil, callID)

	// Create request
	inputs := map[string]interface{}{
		"headerParam": "headerVal",
		"cookieParam": "cookieVal",
		"queryParam":  "queryVal",
	}
	inputBytes, _ := fastJSON.Marshal(inputs)
	req := &ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: inputBytes,
	}

	ctx := context.Background()

	// Call prepareInputsAndURL
	processedInputs, urlStr, headerParams, cookieParams, queryParams, _, err := httpTool.prepareInputsAndURL(ctx, req)

	assert.NoError(t, err)
	assert.Equal(t, "http://example.com/test", urlStr) // Base URL + path (no params in path)

	assert.Equal(t, "headerVal", headerParams["headerParam"])
	assert.Equal(t, "cookieVal", cookieParams["cookieParam"])
	assert.Equal(t, "queryVal", queryParams["queryParam"])

	// Check that processedInputs is empty (all consumed)
	assert.Empty(t, processedInputs)

	// Now test createHTTPRequest
	httpReq, err := httpTool.createHTTPRequest(ctx, urlStr, nil, "", processedInputs, headerParams, cookieParams, queryParams)
	assert.NoError(t, err)

	assert.Equal(t, "headerVal", httpReq.Header.Get("headerParam"))

	cookie, err := httpReq.Cookie("cookieParam")
	assert.NoError(t, err)
	assert.Equal(t, "cookieVal", cookie.Value)

	assert.Equal(t, "queryVal", httpReq.URL.Query().Get("queryParam"))
}
