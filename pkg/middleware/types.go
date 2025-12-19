package middleware

import (
	"context"
	"github.com/mcpany/core/pkg/tool"
)

// Middleware is the interface for tool execution middleware.
type Middleware interface {
	Execute(ctx context.Context, req *tool.ExecutionRequest, next tool.ExecutionFunc) (any, error)
}
