// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// SeedRequest represents the payload for seeding the server state.
type SeedRequest struct {
	Services []json.RawMessage       `json:"services"`
	Tools    []*SeedTool             `json:"tools"`
	Traffic  []topology.TrafficPoint `json:"traffic"`
}

// SeedTool represents a tool to be seeded.
type SeedTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	ServiceId   string          `json:"serviceId"`
	InputSchema map[string]any  `json:"inputSchema"`
	Output      json.RawMessage `json:"output"` // Pre-canned output
}

// handleDebugSeed handles the seeding of mock data.
func (a *Application) handleDebugSeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SeedRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		log := logging.GetLogger()

		// 1. Seed Services
		for _, rawSvc := range req.Services {
			var svc configv1.UpstreamServiceConfig
			if err := protojson.Unmarshal(rawSvc, &svc); err != nil {
				log.Error("Failed to unmarshal service", "error", err)
				http.Error(w, "Invalid service config: "+err.Error(), http.StatusBadRequest)
				return
			}
			if a.ServiceRegistry != nil {
				if err := a.ServiceRegistry.RegisterMockService(ctx, &svc); err != nil {
					log.Error("Failed to register mock service", "service", svc.GetName(), "error", err)
					http.Error(w, "Failed to register mock service: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// 2. Seed Tools
		for _, t := range req.Tools {
			mockTool := &tool.MockTool{
				ToolFunc: func() *v1.Tool {
					inputSchema, _ := structpb.NewStruct(t.InputSchema)
					return &v1.Tool{
						Name:        proto.String(t.Name),
						Description: proto.String(t.Description),
						ServiceId:   proto.String(t.ServiceId),
						InputSchema: inputSchema,
					}
				},
				MCPToolFunc: func() *mcp.Tool {
					return &mcp.Tool{
						Name:        t.Name,
						Description: t.Description,
						InputSchema: t.InputSchema,
					}
				},
				ExecuteFunc: func(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
					// Return the pre-canned output
					var output any
					if len(t.Output) > 0 {
						if err := json.Unmarshal(t.Output, &output); err != nil {
							return nil, fmt.Errorf("failed to unmarshal mock output: %w", err)
						}
					}
					return output, nil
				},
			}

			if a.ToolManager != nil {
				if err := a.ToolManager.AddTool(mockTool); err != nil {
					log.Error("Failed to add mock tool", "tool", t.Name, "error", err)
					http.Error(w, "Failed to add mock tool: "+err.Error(), http.StatusInternalServerError)
					return
				}
			}
		}

		// 3. Seed Traffic
		if len(req.Traffic) > 0 && a.TopologyManager != nil {
			a.TopologyManager.SeedTrafficHistory(req.Traffic)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("{}"))
	}
}
