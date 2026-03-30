package security

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestScopesMiddleware_Exploit(t *testing.T) {
	config := middleware.ScopesConfig{
		Roles: map[string][]string{
			"default": {"fs:read"},
		},
	}
	m := middleware.NewScopesMiddleware(config)

	mockNext := func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
		return "success", nil
	}

	ctx := middleware.WithAgentRole(context.Background(), "default")

	// This should be allowed
	req := &tool.ExecutionRequest{ToolName: "fs:read"}
	_, err := m.Execute(ctx, req, mockNext)
	assert.NoError(t, err)

	// This should be blocked by our new granular scope logic
	reqExploit := &tool.ExecutionRequest{ToolName: "fs:read_foo"}
	_, errExploit := m.Execute(ctx, reqExploit, mockNext)
	if assert.Error(t, errExploit) {
		assert.Contains(t, errExploit.Error(), "access denied")
		s, ok := status.FromError(errExploit)
		assert.True(t, ok, "Expected gRPC status error")
		assert.Equal(t, codes.PermissionDenied, s.Code(), "Expected PermissionDenied code (maps to 403)")
	}
}
