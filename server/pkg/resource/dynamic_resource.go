// Package resource provides resource management functionality.

package resource

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mcpany/core/server/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// DynamicResource implements the Resource interface for resources that are
// fetched dynamically by executing a tool.
type DynamicResource struct {
	resource *mcp.Resource
	tool     tool.Tool
}

// NewDynamicResource creates a new instance of DynamicResource.
//
// def is the def.
// t is the t.
//
// Returns the result.
// Returns an error if the operation fails.
func NewDynamicResource(def *configv1.ResourceDefinition, t tool.Tool) (*DynamicResource, error) {
	if def == nil {
		return nil, fmt.Errorf("resource definition is nil")
	}
	if t == nil {
		return nil, fmt.Errorf("tool is nil")
	}
	return &DynamicResource{
		resource: &mcp.Resource{
			URI:         def.GetUri(),
			Name:        def.GetName(),
			Title:       def.GetTitle(),
			Description: def.GetDescription(),
			MIMEType:    def.GetMimeType(),
			Size:        def.GetSize(),
		},
		tool: t,
	}, nil
}

// Resource returns the MCP representation of the resource.
//
// Returns the result.
func (r *DynamicResource) Resource() *mcp.Resource {
	return r.resource
}

// Service returns the ID of the service that provides this resource.
//
// Returns the result.
func (r *DynamicResource) Service() string {
	return r.tool.Tool().GetServiceId()
}

// Read executes the associated tool to fetch the resource content.
//
// ctx is the context for the request.
//
// Returns the result.
// Returns an error if the operation fails.
func (r *DynamicResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	// For now, we'll just execute the tool with no inputs.
	// In the future, we may need to pass inputs to the tool.
	result, err := r.tool.Execute(ctx, &tool.ExecutionRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to execute tool for dynamic resource: %w", err)
	}

	// The tool can return a string, a byte slice, or a map[string]interface{}.
	// We need to handle each of these cases.
	switch content := result.(type) {
	case string:
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      r.resource.URI,
					Text:     content,
					MIMEType: r.resource.MIMEType,
				},
			},
		}, nil
	case []byte:
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      r.resource.URI,
					Blob:     content,
					MIMEType: r.resource.MIMEType,
				},
			},
		}, nil
	case map[string]interface{}:
		// If the tool returns a map, we assume it's a JSON object.
		// We'll marshal it to a string and return it as text.
		data, err := json.Marshal(content)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool result to JSON: %w", err)
		}
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      r.resource.URI,
					Text:     string(data),
					MIMEType: r.resource.MIMEType,
				},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported tool result type for dynamic resource: %T", result)
	}
}

// Subscribe is not yet implemented for dynamic resources.
//
// _ is an unused parameter.
//
// Returns an error if the operation fails.
func (r *DynamicResource) Subscribe(_ context.Context) error {
	return fmt.Errorf("subscribing to dynamic resources is not yet implemented")
}
