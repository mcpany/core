package browser

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/playwright-community/playwright-go"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/upstream"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

// Upstream implements the upstream.Upstream interface for browser automation.
type Upstream struct {
	mu        sync.Mutex
	pw        *playwright.Playwright
	browser   playwright.Browser
	context   playwright.BrowserContext
	page      playwright.Page
	serviceID string
}

// NewUpstream creates a new instance of BrowserUpstream.
func NewUpstream() upstream.Upstream {
	return &Upstream{}
}

// Shutdown gracefully terminates the upstream service.
func (u *Upstream) Shutdown(ctx context.Context) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if u.context != nil {
		_ = u.context.Close()
	}
	if u.browser != nil {
		_ = u.browser.Close()
	}
	if u.pw != nil {
		_ = u.pw.Stop()
	}
	return nil
}

// Register inspects the upstream service defined by the serviceConfig,
// discovers its capabilities, and registers them.
func (u *Upstream) Register(
	ctx context.Context,
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
	u.serviceID = sanitizedName

	browserConfig := serviceConfig.GetBrowserService()
	if browserConfig == nil {
		return "", nil, nil, fmt.Errorf("browser service config is nil")
	}

	// Initialize Playwright
	if err := u.startBrowser(browserConfig); err != nil {
		log.Error("Failed to start browser (ensure playwright and browsers are installed)", "error", err)
		return "", nil, nil, fmt.Errorf("failed to start browser: %w", err)
	}

	info := &tool.ServiceInfo{
		Name:   serviceConfig.GetName(),
		Config: serviceConfig,
	}
	toolManager.AddServiceInfo(u.serviceID, info)

	tools := []struct {
		Name    string
		Desc    string
		Input   map[string]interface{}
		Handler func(context.Context, map[string]interface{}) (map[string]interface{}, error)
	}{
		{
			Name: "browse_open",
			Desc: "Opens a URL in the browser",
			Input: map[string]interface{}{
				"url": map[string]interface{}{"type": "string", "description": "The URL to open"},
			},
			Handler: u.handleOpen,
		},
		{
			Name: "browse_screenshot",
			Desc: "Takes a screenshot of the current page",
			Input: map[string]interface{}{},
			Handler: u.handleScreenshot,
		},
		{
			Name: "browse_click",
			Desc: "Clicks an element on the page",
			Input: map[string]interface{}{
				"selector": map[string]interface{}{"type": "string", "description": "CSS selector or text to click"},
			},
			Handler: u.handleClick,
		},
		{
			Name: "browse_content",
			Desc: "Gets the text content of the page",
			Input: map[string]interface{}{},
			Handler: u.handleContent,
		},
	}

	discoveredTools := make([]*configv1.ToolDefinition, 0)

	for _, t := range tools {
		inputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Input,
		})
		outputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type": "object",
		})

		toolDef := configv1.ToolDefinition_builder{
			Name:        proto.String(t.Name),
			Description: proto.String(t.Desc),
			ServiceId:   proto.String(u.serviceID),
		}.Build()

		callable := &browserCallable{handler: t.Handler}
		callableTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
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

	return u.serviceID, discoveredTools, nil, nil
}

func (u *Upstream) startBrowser(cfg *configv1.BrowserUpstreamService) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	var err error
	u.pw, err = playwright.Run()
	if err != nil {
		return err
	}

	opts := playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(cfg.GetHeadless()),
	}

	switch cfg.GetBrowserType() {
	case "firefox":
		u.browser, err = u.pw.Firefox.Launch(opts)
	case "webkit":
		u.browser, err = u.pw.WebKit.Launch(opts)
	default:
		u.browser, err = u.pw.Chromium.Launch(opts)
	}
	if err != nil {
		return err
	}

	u.context, err = u.browser.NewContext()
	if err != nil {
		return err
	}

	u.page, err = u.context.NewPage()
	return err
}

func (u *Upstream) handleOpen(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("url is required")
	}

	if _, err := u.page.Goto(url); err != nil {
		return nil, err
	}

	return map[string]interface{}{"status": "opened", "url": url}, nil
}

func (u *Upstream) handleScreenshot(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	data, err := u.page.Screenshot()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"screenshot_base64": data}, nil
}

func (u *Upstream) handleClick(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	selector, ok := args["selector"].(string)
	if !ok {
		return nil, fmt.Errorf("selector is required")
	}

	if err := u.page.Click(selector); err != nil {
		return nil, err
	}

	return map[string]interface{}{"status": "clicked", "selector": selector}, nil
}

func (u *Upstream) handleContent(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	content, err := u.page.Content()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{"content": content}, nil
}

type browserCallable struct {
	handler func(ctx context.Context, args map[string]interface{}) (map[string]interface{}, error)
}

func (c *browserCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	args := req.Arguments
	if args == nil && len(req.ToolInputs) > 0 {
		if err := json.Unmarshal(req.ToolInputs, &args); err != nil {
			return nil, fmt.Errorf("failed to unmarshal arguments: %w", err)
		}
	}
	return c.handler(ctx, args)
}
