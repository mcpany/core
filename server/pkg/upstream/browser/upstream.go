// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides the browser automation upstream implementation.
package browser

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/playwright-community/playwright-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Upstream implements the upstream.Upstream interface for browser automation services.
type Upstream struct {
	mu             sync.Mutex // Protects lifecycle (start/stop)
	pageMu         sync.Mutex // Serializes access to the single page (session)
	pw             *playwright.Playwright
	browser        playwright.Browser
	page           playwright.Page
	serviceName    string

	// Config
	browserType    string
	headless       bool
	userAgent      string
	viewportWidth  int32
	viewportHeight int32
	screenshotDir  string
}

// NewUpstream creates a new instance of BrowserUpstream.
//
// Returns the result.
func NewUpstream() upstream.Upstream {
	return &Upstream{}
}

// Shutdown implements the upstream.Upstream interface.
//
// ctx is the context for the request.
//
// Returns an error if the operation fails.
func (u *Upstream) Shutdown(_ context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	// Close page first
	if u.page != nil {
		if err := u.page.Close(); err != nil {
			logging.GetLogger().Error("Failed to close page", "error", err)
		}
		u.page = nil
	}

	if u.browser != nil {
		if err := u.browser.Close(); err != nil {
			logging.GetLogger().Error("Failed to close browser", "error", err)
		}
		u.browser = nil
	}
	if u.pw != nil {
		if err := u.pw.Stop(); err != nil {
			logging.GetLogger().Error("Failed to stop playwright", "error", err)
		}
		u.pw = nil
	}
	return nil
}

// Register processes the configuration for a browser service.
//
// ctx is the context for the request.
// serviceConfig is the serviceConfig.
// toolManager is the toolManager.
// _ is an unused parameter.
// _ is an unused parameter.
// _ is an unused parameter.
//
// Returns the result.
func (u *Upstream) Register(
	_ context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	_ prompt.ManagerInterface,
	_ resource.ManagerInterface,
	_ bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	log := logging.GetLogger()

	if serviceConfig.GetId() == "" {
		h := sha256.New()
		h.Write([]byte(serviceConfig.GetName()))
		serviceConfig.SetId(hex.EncodeToString(h.Sum(nil)))
	}

	sanitizedName, err := util.SanitizeServiceName(serviceConfig.GetName())
	if err != nil {
		return "", nil, nil, err
	}
	serviceConfig.SetSanitizedName(sanitizedName)
	serviceID := sanitizedName
	u.serviceName = sanitizedName

	browserConfig := serviceConfig.GetBrowserService()
	if browserConfig == nil {
		return "", nil, nil, fmt.Errorf("browser service config is nil")
	}

	u.browserType = browserConfig.GetBrowserType()
	if u.browserType == "" {
		u.browserType = "chromium"
	}
	u.headless = browserConfig.GetHeadless()
	u.userAgent = browserConfig.GetUserAgent()
	u.viewportWidth = browserConfig.GetViewportWidth()
	u.viewportHeight = browserConfig.GetViewportHeight()
	u.screenshotDir = browserConfig.GetScreenshotDir()

	// Initialize Playwright
	if err := u.startBrowser(); err != nil {
		return "", nil, nil, fmt.Errorf("failed to start browser: %w", err)
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0)

	// Define tools
	tools := []struct {
		Name        string
		Description string
		Input       map[string]interface{}
		Handler     func(context.Context, map[string]interface{}) (map[string]interface{}, error)
	}{
		{
			Name:        "navigate",
			Description: "Navigate to a URL",
			Input: map[string]interface{}{
				"url": map[string]interface{}{"type": "string", "description": "The URL to navigate to"},
			},
			Handler: u.navigate,
		},
		{
			Name:        "screenshot",
			Description: "Take a screenshot of the current page",
			Input: map[string]interface{}{
				"full_page": map[string]interface{}{"type": "boolean", "description": "Whether to take a full page screenshot"},
			},
			Handler: u.screenshot,
		},
		{
			Name:        "content",
			Description: "Get the content of the current page",
			Input:       map[string]interface{}{},
			Handler:     u.content,
		},
	}

	for _, t := range tools {
		inputSchema, err := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Input,
		})
		if err != nil {
			log.Error("Failed to create input schema", "tool", t.Name, "error", err)
			continue
		}

		// Add required fields
		required := []interface{}{}
		if _, ok := t.Input["url"]; ok {
			required = append(required, "url")
		}
		if len(required) > 0 {
			m := map[string]interface{}{
				"type":       "object",
				"properties": t.Input,
				"required":   required,
			}
			inputSchema, _ = structpb.NewStruct(m)
		}

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String(t.Name),
			Description: proto.String(t.Description),
			ServiceId:   proto.String(serviceID),
			InputSchema: inputSchema,
		}.Build()

		callable := &browserCallable{handler: t.Handler}
		callableTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, nil)
		if err != nil {
			log.Error("Failed to create callable tool", "tool", t.Name, "error", err)
			continue
		}

		if err := toolManager.AddTool(callableTool); err != nil {
			log.Error("Failed to add tool", "tool", t.Name, "error", err)
			continue
		}
		discoveredTools = append(discoveredTools, toolDef)
	}

	log.Info("Registered browser service", "serviceID", serviceID, "tools", len(discoveredTools))
	return serviceID, discoveredTools, nil, nil
}

