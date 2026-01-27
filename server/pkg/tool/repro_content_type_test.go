package tool

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHTTPTool_PrepareBody_Template_ContentType_Missing(t *testing.T) {
	// This test reproduces the issue where Content-Type is missing when using InputTransformer template.
	urlStr := "http://example.com/api"
	fqn := "POST " + urlStr
	toolProto := &pb.Tool{
		Name:                proto.String("test-tool"),
		UnderlyingMethodFqn: proto.String(fqn),
	}

	callDef := &configv1.HttpCallDefinition{
		Method:       configv1.HttpCallDefinition_HTTP_METHOD_POST.Enum(),
		EndpointPath: proto.String(urlStr),
		InputTransformer: &configv1.InputTransformer{
			Template: proto.String(`{"key": "{{arg}}"}`),
		},
	}

	pm := pool.NewManager()
	httpTool := NewHTTPTool(toolProto, pm, "service-id", nil, callDef, nil, nil, "call-id")

	ctx := context.Background()
	inputs := map[string]any{"arg": "value"}

	// We expect contentType to be "application/json" if the template output is valid JSON.
	// But current implementation returns empty string.
	_, contentType, err := httpTool.prepareBody(ctx, inputs, "POST", "test-tool", nil, false)
	assert.NoError(t, err)

	// This assertion will fail if the bug exists (and we expect it to be "application/json")
	// If we assert it is "", it passes, proving the current behavior.
	// We want to fix it, so we should expect "application/json".
	assert.Equal(t, "application/json", contentType, "Content-Type should be application/json for JSON template output")
}
