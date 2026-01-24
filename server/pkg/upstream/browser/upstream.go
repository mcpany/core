// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	routerv1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/playwright-community/playwright-go"
	"google.golang.org/protobuf/types/known/structpb"
)

type BrowserUpstream struct {
	pw      *playwright.Playwright
	browser playwright.Browser
	mu      sync.Mutex
	config  *configv1.BrowserUpstreamService
}

func NewUpstream() *BrowserUpstream {
	return &BrowserUpstream{}
}

func (u *BrowserUpstream) Shutdown(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.browser != nil {
		if err := u.browser.Close(); err != nil {
			return err
		}
	}
	if u.pw != nil {
		if err := u.pw.Stop(); err != nil {
			return err
		}
	}
	return nil
}

func (u *BrowserUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger().With("service", "browser", "upstream", serviceConfig.GetName())

	// Only initialize playwright once
	u.mu.Lock()
	if u.pw == nil {
		// Attempt to install playwright dependencies if needed
		// In a real environment, this should probably be done at build/setup time
		// playwright.Install()

		pw, err := playwright.Run()
		if err != nil {
			u.mu.Unlock()
			return "", nil, nil, fmt.Errorf("could not start playwright: %w", err)
		}
		u.pw = pw
	}

	config := serviceConfig.GetBrowserService()
	u.config = config

	if u.browser == nil {
		opts := playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(config.GetHeadless()),
		}

		browserType := config.GetBrowserType()
		if browserType == "" {
			browserType = "chromium"
		}

		var err error
		var bt playwright.BrowserType
		switch browserType {
		case "firefox":
			bt = u.pw.Firefox
		case "webkit":
			bt = u.pw.WebKit
		default:
			bt = u.pw.Chromium
		}

		u.browser, err = bt.Launch(opts)
		if err != nil {
			u.mu.Unlock()
			return "", nil, nil, fmt.Errorf("could not launch browser: %w", err)
		}
	}
	u.mu.Unlock()

	tools := []*configv1.ToolDefinition{
		{
			Name:        "navigate",
			Description: "Navigate to a URL",
			Arguments: []*configv1.ArgumentDefinition{
				{Name: "url", Description: "The URL to navigate to", Type: "string", Required: true},
			},
		},
		{
			Name:        "screenshot",
			Description: "Take a screenshot of the current page",
			Arguments: []*configv1.ArgumentDefinition{
				{Name: "full_page", Description: "Capture full page", Type: "boolean"},
			},
		},
		{
			Name:        "get_content",
			Description: "Get the content of the current page",
			Arguments: []*configv1.ArgumentDefinition{},
		},
		{
			Name:        "click",
			Description: "Click an element",
			Arguments: []*configv1.ArgumentDefinition{
				{Name: "selector", Description: "CSS selector or XPath", Type: "string", Required: true},
			},
		},
		{
			Name:        "fill",
			Description: "Fill an input field",
			Arguments: []*configv1.ArgumentDefinition{
				{Name: "selector", Description: "CSS selector or XPath", Type: "string", Required: true},
				{Name: "value", Description: "Value to fill", Type: "string", Required: true},
			},
		},
	}

	// Register tools
	for _, t := range tools {
		inputSchemaMap := map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
			"required":   []interface{}{},
		}
		props := inputSchemaMap["properties"].(map[string]interface{})
		required := []interface{}{}

		for _, arg := range t.Arguments {
			props[arg.Name] = map[string]interface{}{
				"type":        arg.Type,
				"description": arg.Description,
			}
			if arg.Required {
				required = append(required, arg.Name)
			}
		}
		inputSchemaMap["required"] = required

		inputSchema, err := structpb.NewStruct(inputSchemaMap)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to create input schema for tool %s: %w", t.Name, err)
		}

		t.ServiceId = serviceConfig.GetId() // Important!

		v1Tool, err := tool.ConvertToolDefinitionToProto(t, inputSchema, nil)
		if err != nil {
			return "", nil, nil, fmt.Errorf("failed to convert tool definition for %s: %w", t.Name, err)
		}

		handler := u.createHandler(t.Name)
		browserTool := &BrowserTool{
			toolDef: v1Tool,
			handler: handler,
		}

		if err := toolManager.AddTool(browserTool); err != nil {
			log.Error("Failed to register tool", "tool", t.Name, "error", err)
			return "", nil, nil, err
		}
	}

	return serviceConfig.GetId(), tools, nil, nil
}

