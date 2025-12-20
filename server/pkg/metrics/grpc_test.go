// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package metrics

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/grpc/stats"
)

func TestGrpcStatsHandler(t *testing.T) {
	// Initialize the metrics system
	if err := Initialize(); err != nil {
		t.Fatalf("failed to initialize metrics: %v", err)
	}

	h := &GrpcStatsHandler{}

	// Test HandleConn
	h.HandleConn(context.Background(), &stats.ConnBegin{})
	h.HandleConn(context.Background(), &stats.ConnEnd{})

	// Test TagRPC
	if ctx := h.TagRPC(context.Background(), &stats.RPCTagInfo{}); ctx == nil {
		t.Error("TagRPC returned a nil context")
	}

	// Test HandleRPC
	h.HandleRPC(context.Background(), &stats.Begin{})
	h.HandleRPC(context.Background(), &stats.End{})

	// Test TagConn
	if ctx := h.TagConn(context.Background(), &stats.ConnTagInfo{}); ctx == nil {
		t.Error("TagConn returned a nil context")
	}

	// Create a test server
	ts := httptest.NewServer(Handler())
	defer ts.Close()

	// Make a request to the /metrics endpoint
	resp, err := http.Get(ts.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	// Check the response body for the expected metrics
	if !strings.Contains(string(body), "mcpany_grpc_connections_opened_total 1") {
		t.Errorf("Expected metric mcpany_grpc_connections_opened_total not found in response body")
	}
	if !strings.Contains(string(body), "mcpany_grpc_connections_closed_total 1") {
		t.Errorf("Expected metric mcpany_grpc_connections_closed_total not found in response body")
	}

	// Check for RPC metrics
	if !strings.Contains(string(body), "mcpany_grpc_rpc_started_total 1") {
		t.Errorf("Expected metric mcpany_grpc_rpc_started_total not found in response body")
	}
	if !strings.Contains(string(body), "mcpany_grpc_rpc_finished_total 1") {
		t.Errorf("Expected metric mcpany_grpc_rpc_finished_total not found in response body")
	}
}
