package tool

import (
	"context"
	"encoding/json"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/types/known/structpb"
)

// LocalTool implements the Tool interface for a tool that is executed as a
// local Go function.
type LocalTool struct {
	tool    *v1.Tool
	handler func(context.Context, *structpb.Struct) (*structpb.Value, error)
}

// NewLocalTool creates a new LocalTool.
func NewLocalTool(tool *v1.Tool, handler func(context.Context, *structpb.Struct) (*structpb.Value, error)) Tool {
	return &LocalTool{
		tool:    tool,
		handler: handler,
	}
}

// Tool returns the protobuf definition of the tool.
func (t *LocalTool) Tool() *v1.Tool {
	return t.tool
}

// Execute runs the tool with the provided context and request, returning
// the result or an error.
func (t *LocalTool) Execute(ctx context.Context, req *ExecutionRequest) (any, error) {
	args, err := t.parseArguments(req.ToolInputs)
	if err != nil {
		return nil, err
	}

	return t.handler(ctx, args)
}

// GetCacheConfig returns the cache configuration for the tool.
func (t *LocalTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

func (t *LocalTool) parseArguments(rawArgs json.RawMessage) (*structpb.Struct, error) {
	var args map[string]interface{}
	if err := json.Unmarshal(rawArgs, &args); err != nil {
		return nil, err
	}

	return structpb.NewStruct(args)
}
