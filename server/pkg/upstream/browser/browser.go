// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides the upstream service implementation for browser automation.
package browser

import (
	"context"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/playwright-community/playwright-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Upstream implements the upstream.Upstream interface for browser automation.
type Upstream struct {
	pw       *playwright.Playwright
	browser  playwright.Browser
	mu       sync.Mutex
	isClosed bool
}

// NewUpstream creates a new Browser upstream.
func NewUpstream() (*Upstream, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not start playwright: %w", err)
	}

	browser, err := pw.Chromium.Launch()
	if err != nil {
		_ = pw.Stop()
		return nil, fmt.Errorf("could not launch browser: %w", err)
	}

	return &Upstream{
		pw:      pw,
		browser: browser,
	}, nil
}

// Shutdown gracefully shuts down the browser upstream.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.isClosed {
		return nil
	}
	u.isClosed = true
	if u.browser != nil {
		_ = u.browser.Close()
	}
	if u.pw != nil {
		_ = u.pw.Stop()
	}
	return nil
}

// Register registers the browser tool with the manager.
func (u *Upstream) Register(
	_ context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	_ prompt.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	browserService := serviceConfig.GetBrowserService()
	if browserService == nil {
		return "", nil, nil, fmt.Errorf("browser service config is nil")
	}

	serviceID := serviceConfig.GetName()
	toolName := "browse"
	if len(browserService.GetTools()) > 0 {
		toolName = browserService.GetTools()[0].GetName()
	}

	// Create input schema as structpb
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{
				"type":        "string",
				"description": "The URL to visit",
			},
		},
		"required": []interface{}{"url"},
	})

	// Create routerv1.Tool using builder
	routerTool := routerv1.Tool_builder{
		Name:        proto.String(toolName),
		Description: proto.String("Browse a webpage and return its content (HTML)"),
		ServiceId:   proto.String(serviceID),
		Annotations: routerv1.ToolAnnotations_builder{
			InputSchema: inputSchema,
		}.Build(),
	}.Build()

	handler := func(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
		urlVal, ok := args["url"].(string)
		if !ok {
			return nil, fmt.Errorf("url argument is required and must be a string")
		}

		u.mu.Lock()
		defer u.mu.Unlock()
		if u.isClosed {
			return nil, fmt.Errorf("upstream is closed")
		}

		page, err := u.browser.NewPage()
		if err != nil {
			return nil, fmt.Errorf("could not create page: %w", err)
		}
		defer func() {
			_ = page.Close()
		}()

		if _, err = page.Goto(urlVal); err != nil {
			return nil, fmt.Errorf("could not goto url: %w", err)
		}

		content, err := page.Content()
		if err != nil {
			return nil, fmt.Errorf("could not get content: %w", err)
		}

		return map[string]interface{}{
			"content": content,
		}, nil
	}

	browserTool := &Tool{
		toolDef: routerTool,
		handler: handler,
	}

	if err := toolManager.AddTool(browserTool); err != nil {
		return "", nil, nil, fmt.Errorf("failed to register tool: %w", err)
	}

	// Create configv1.ToolDefinition using builder
	configToolDef := configv1.ToolDefinition_builder{
		Name:        proto.String(toolName),
		Description: proto.String("Browse a webpage and return its content (HTML)"),
		ServiceId:   proto.String(serviceID),
	}.Build()

	return serviceID, []*configv1.ToolDefinition{configToolDef}, nil, nil
}

// Tool implements tool.Tool interface.
type Tool struct {
	toolDef *routerv1.Tool
	handler func(context.Context, map[string]interface{}) (map[string]interface{}, error)
}

// Tool returns the protobuf tool definition.
func (t *Tool) Tool() *routerv1.Tool {
	return t.toolDef
}

// MCPTool returns the MCP tool definition.
func (t *Tool) MCPTool() *mcp.Tool {
	mcpTool, _ := tool.ConvertProtoToMCPTool(t.toolDef)
	return mcpTool
}

// Execute executes the tool.
func (t *Tool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return t.handler(ctx, req.Arguments)
}

// GetCacheConfig returns the cache configuration.
func (t *Tool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
