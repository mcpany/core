// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"math"
	"net/http"
	"time"

	config_v1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/topology"
	"google.golang.org/protobuf/proto"
)

// handleDebugSeedState seeds the application state with a standard test dataset.
// It registers a few services and seeds traffic data.
func (a *Application) handleDebugSeedState() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if a.ServiceRegistry == nil {
			http.Error(w, "Service Registry not initialized", http.StatusServiceUnavailable)
			return
		}

		ctx := r.Context()

		// Helper to register service safely
		register := func(cfg *config_v1.UpstreamServiceConfig) {
			if _, _, _, err := a.ServiceRegistry.RegisterService(ctx, cfg); err != nil {
				// Log but don't fail
				_ = err
			}
		}

		// 1. Echo Service (Base)
		register(config_v1.UpstreamServiceConfig_builder{
			Name: proto.String("echo-service"),
			McpService: config_v1.McpUpstreamService_builder{
				StdioConnection: config_v1.McpStdioConnection_builder{
					Command: proto.String("echo"),
					Args:    []string{"hello"},
				}.Build(),
				ToolAutoDiscovery: proto.Bool(true),
			}.Build(),
		}.Build())

		// 2. Filesystem Service
		register(config_v1.UpstreamServiceConfig_builder{
			Name: proto.String("filesystem-service"),
			McpService: config_v1.McpUpstreamService_builder{
				StdioConnection: config_v1.McpStdioConnection_builder{
					Command: proto.String("ls"),
					Args:    []string{"-la"},
				}.Build(),
				ToolAutoDiscovery: proto.Bool(true),
			}.Build(),
		}.Build())

		// 3. Payment Gateway (Simulated for E2E)
		callID := "process_payment_call"
		register(config_v1.UpstreamServiceConfig_builder{
			Name: proto.String("Payment Gateway"),
			McpService: config_v1.McpUpstreamService_builder{
				StdioConnection: config_v1.McpStdioConnection_builder{
					Command: proto.String("echo"),
					Args:    []string{"payment processed"},
				}.Build(),
				Tools: []*config_v1.ToolDefinition{
					config_v1.ToolDefinition_builder{
						Name:        proto.String("process_payment"),
						Description: proto.String("Process a payment"),
						CallId:      &callID,
					}.Build(),
				},
				Calls: map[string]*config_v1.MCPCallDefinition{
					callID: config_v1.MCPCallDefinition_builder{
						Id: &callID,
					}.Build(),
				},
			}.Build(),
		}.Build())

		// 4. User Service (Simulated)
		callIDUser := "get_user_call"
		register(config_v1.UpstreamServiceConfig_builder{
			Name: proto.String("User Service"),
			McpService: config_v1.McpUpstreamService_builder{
				StdioConnection: config_v1.McpStdioConnection_builder{
					Command: proto.String("echo"),
					Args:    []string{"user data"},
				}.Build(),
				Tools: []*config_v1.ToolDefinition{
					config_v1.ToolDefinition_builder{
						Name:        proto.String("get_user"),
						Description: proto.String("Get user details"),
						CallId:      &callIDUser,
					}.Build(),
				},
				Calls: map[string]*config_v1.MCPCallDefinition{
					callIDUser: config_v1.MCPCallDefinition_builder{
						Id: &callIDUser,
					}.Build(),
				},
			}.Build(),
		}.Build())

		// 5. Math Service (Simulated)
		callIDMath := "calculator_call"
		register(config_v1.UpstreamServiceConfig_builder{
			Name: proto.String("Math"),
			McpService: config_v1.McpUpstreamService_builder{
				StdioConnection: config_v1.McpStdioConnection_builder{
					Command: proto.String("echo"),
					Args:    []string{"42"},
				}.Build(),
				Tools: []*config_v1.ToolDefinition{
					config_v1.ToolDefinition_builder{
						Name:        proto.String("calculator"),
						Description: proto.String("Calculate stuff"),
						CallId:      &callIDMath,
					}.Build(),
				},
				Calls: map[string]*config_v1.MCPCallDefinition{
					callIDMath: config_v1.MCPCallDefinition_builder{
						Id: &callIDMath,
					}.Build(),
				},
			}.Build(),
		}.Build())

		// 6. Seed Traffic Data
		if a.TopologyManager != nil {
			now := time.Now()
			var points []topology.TrafficPoint

			for i := 0; i < 24; i++ {
				t := now.Add(-time.Duration(24-i) * time.Hour)
				requests := 100 + int(50*math.Sin(float64(i)/3.0))
				points = append(points, topology.TrafficPoint{
					Time:  t.Format("15:04"),
					Total: int64(requests),
				})
			}
			a.TopologyManager.SeedTrafficHistory(points)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Seeded successfully"))
	}
}
