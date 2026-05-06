package security

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestMCPTool_Execute_ScopeEnforcement(t *testing.T) {
	t.Parallel()

	// Set up a call policy that DENIES everything to simulate a granular scope failure
	policy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_DENY.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				Action:    configv1.CallPolicy_DENY.Enum(),
				NameRegex: proto.String(".*"),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("mcp-tool-secure"),
	}.Build()




	mcpTool := tool.NewMCPTool(
		toolProto,
		nil, // No client needed, should fail before network
		&configv1.MCPCallDefinition{},
		[]*configv1.CallPolicy{policy},
		"test-call-id",
	)

	req := &tool.ExecutionRequest{
		ToolName:   "mcp-tool-secure",
		ToolInputs: []byte(`{"data": "exploit payload"}`),
	}

	_, err := mcpTool.Execute(context.Background(), req)

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "tool execution blocked by policy")
	}
}

func TestOpenAPITool_Execute_ScopeEnforcement(t *testing.T) {
	t.Parallel()

	// Set up a call policy that DENIES everything to simulate a granular scope failure
	policy := configv1.CallPolicy_builder{
		DefaultAction: configv1.CallPolicy_DENY.Enum(),
		Rules: []*configv1.CallPolicyRule{
			configv1.CallPolicyRule_builder{
				Action:    configv1.CallPolicy_DENY.Enum(),
				NameRegex: proto.String(".*"),
			}.Build(),
		},
	}.Build()

	toolProto := pb.Tool_builder{
		Name: proto.String("openapi-tool-secure"),
	}.Build()




	openapiTool := tool.NewOpenAPITool(
		toolProto,
		nil, // No client needed, should fail before network
		nil,
		"POST",
		"http://example.com",
		nil,
		&configv1.OpenAPICallDefinition{},
		[]*configv1.CallPolicy{policy},
		"test-call-id",
	)

	req := &tool.ExecutionRequest{
		ToolName:   "openapi-tool-secure",
		ToolInputs: []byte(`{"data": "exploit payload"}`),
	}

	_, err := openapiTool.Execute(context.Background(), req)

	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "tool execution blocked by policy")
	}
}