func (u *BrowserUpstream) createHandler(toolName string) tool.HandlerFunc {
	return func(ctx context.Context, req *tool.ExecutionRequest) (interface{}, error) {
		u.mu.Lock()
		defer u.mu.Unlock()

		args := req.Arguments

		if u.browser == nil {
			return nil, fmt.Errorf("browser not initialized")
		}

		browserContexts := u.browser.Contexts()
		var bCtx playwright.BrowserContext
		if len(browserContexts) == 0 {
			var err error
			bCtx, err = u.browser.NewContext()
			if err != nil {
				return nil, err
			}
		} else {
			bCtx = browserContexts[0]
		}

		pages := bCtx.Pages()
		var page playwright.Page
		if len(pages) == 0 {
			var err error
			page, err = bCtx.NewPage()
			if err != nil {
				return nil, err
			}
		} else {
			page = pages[0]
		}

		switch toolName {
		case "navigate":
			url, ok := args["url"].(string)
			if !ok {
				return nil, fmt.Errorf("url argument required")
			}
			_, err := page.Goto(url)
			if err != nil {
				return nil, err
			}
			return map[string]string{"status": "navigated", "url": url}, nil

		case "screenshot":
			fullPage := false
			if fp, ok := args["full_page"].(bool); ok {
				fullPage = fp
			}
			bytes, err := page.Screenshot(playwright.PageScreenshotOptions{
				FullPage: playwright.Bool(fullPage),
			})
			if err != nil {
				return nil, err
			}
			// Returning bytes directly might be handled by the tool manager to encode.
			// However, usually we return base64 or similar.
			// Assuming the caller handles []byte correctly (e.g. JSON marshalling might base64 it).
			return map[string]interface{}{"screenshot_bytes": bytes}, nil

		case "get_content":
			content, err := page.Content()
			if err != nil {
				return nil, err
			}
			return map[string]string{"content": content}, nil

		case "click":
			selector, ok := args["selector"].(string)
			if !ok {
				return nil, fmt.Errorf("selector argument required")
			}
			err := page.Click(selector)
			if err != nil {
				return nil, err
			}
			return map[string]string{"status": "clicked", "selector": selector}, nil

		case "fill":
			selector, ok := args["selector"].(string)
			if !ok {
				return nil, fmt.Errorf("selector argument required")
			}
			value, ok := args["value"].(string)
			if !ok {
				return nil, fmt.Errorf("value argument required")
			}
			err := page.Fill(selector, value)
			if err != nil {
				return nil, err
			}
			return map[string]string{"status": "filled", "selector": selector, "value": value}, nil
		}

		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}

// BrowserTool implements tool.Tool interface
type BrowserTool struct {
	toolDef     *routerv1.Tool // proto tool definition
	handler     tool.HandlerFunc
	mcpTool     *mcp.Tool
	mcpToolOnce sync.Once
}

func (t *BrowserTool) Tool() *routerv1.Tool {
	return t.toolDef
}

func (t *BrowserTool) MCPTool() *mcp.Tool {
	t.mcpToolOnce.Do(func() {
		var err error
		t.mcpTool, err = tool.ConvertProtoToMCPTool(t.toolDef)
		if err != nil {
			logging.GetLogger().Error("Failed to convert tool to MCP tool", "toolName", t.toolDef.GetName(), "error", err)
		}
	})
	return t.mcpTool
}

func (t *BrowserTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return t.handler(ctx, req)
}

func (t *BrowserTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}
