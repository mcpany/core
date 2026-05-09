package security

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type mockMethodDescriptor struct {
	protoreflect.MethodDescriptor
}

func (m *mockMethodDescriptor) Input() protoreflect.MessageDescriptor {
	return nil // sufficient for this test if unmarshalling doesn't panic on nil before policy check
}

func TestGRPCTool_Execute_ScopeEnforcement(t *testing.T) {
	t.Parallel()

	// 1. Set up a call policy that DENIES everything to simulate a granular scope failure
	// We want to test that the policy is evaluated and blocks execution before reaching the network.
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
		Name: proto.String("grpc-tool-secure"),
	}.Build()

	methodDesc := &mockMethodDescriptor{}
	pm := pool.NewManager()
	callDef := &configv1.GrpcCallDefinition{}

	grpcTool := tool.NewGRPCTool(
		toolProto,
		pm,
		"grpc-service",
		methodDesc,
		callDef,
		nil,
		[]*configv1.CallPolicy{policy},
		"test-call-id",
	)

	// 2. Execute the tool with arbitrary input
	req := &tool.ExecutionRequest{
		ToolName:   "grpc-tool-secure",
		ToolInputs: []byte(`{"data": "exploit payload"}`),
	}

	_, err := grpcTool.Execute(context.Background(), req)

	// 3. Verify that execution is blocked by the policy (403 Forbidden semantics)
	assert.Error(t, err)
	if err != nil {
		assert.Contains(t, err.Error(), "tool execution blocked by policy")
	}
}
