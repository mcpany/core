// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"context"
	"strings"
	"testing"

	"github.com/mcpany/core/server/pkg/bus"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

// We reuse mockToolSimple from extra_management_test.go since we are in the same package.

func TestExecuteTool_FuzzyMatching(t *testing.T) {
	// Setup
	b, _ := bus.NewProvider(nil)
	tm := NewManager(b)

	// Add a tool
	toolDef := &v1.Tool{
		Name:      proto.String("get_weather"),
		ServiceId: proto.String("weather_service"),
	}

	// Create mockToolSimple using struct from extra_management_test.go
	mt := &mockToolSimple{
		toolDef: toolDef,
		serviceID: "weather_service",
	}

	err := tm.AddTool(mt)
	if err != nil {
		t.Fatalf("Failed to add tool: %v", err)
	}

	// Test Case 1: Fuzzy Match
	// Typo: "weather_service.get_wether"
	// We expect "Did you mean 'weather_service.get_weather'?"

	req := &ExecutionRequest{
		ToolName: "weather_service.get_wether",
	}

	_, err = tm.ExecuteTool(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for non-existent tool, got nil")
	}

	expectedSuggestion := "weather_service.get_weather"
	if !strings.Contains(err.Error(), expectedSuggestion) {
		t.Errorf("Expected error message to contain suggestion '%s', got: %v", expectedSuggestion, err)
	}

	if !strings.Contains(err.Error(), "Did you mean") {
		t.Errorf("Expected error message to contain 'Did you mean', got: %v", err)
	}

	// Test Case 2: No Match (Too far)
	req = &ExecutionRequest{
		ToolName: "weather_service.completely_wrong",
	}
	_, err = tm.ExecuteTool(context.Background(), req)
	if err == nil {
		t.Fatal("Expected error for non-existent tool, got nil")
	}

	if strings.Contains(err.Error(), "Did you mean") {
		t.Errorf("Did not expect suggestion for wildly different name, got: %v", err)
	}
}
