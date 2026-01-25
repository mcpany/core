package browser

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/playwright-community/playwright-go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type BrowserUpstream struct {
	config     *configv1.BrowserUpstreamService
	playwright *playwright.Playwright
	browser    playwright.Browser
	mu         sync.Mutex
}

func NewBrowserUpstream() *BrowserUpstream {
	return &BrowserUpstream{}
}

func (u *BrowserUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	browserConfig := serviceConfig.GetBrowserService()
	if browserConfig == nil {
		return "", nil, nil, fmt.Errorf("browser service config is nil")
	}
	u.config = browserConfig

	// Initialize Playwright if needed
	if err := u.ensureBrowser(); err != nil {
		return "", nil, nil, err
	}

	serviceName := serviceConfig.GetName()
	tools := []*configv1.ToolDefinition{}

	// Define tools
	navigateToolDef := &configv1.ToolDefinition{
		Name:        proto.String("navigate"),
		Title:       proto.String("Navigate"),
		Description: proto.String("Navigates the browser to a URL"),
		ServiceId:   proto.String(serviceName),
	}
	navigateInput, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{"type": "string", "description": "URL to navigate to"},
		},
		"required": []interface{}{"url"},
	})

	navigateCallback := &browserCallback{
		fn: u.handleNavigate,
	}

	navigateTool, err := tool.NewCallableTool(navigateToolDef, serviceConfig, navigateCallback, navigateInput, nil)
	if err != nil {
		return "", nil, nil, err
	}
	if err := toolManager.AddTool(navigateTool); err != nil {
		return "", nil, nil, err
	}
	tools = append(tools, navigateToolDef)

	screenshotToolDef := &configv1.ToolDefinition{
		Name:        proto.String("screenshot"),
		Title:       proto.String("Screenshot"),
		Description: proto.String("Takes a screenshot of the current page"),
		ServiceId:   proto.String(serviceName),
	}
	screenshotInput, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"full_page": map[string]interface{}{"type": "boolean", "description": "Capture full page"},
		},
	})
	screenshotCallback := &browserCallback{
		fn: u.handleScreenshot,
	}
	screenshotTool, err := tool.NewCallableTool(screenshotToolDef, serviceConfig, screenshotCallback, screenshotInput, nil)
	if err != nil {
		return "", nil, nil, err
	}
	if err := toolManager.AddTool(screenshotTool); err != nil {
		return "", nil, nil, err
	}
	tools = append(tools, screenshotToolDef)

	contentToolDef := &configv1.ToolDefinition{
		Name:        proto.String("content"),
		Title:       proto.String("Get Content"),
		Description: proto.String("Get the content of the current page"),
		ServiceId:   proto.String(serviceName),
	}
	contentInput, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
	})
	contentCallback := &browserCallback{
		fn: u.handleContent,
	}
	contentTool, err := tool.NewCallableTool(contentToolDef, serviceConfig, contentCallback, contentInput, nil)
	if err != nil {
		return "", nil, nil, err
	}
	if err := toolManager.AddTool(contentTool); err != nil {
		return "", nil, nil, err
	}
	tools = append(tools, contentToolDef)

	return serviceConfig.GetId(), tools, nil, nil
}

func (u *BrowserUpstream) Shutdown(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.browser != nil {
		if err := u.browser.Close(); err != nil {
			return err
		}
		u.browser = nil
	}
	if u.playwright != nil {
		if err := u.playwright.Stop(); err != nil {
			return err
		}
		u.playwright = nil
	}
	return nil
}

func (u *BrowserUpstream) ensureBrowser() error {
	if u.playwright == nil {
		pw, err := playwright.Run()
		if err != nil {
			return fmt.Errorf("could not launch playwright: %w", err)
		}
		u.playwright = pw
	}

	if u.browser == nil {
		browserType := u.playwright.Chromium
		if u.config.GetBrowser() == "firefox" {
			browserType = u.playwright.Firefox
		} else if u.config.GetBrowser() == "webkit" {
			browserType = u.playwright.WebKit
		}

		browser, err := browserType.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(u.config.GetHeadless()),
		})
		if err != nil {
			return fmt.Errorf("could not launch browser: %w", err)
		}
		u.browser = browser
	}
	return nil
}

func (u *BrowserUpstream) getPage() (playwright.Page, error) {
	if u.browser == nil {
		return nil, fmt.Errorf("browser not initialized")
	}
	contexts := u.browser.Contexts()
	if len(contexts) == 0 {
		ctx, err := u.browser.NewContext()
		if err != nil {
			return nil, err
		}
		return ctx.NewPage()
	}
	pages := contexts[0].Pages()
	if len(pages) == 0 {
		return contexts[0].NewPage()
	}
	return pages[0], nil
}

type browserCallback struct {
	fn func(context.Context, map[string]interface{}) (any, error)
}

func (c *browserCallback) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return c.fn(ctx, req.Arguments)
}

func (u *BrowserUpstream) handleNavigate(ctx context.Context, args map[string]interface{}) (any, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required")
	}

	page, err := u.getPage()
	if err != nil {
		return nil, err
	}

	if _, err := page.Goto(url); err != nil {
		return nil, err
	}

	return map[string]interface{}{"status": "navigated", "url": url}, nil
}

func (u *BrowserUpstream) handleScreenshot(ctx context.Context, args map[string]interface{}) (any, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	fullPage := false
	if val, ok := args["full_page"].(bool); ok {
		fullPage = val
	}

	page, err := u.getPage()
	if err != nil {
		return nil, err
	}

	data, err := page.Screenshot(playwright.PageScreenshotOptions{FullPage: playwright.Bool(fullPage)})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"screenshot_base64": base64.StdEncoding.EncodeToString(data),
	}, nil
}

func (u *BrowserUpstream) handleContent(ctx context.Context, args map[string]interface{}) (any, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	page, err := u.getPage()
	if err != nil {
		return nil, err
	}

	content, err := page.Content()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"content": content}, nil
}
