// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides browser automation tools.
package browser

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/playwright-community/playwright-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Upstream implements the Upstream interface for browser automation.
type Upstream struct {
	mu          sync.Mutex
	pw          *playwright.Playwright
	browser     playwright.Browser
	context     playwright.BrowserContext
	page        playwright.Page
	serviceKey  string
	initialized bool
}

// NewUpstream creates a new browser upstream instance.
func NewUpstream() *Upstream {
	return &Upstream{}
}

// Shutdown stops the playwright instance and closes the browser.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.page != nil {
		_ = u.page.Close()
	}
	if u.context != nil {
		_ = u.context.Close()
	}
	if u.browser != nil {
		_ = u.browser.Close()
	}
	if u.pw != nil {
		_ = u.pw.Stop()
	}
	u.initialized = false
	return nil
}

// CheckHealth checks if the browser is responsive.
func (u *Upstream) CheckHealth(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.initialized {
		return fmt.Errorf("browser upstream not initialized")
	}
	if u.page == nil {
		return fmt.Errorf("browser page not available")
	}
	// Simple liveness check
	_, err := u.page.Title()
	return err
}

type browserToolDef struct {
	Name        string
	Description string
	Input       map[string]interface{}
}

// Register initializes the browser and registers automation tools.
func (u *Upstream) Register(
	_ context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	_ prompt.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	log := logging.GetLogger()

	// TODO: Handle serviceConfig for options like headless mode
	u.serviceKey = serviceConfig.GetId()

	if !u.initialized {
		if err := u.initializeBrowser(); err != nil {
			return "", nil, nil, fmt.Errorf("failed to initialize browser: %w", err)
		}
		u.initialized = true
	}

	definitions := []browserToolDef{
		{
			Name:        "browser_navigate",
			Description: "Navigate the browser to a specific URL",
			Input: map[string]interface{}{
				"url": map[string]interface{}{"type": "string", "description": "URL to navigate to"},
			},
		},
		{
			Name:        "browser_click",
			Description: "Click an element matching the selector",
			Input: map[string]interface{}{
				"selector": map[string]interface{}{"type": "string", "description": "CSS selector of element to click"},
			},
		},
		{
			Name:        "browser_fill",
			Description: "Fill an input element with text",
			Input: map[string]interface{}{
				"selector": map[string]interface{}{"type": "string", "description": "CSS selector of input to fill"},
				"value":    map[string]interface{}{"type": "string", "description": "Text value to fill"},
			},
		},
		{
			Name:        "browser_screenshot",
			Description: "Take a screenshot of the current page",
			Input:       map[string]interface{}{},
		},
		{
			Name:        "browser_content",
			Description: "Get the HTML content of the current page",
			Input:       map[string]interface{}{},
		},
		{
			Name:        "browser_evaluate",
			Description: "Evaluate JavaScript on the page",
			Input: map[string]interface{}{
				"script": map[string]interface{}{"type": "string", "description": "JavaScript code to evaluate"},
			},
		},
	}

	registeredTools := make([]*configv1.ToolDefinition, 0)

	for _, def := range definitions {
		inputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": def.Input,
		})
		if err != nil {
			log.Error("Failed to create input schema", "tool", def.Name, "error", err)
			continue
		}

		// Empty output schema for now as we return unstructured maps mostly
		outputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type": "object",
		})

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String(def.Name),
			Description: proto.String(def.Description),
			ServiceId:   proto.String(u.serviceKey),
			InputSchema: inputSchema,
		}.Build()

		callable := &browserCallable{
			upstream: u,
			toolName: def.Name,
		}

		t, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		if err != nil {
			log.Error("Failed to create callable tool", "tool", def.Name, "error", err)
			continue
		}

		if err := toolManager.AddTool(t); err != nil {
			log.Error("Failed to add tool", "tool", def.Name, "error", err)
			continue
		}

		registeredTools = append(registeredTools, toolDef)
	}

	return u.serviceKey, registeredTools, nil, nil
}

type browserCallable struct {
	upstream *Upstream
	toolName string
}

func (c *browserCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return c.upstream.handleToolExecution(ctx, c.toolName, req.Arguments)
}

func (u *Upstream) initializeBrowser() error {
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not launch playwright: %w", err)
	}
	u.pw = pw

	// TODO: Make headless configurable
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		_ = pw.Stop()
		return fmt.Errorf("could not launch chromium: %w", err)
	}
	u.browser = browser

	bgCtx, err := browser.NewContext()
	if err != nil {
		_ = browser.Close()
		_ = pw.Stop()
		return fmt.Errorf("could not create browser context: %w", err)
	}
	u.context = bgCtx

	page, err := bgCtx.NewPage()
	if err != nil {
		_ = bgCtx.Close()
		_ = browser.Close()
		_ = pw.Stop()
		return fmt.Errorf("could not create page: %w", err)
	}
	u.page = page

	return nil
}

func (u *Upstream) handleToolExecution(_ context.Context, name string, args map[string]interface{}) (map[string]interface{}, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	if !u.initialized {
		return nil, fmt.Errorf("browser not initialized")
	}

	switch name {
	case "browser_navigate":
		url, ok := args["url"].(string)
		if !ok {
			return nil, fmt.Errorf("url argument required")
		}
		// Use a reasonable timeout
		_, err := u.page.Goto(url, playwright.PageGotoOptions{
			Timeout: playwright.Float(30000),
		})
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"result": "navigated to " + url}, nil

	case "browser_click":
		selector, ok := args["selector"].(string)
		if !ok {
			return nil, fmt.Errorf("selector argument required")
		}
		err := u.page.Locator(selector).Click()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"result": "clicked " + selector}, nil

	case "browser_fill":
		selector, ok := args["selector"].(string)
		if !ok {
			return nil, fmt.Errorf("selector argument required")
		}
		value, ok := args["value"].(string)
		if !ok {
			return nil, fmt.Errorf("value argument required")
		}
		err := u.page.Locator(selector).Fill(value)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"result": "filled " + selector}, nil

	case "browser_screenshot":
		data, err := u.page.Screenshot()
		if err != nil {
			return nil, err
		}
		encoded := base64.StdEncoding.EncodeToString(data)
		return map[string]interface{}{"image_base64": encoded}, nil

	case "browser_content":
		content, err := u.page.Content()
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"content": content}, nil

	case "browser_evaluate":
		script, ok := args["script"].(string)
		if !ok {
			return nil, fmt.Errorf("script argument required")
		}
		result, err := u.page.Evaluate(script)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"result": result}, nil
	}

	return nil, fmt.Errorf("unknown tool: %s", name)
}
