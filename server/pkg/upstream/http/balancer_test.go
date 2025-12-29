// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/balancer"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/tool"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"google.golang.org/protobuf/proto"
)

func TestLoadBalancing(t *testing.T) {
	// Create two mock servers
	server1Hits := 0
	server2Hits := 0

	s1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server1Hits++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"server": 1}`))
	}))
	defer s1.Close()

	s2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server2Hits++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"server": 2}`))
	}))
	defer s2.Close()

	addresses := []string{s1.URL, s2.URL}
	lb := balancer.NewRoundRobinBalancer(addresses)

	// Configure HTTP Tool
	serviceID := "test-service"
	callID := "test-call"
	poolManager := pool.NewManager()

	// Register pool (using first address as primary for pool creation)
	poolConfig := &configv1.UpstreamServiceConfig{}
	httpPool, _ := NewHTTPPool(10, 10, time.Second, poolConfig)
	poolManager.Register(serviceID, httpPool)

	toolProto := pb.Tool_builder{
		Name:                proto.String("test_tool"),
		UnderlyingMethodFqn: proto.String("GET " + s1.URL + "/test"), // This URL will be overridden by balancer
	}.Build()

	method := configv1.HttpCallDefinition_HTTP_METHOD_GET
	callDef := &configv1.HttpCallDefinition{
		Method:       &method,
		EndpointPath: proto.String("/test"),
	}

	httpTool := tool.NewHTTPTool(toolProto, poolManager, serviceID, nil, callDef, nil, nil, callID)
	httpTool.SetBalancer(lb)

	// Execute multiple times
	ctx := context.Background()
	req := &tool.ExecutionRequest{
		ToolName:   "test_tool",
		ToolInputs: json.RawMessage(`{}`),
	}

	for i := 0; i < 10; i++ {
		_, err := httpTool.Execute(ctx, req)
		if err != nil {
			t.Fatalf("Execute failed: %v", err)
		}
	}

	// Verify distribution
	if server1Hits != 5 || server2Hits != 5 {
		t.Errorf("Expected 5 hits each, got server1=%d, server2=%d", server1Hits, server2Hits)
	}
}
