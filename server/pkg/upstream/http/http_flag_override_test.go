// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestHTTPUpstream_URLConstruction_FlagOverride(t *testing.T) {
	tests := []struct {
		name        string
		baseAddr    string
		endpoint    string
		expectedURL string
	}{
		{
			name:        "Endpoint overrides flag with empty value",
			baseAddr:    "http://api.example.com?debug",
			endpoint:    "/status?debug=",
			expectedURL: "http://api.example.com/status?debug=",
		},
		{
			name:        "Endpoint overrides empty value with flag",
			baseAddr:    "http://api.example.com?debug=",
			endpoint:    "/status?debug",
			expectedURL: "http://api.example.com/status?debug",
		},
		{
			name:        "Base has flag, endpoint absent",
			baseAddr:    "http://api.example.com?debug",
			endpoint:    "/status",
			expectedURL: "http://api.example.com/status?debug",
		},
		{
			name:        "Base has empty value, endpoint absent",
			baseAddr:    "http://api.example.com?debug=",
			endpoint:    "/status",
			expectedURL: "http://api.example.com/status?debug=",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			pm := pool.NewManager()
			u := NewUpstream(pm)

			cfg := &configv1.UpstreamServiceConfig{
				Name: proto.String("test-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String(tt.baseAddr),
						Calls: map[string]*configv1.HttpCallDefinition{
							"test-call": {
								EndpointPath: proto.String(tt.endpoint),
								Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
							},
						},
						Tools: []*configv1.ToolDefinition{
							{
								Name:   proto.String("test-tool"),
								CallId: proto.String("test-call"),
							},
						},
					},
				},
			}

			tm := tool.NewManager(nil)
			rm := resource.NewManager()
			sm := prompt.NewManager()

			_, tools, _, err := u.Register(ctx, cfg, tm, sm, rm, false)
			assert.NoError(t, err)
			assert.Len(t, tools, 1)

			// Get the tool and check its FQN which contains the URL
			registeredTool, ok := tm.GetTool("test-service.test-tool")
			assert.True(t, ok)
			assert.Contains(t, registeredTool.Tool().GetUnderlyingMethodFqn(), tt.expectedURL)
		})
	}
}
