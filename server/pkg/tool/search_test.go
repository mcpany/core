// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool_test

import (
	"context"
	"strings"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// MockEmbeddingProvider is a simple mock that returns deterministic embeddings.
// It maps the first letter of the input to a dimension in the vector.
type MockEmbeddingProvider struct{}

func (m *MockEmbeddingProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	// Simple deterministic embedding:
	// Use 5 dimensions.
	// Assign 1.0 to a dimension based on keywords.
	vec := make([]float32, 5)
	lower := strings.ToLower(text)
	if strings.Contains(lower, "weather") {
		vec[0] = 1.0
	}
	if strings.Contains(lower, "docker") {
		vec[1] = 1.0
	}
	if strings.Contains(lower, "stock") {
		vec[2] = 1.0
	}
	return vec, nil
}

// MockTool implements tool.Tool interface for testing.
type MockTool struct {
	def *v1.Tool
}

func (m *MockTool) Tool() *v1.Tool {
	return m.def
}

func (m *MockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return "ok", nil
}

func (m *MockTool) MCPTool() *mcp.Tool {
	return &mcp.Tool{Name: m.def.Name}
}

func TestSemanticSearch(t *testing.T) {
	// Initialize Manager
	bp, _ := bus.NewProvider(nil)
	tm := tool.NewManager(bp)
	tm.SetEmbeddingProvider(&MockEmbeddingProvider{})

	// Add Tools
	tools := []*v1.Tool{
		{
			Name:        "get_weather",
			Description: "Get current weather for a location",
			ServiceId:   "weather_service",
		},
		{
			Name:        "list_containers",
			Description: "List running docker containers",
			ServiceId:   "docker_service",
		},
		{
			Name:        "get_stock_price",
			Description: "Get stock price for a symbol",
			ServiceId:   "finance_service",
		},
	}

	for _, def := range tools {
		if err := tm.AddTool(&MockTool{def: def}); err != nil {
			t.Fatalf("AddTool failed for %s: %v", def.Name, err)
		}
	}

	// Allow indexing to complete (it's async)
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name     string
		query    string
		wantTool string
	}{
		{
			name:     "Search for weather",
			query:    "I want to check the weather",
			wantTool: "weather_service.get_weather",
		},
		{
			name:     "Search for docker",
			query:    "manage docker containers",
			wantTool: "docker_service.list_containers",
		},
		// Note: stock is not tested for mismatch because simple dot product logic might conflict if vectors are sparse
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, scores, err := tm.SearchTools(context.Background(), tt.query, 1)
			if err != nil {
				t.Fatalf("SearchTools error: %v", err)
			}
			if len(results) == 0 {
				t.Fatalf("SearchTools returned no results")
			}

			// We only get 1 result for now as per current implementation limitation
			// Construct the expected name which includes the service ID namespace
			gotName := results[0].Tool().GetServiceId() + "." + results[0].Tool().GetName()
			if gotName != tt.wantTool {
				t.Errorf("SearchTools got %s, want %s (score: %f)", gotName, tt.wantTool, scores[0])
			}
		})
	}
}

func TestSemanticSearch_NoProvider(t *testing.T) {
	bp, _ := bus.NewProvider(nil)
	tm := tool.NewManager(bp)
	// No provider set

	_, _, err := tm.SearchTools(context.Background(), "test", 1)
	if err == nil {
		t.Error("SearchTools should fail without embedding provider")
	}
}

func TestSemanticSearch_NoResults(t *testing.T) {
	bp, _ := bus.NewProvider(nil)
	tm := tool.NewManager(bp)
	tm.SetEmbeddingProvider(&MockEmbeddingProvider{})

	// Index empty
	results, _, err := tm.SearchTools(context.Background(), "weather", 1)
	if err != nil {
		t.Fatalf("SearchTools error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("SearchTools returned results for empty index")
	}
}

// TestAddTool_Indexing ensures indexing happens
func TestAddTool_Indexing(t *testing.T) {
	bp, _ := bus.NewProvider(nil)
	tm := tool.NewManager(bp)
	tm.SetEmbeddingProvider(&MockEmbeddingProvider{})

	def := &v1.Tool{
		Name:        "test_tool",
		Description: "A test tool description with keyword weather",
		ServiceId:   "test_service",
	}

	if err := tm.AddTool(&MockTool{def: def}); err != nil {
		t.Fatalf("AddTool failed: %v", err)
	}

	// Give time for async indexing
	time.Sleep(50 * time.Millisecond)

	results, _, err := tm.SearchTools(context.Background(), "weather", 1)
	if err != nil {
		t.Fatalf("SearchTools error: %v", err)
	}
	if len(results) == 0 {
		t.Error("Tool was not indexed")
	}
}

// Helpers
func (m *MockTool) Validate(configv1.ToolDefinition_MergeStrategy) error { return nil }
