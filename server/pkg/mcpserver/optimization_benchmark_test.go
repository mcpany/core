// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Mock resource implementation
type mockBenchResource struct {
	uri      string
	resource *mcp.Resource
}

func newMockBenchResource(uri string) *mockBenchResource {
	return &mockBenchResource{
		uri:      uri,
		resource: &mcp.Resource{URI: uri},
	}
}

func (r *mockBenchResource) Resource() *mcp.Resource {
	return r.resource
}

func (r *mockBenchResource) Service() string { return "service" }
func (r *mockBenchResource) Read(_ context.Context) (*mcp.ReadResourceResult, error) { return nil, nil }
func (r *mockBenchResource) Subscribe(_ context.Context) error { return nil }

func BenchmarkListResources_Overhead(b *testing.B) {
	rm := resource.NewManager()
	// Add 1000 resources
	for i := 0; i < 1000; i++ {
		rm.AddResource(newMockBenchResource(fmt.Sprintf("res://%d", i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate middleware logic
		managedResources := rm.ListResources()
		refreshedResources := make([]*mcp.Resource, 0, len(managedResources))
		for _, resourceInstance := range managedResources {
			if res := resourceInstance.Resource(); res != nil {
				refreshedResources = append(refreshedResources, res)
			}
		}
	}
}

func BenchmarkListResources_Optimized(b *testing.B) {
	rm := resource.NewManager()
	// Add 1000 resources
	for i := 0; i < 1000; i++ {
		rm.AddResource(newMockBenchResource(fmt.Sprintf("res://%d", i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use optimized method
		_ = rm.ListMCPResources()
	}
}

// Mock prompt implementation
type mockBenchPrompt struct {
	name   string
	prompt *mcp.Prompt
}

func newMockBenchPrompt(name string) *mockBenchPrompt {
	return &mockBenchPrompt{
		name:   name,
		prompt: &mcp.Prompt{Name: name},
	}
}

func (p *mockBenchPrompt) Prompt() *mcp.Prompt {
	return p.prompt
}

func (p *mockBenchPrompt) Service() string { return "service" }
func (p *mockBenchPrompt) Get(_ context.Context, _ json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, nil
}

func BenchmarkListPrompts_Overhead(b *testing.B) {
	pm := prompt.NewManager()
	// Add 1000 prompts
	for i := 0; i < 1000; i++ {
		pm.AddPrompt(newMockBenchPrompt(fmt.Sprintf("prompt-%d", i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate middleware logic
		managedPrompts := pm.ListPrompts()
		refreshedPrompts := make([]*mcp.Prompt, 0, len(managedPrompts))
		for _, promptInstance := range managedPrompts {
			if p := promptInstance.Prompt(); p != nil {
				refreshedPrompts = append(refreshedPrompts, p)
			}
		}
	}
}

func BenchmarkListPrompts_Optimized(b *testing.B) {
	pm := prompt.NewManager()
	// Add 1000 prompts
	for i := 0; i < 1000; i++ {
		pm.AddPrompt(newMockBenchPrompt(fmt.Sprintf("prompt-%d", i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Use optimized method
		_ = pm.ListMCPPrompts()
	}
}