func (u *Upstream) startBrowser() error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.pw != nil {
		return nil
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %w", err)
	}
	u.pw = pw

	var browserType playwright.BrowserType
	switch u.browserType {
	case "firefox":
		browserType = pw.Firefox
	case "webkit":
		browserType = pw.WebKit
	default:
		browserType = pw.Chromium
	}

	browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(u.headless),
	})
	if err != nil {
		_ = pw.Stop()
		u.pw = nil
		return fmt.Errorf("could not launch browser: %w", err)
	}
	u.browser = browser
	return nil
}

func (u *Upstream) ensurePage() (playwright.Page, error) {
	// Assumes u.pageMu is held by caller

	if u.page != nil {
		if !u.page.IsClosed() {
			return u.page, nil
		}
		u.page = nil
	}

	u.mu.Lock()
	defer u.mu.Unlock()

	if u.browser == nil {
		return nil, fmt.Errorf("browser not initialized")
	}

	opts := playwright.BrowserNewContextOptions{}
	if u.userAgent != "" {
		opts.UserAgent = playwright.String(u.userAgent)
	}
	if u.viewportWidth > 0 && u.viewportHeight > 0 {
		opts.Viewport = &playwright.Size{
			Width:  int(u.viewportWidth),
			Height: int(u.viewportHeight),
		}
	}

	ctx, err := u.browser.NewContext(opts)
	if err != nil {
		return nil, err
	}

	page, err := ctx.NewPage()
	if err != nil {
		return nil, err
	}
	u.page = page
	return page, nil
}

func (u *Upstream) navigate(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required and must be a string")
	}

	u.pageMu.Lock()
	defer u.pageMu.Unlock()

	page, err := u.ensurePage()
	if err != nil {
		return nil, err
	}

	if _, err := page.Goto(url); err != nil {
		return nil, err
	}

	title, _ := page.Title()
	return map[string]interface{}{
		"title":  title,
		"url":    url,
		"status": "success",
	}, nil
}

func (u *Upstream) screenshot(_ context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	fullPage, _ := args["full_page"].(bool)

	u.pageMu.Lock()
	defer u.pageMu.Unlock()

	page, err := u.ensurePage()
	if err != nil {
		return nil, err
	}

	// Prepare path if directory is set
	var path *string
	if u.screenshotDir != "" {
		if err := os.MkdirAll(u.screenshotDir, 0750); err != nil {
			return nil, fmt.Errorf("failed to create screenshot directory: %w", err)
		}
		filename := fmt.Sprintf("screenshot-%d.png", time.Now().UnixNano())
		p := filepath.Join(u.screenshotDir, filename)
		path = &p
	}

	data, err := page.Screenshot(playwright.PageScreenshotOptions{
		FullPage: playwright.Bool(fullPage),
		Path:     path,
	})
	if err != nil {
		return nil, err
	}

	if path != nil {
		return map[string]interface{}{
			"path":   *path,
			"format": "png",
		}, nil
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return map[string]interface{}{
		"image_base64": encoded,
		"format":       "png",
	}, nil
}

func (u *Upstream) content(_ context.Context, _ map[string]interface{}) (map[string]interface{}, error) {
	u.pageMu.Lock()
	defer u.pageMu.Unlock()

	page, err := u.ensurePage()
	if err != nil {
		return nil, err
	}

	content, err := page.Content()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"content": content,
	}, nil
}

type browserCallable struct {
	handler func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}

// Call executes the browser tool.
func (c *browserCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := req.Arguments
	if args == nil && len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}
	}
	return c.handler(ctx, args)
}
