package tool

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestNewHTTPTool_SpaceInURL(t *testing.T) {
	// This test specifically verifies that NewHTTPTool correctly parses URLs with spaces.
	// This was a bug where strings.Fields splits URL incorrectly.

	method := "GET"
	// URL with a space (e.g. from an invalid query param)
	urlWithSpace := "http://example.com/api?q=hello world"
	fqn := fmt.Sprintf("%s %s", method, urlWithSpace)

	toolProto := pb.Tool_builder{
		Name:                proto.String("test-tool"),
		UnderlyingMethodFqn: proto.String(fqn),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
		EndpointPath: proto.String(urlWithSpace),
	}.Build()

	pm := pool.NewManager()
	httpTool := NewHTTPTool(toolProto, pm, "service-id", nil, callDef, nil, nil, "call-id")

	// If initialization failed, initError would be set.
	// We verify this by attempting to Execute.
	// If initialization succeeded, Execute will fail with "no http pool found" (since we didn't register a pool).
	// If initialization failed (regression), it would return the "invalid http tool definition" initError.

	req := &ExecutionRequest{
		ToolName: "test-tool",
	}

	_, err := httpTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no http pool found", "Expected execution to proceed past initialization")
	assert.NotContains(t, err.Error(), "invalid http tool definition", "Initialization should not fail due to space in URL")
}

func TestNewHTTPTool_InvalidFormat(t *testing.T) {
	// Test the case where FQN is indeed invalid (no space)
	fqn := "INVALID_FQN"
	toolProto := pb.Tool_builder{
		Name:                proto.String("test-tool"),
		UnderlyingMethodFqn: proto.String(fqn),
	}.Build()
	callDef := configv1.HttpCallDefinition_builder{}.Build()

	pm := pool.NewManager()
	httpTool := NewHTTPTool(toolProto, pm, "service-id", nil, callDef, nil, nil, "call-id")

	req := &ExecutionRequest{
		ToolName: "test-tool",
	}

	_, err := httpTool.Execute(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http tool definition")
}

func TestHTTPTool_PrepareBody_Template(t *testing.T) {
	// Test prepareBody with template
	urlStr := "http://example.com/api"
	fqn := "POST " + urlStr
	toolProto := pb.Tool_builder{
		Name:                proto.String("test-tool"),
		UnderlyingMethodFqn: proto.String(fqn),
	}.Build()

	callDef := configv1.HttpCallDefinition_builder{
		Method:       configv1.HttpCallDefinition_HTTP_METHOD_POST.Enum(),
		EndpointPath: proto.String(urlStr),
		InputTransformer: configv1.InputTransformer_builder{
			Template: proto.String(`{"key": "{{arg}}"}`),
		}.Build(),
	}.Build()

	pm := pool.NewManager()
	httpTool := NewHTTPTool(toolProto, pm, "service-id", nil, callDef, nil, nil, "call-id")

	// Since we are in the same package, we can directly call private methods like prepareBody.
	ctx := context.Background()
	inputs := map[string]any{"arg": "value"}

	bodyReader, contentType, err := httpTool.prepareBody(ctx, inputs, "POST", "test-tool", nil, false)
	assert.NoError(t, err)
	// When using a template, contentType is not explicitly set by default logic unless detected.
	assert.Equal(t, "", contentType)

	// Verify body content
	buf := new(bytes.Buffer)
	buf.ReadFrom(bodyReader)
	assert.Equal(t, `{"key": "value"}`, buf.String())
}
