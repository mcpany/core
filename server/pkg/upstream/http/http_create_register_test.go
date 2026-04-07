package http

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/pool"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
)

func TestUpstream_createAndRegisterHTTPTools(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		address       string
		serviceConfig *configv1.UpstreamServiceConfig
		expectedTools int
		wantErr       bool
	}{
		{
			name:    "Happy Path - Auto Discover Tools",
			address: "http://example.com",
			serviceConfig: &configv1.UpstreamServiceConfig{
				AutoDiscoverTool: true,
				HttpService: &configv1.HttpService{
					Calls: map[string]*configv1.HttpCall{
						"call1": {
							Method:       configv1.HttpMethod_HTTP_METHOD_GET,
							EndpointPath: "/api/v1/resource",
						},
					},
				},
			},
			expectedTools: 1,
			wantErr:       false,
		},
		{
			name:    "Invalid Address - Base URL Parse Error",
			address: "://invalid-url",
			serviceConfig: &configv1.UpstreamServiceConfig{
				HttpService: &configv1.HttpService{
					Tools: []*configv1.ToolDefinition{
						{Name: func() *string { s := "tool1"; return &s }(), CallId: func() *string { s := "call1"; return &s }()},
					},
					Calls: map[string]*configv1.HttpCall{
						"call1": {Method: configv1.HttpMethod_HTTP_METHOD_GET, EndpointPath: "/"},
					},
				},
			},
			expectedTools: 0,
			wantErr:       false, // It just returns nil gracefully
		},
		{
			name:    "Missing Tool Definition",
			address: "http://example.com",
			serviceConfig: &configv1.UpstreamServiceConfig{
				HttpService: &configv1.HttpService{
					Calls: map[string]*configv1.HttpCall{
						"call1": {Method: configv1.HttpMethod_HTTP_METHOD_GET, EndpointPath: "/"},
					},
				},
			},
			expectedTools: 0,
			wantErr:       false,
		},
		{
			name:    "Disabled Tool",
			address: "http://example.com",
			serviceConfig: &configv1.UpstreamServiceConfig{
				HttpService: &configv1.HttpService{
					Tools: []*configv1.ToolDefinition{
						{Name: func() *string { s := "tool1"; return &s }(), CallId: func() *string { s := "call1"; return &s }(), Disable: true},
					},
					Calls: map[string]*configv1.HttpCall{
						"call1": {Method: configv1.HttpMethod_HTTP_METHOD_GET, EndpointPath: "/"},
					},
				},
			},
			expectedTools: 0,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := pool.NewManager()
			tm := tool.NewManager(nil)
			rm := resource.NewManager()
			u := NewUpstream(pm, nil)

			discovered := u.createAndRegisterHTTPTools(ctx, "test-service", tt.address, tt.serviceConfig, tm, rm, false)
			assert.Len(t, discovered, tt.expectedTools)
		})
	}
}

func TestUpstream_createAndRegisterPrompts(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name            string
		serviceConfig   *configv1.UpstreamServiceConfig
		expectedPrompts int
	}{
		{
			name: "Happy Path - Register Prompt",
			serviceConfig: &configv1.UpstreamServiceConfig{
				HttpService: &configv1.HttpService{
					Prompts: []*configv1.PromptDefinition{
						{
							Name:        "prompt1",
							Description: "A test prompt",
							CallId:      "call1",
						},
					},
					Calls: map[string]*configv1.HttpCall{
						"call1": {Method: configv1.HttpMethod_HTTP_METHOD_GET, EndpointPath: "/prompt"},
					},
				},
			},
			expectedPrompts: 1, // Only tracking execution success since manager doesn't expose len easily here
		},
		{
			name: "Disabled Prompt",
			serviceConfig: &configv1.UpstreamServiceConfig{
				HttpService: &configv1.HttpService{
					Prompts: []*configv1.PromptDefinition{
						{
							Name:    "prompt1",
							Disable: true,
							CallId:  "call1",
						},
					},
					Calls: map[string]*configv1.HttpCall{
						"call1": {Method: configv1.HttpMethod_HTTP_METHOD_GET, EndpointPath: "/prompt"},
					},
				},
			},
			expectedPrompts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pm := pool.NewManager()
			promptManager := prompt.NewManager(nil)
			u := NewUpstream(pm, nil)

			u.createAndRegisterPrompts(ctx, "test-service", tt.serviceConfig, promptManager, false)

			// Verify behavior by checking the state of the prompt manager
			registeredPrompts := promptManager.ListPrompts()
			assert.Len(t, registeredPrompts, tt.expectedPrompts)

			if tt.expectedPrompts > 0 {
				p, ok := promptManager.GetPrompt(tt.serviceConfig.HttpService.Prompts[0].Name)
				assert.True(t, ok, "Expected prompt to be registered")
				assert.Equal(t, tt.serviceConfig.HttpService.Prompts[0].Name, p.Prompt().Name)
			}
		})
	}
}
