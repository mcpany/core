package security

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/middleware"
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

type contextKey string

func TestGRPCTool_Execute_ScopeEnforcement(t *testing.T) {
	t.Parallel()

	scopesConfig := middleware.ScopesConfig{
		Roles: map[string][]string{
			"default": {"grpc-tool-secure"},
			"unprivileged": {"some-other-tool"},
		},
	}
	scopesMiddleware := middleware.NewScopesMiddleware(scopesConfig)

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
		nil,
		"test-call-id",
	)

	// Test 1: Agent role without correct scope
	ctxDenied := context.WithValue(context.Background(), middleware.AgentRoleKeyForTest(), "unprivileged")
	reqDenied := &tool.ExecutionRequest{
		ToolName:   "grpc-tool-secure",
		ToolInputs: []byte(`{"data": "exploit payload"}`),
	}

	_, err := scopesMiddleware.Execute(ctxDenied, reqDenied, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return grpcTool.Execute(ctx, req)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied: tool 'grpc-tool-secure' is outside granted scopes")

	// Test 2: Unknown role
	ctxUnknown := context.WithValue(context.Background(), middleware.AgentRoleKeyForTest(), "unknown")
	reqUnknown := &tool.ExecutionRequest{
		ToolName:   "grpc-tool-secure",
		ToolInputs: []byte(`{"data": "exploit payload"}`),
	}

	_, err = scopesMiddleware.Execute(ctxUnknown, reqUnknown, func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return grpcTool.Execute(ctx, req)
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access denied: no scope configuration for role")
}
