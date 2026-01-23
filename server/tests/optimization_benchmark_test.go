package tests

import (
	"context"
	"testing"
	"fmt"

	"github.com/mcpany/core/server/pkg/resource"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MockResource for benchmarking
type MockResource struct {
	name    string
	service string
}

func (r *MockResource) Resource() *mcp.Resource {
	return &mcp.Resource{
		Name: r.name,
		URI:  "resource://" + r.name,
	}
}
func (r *MockResource) Service() string { return r.service }
func (r *MockResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return nil, nil
}
func (r *MockResource) Subscribe(ctx context.Context) error { return nil }
func (r *MockResource) MCPResource() *mcp.Resource { return r.Resource() }

// BenchmarkListResources compares the performance of ListResources (converting) vs ListMCPResources (cached)
func BenchmarkListResources(b *testing.B) {
	mgr := resource.NewManager()
	// Add 1000 resources
	for i := 0; i < 1000; i++ {
		mgr.AddResource(&MockResource{name: fmt.Sprintf("res%d", i), service: "svc"})
	}

	b.Run("ListResources_Legacy", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			// Simulate what the legacy code did: get resources and convert to MCP
			resources := mgr.ListResources()
			mcpResources := make([]*mcp.Resource, 0, len(resources))
			for _, r := range resources {
				mcpResources = append(mcpResources, r.Resource())
			}
		}
	})

	b.Run("ListMCPResources_Cached", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = mgr.ListMCPResources()
		}
	})
}
