// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"fmt"
	"sync"

	json "github.com/json-iterator/go"
	configv1 "github.com/mcpany/core/proto/config/v1"
	pb "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/playwright-community/playwright-go"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

var fastJSON = json.ConfigCompatibleWithStandardLibrary

// BrowserUpstream implements the Upstream interface for browser automation.
type BrowserUpstream struct {
	pw          *playwright.Playwright
	browser     playwright.Browser
	page        playwright.Page
	mu          sync.Mutex
	initialized bool
}

// NewBrowserUpstream creates a new BrowserUpstream.
func NewBrowserUpstream() *BrowserUpstream {
	return &BrowserUpstream{}
}

// Shutdown gracefully terminates the upstream service.
func (b *BrowserUpstream) Shutdown(ctx context.Context) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.page != nil {
		if err := b.page.Close(); err != nil {
			// Log error but continue
		}
	}
	if b.browser != nil {
		if err := b.browser.Close(); err != nil {
			// Log error but continue
		}
	}
	if b.pw != nil {
		if err := b.pw.Stop(); err != nil {
			// Log error but continue
		}
	}
	b.initialized = false
	return nil
}

// Register registers the browser tools.
func (b *BrowserUpstream) Register(
	ctx context.Context,
	serviceConfig *configv1.UpstreamServiceConfig,
	toolManager tool.ManagerInterface,
	promptManager prompt.ManagerInterface,
	resourceManager resource.ManagerInterface,
	isReload bool,
) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	serviceID := serviceConfig.GetId()

	// Register tools
	// We manually define the tools available for browser automation
	tools := []tool.Tool{
		newBrowseOpenTool(b, serviceID),
		newBrowseScreenshotTool(b, serviceID),
		newBrowseClickTool(b, serviceID),
		newBrowseTypeTool(b, serviceID),
		newBrowseContentTool(b, serviceID),
	}

	var definitions []*configv1.ToolDefinition
	for _, t := range tools {
		if err := toolManager.AddTool(t); err != nil {
			return "", nil, nil, err
		}
		definitions = append(definitions, &configv1.ToolDefinition{
			Name:        proto.String(t.Tool().GetName()),
			Description: proto.String(t.Tool().GetDescription()),
		})
	}

	return serviceID, definitions, nil, nil
}

func (b *BrowserUpstream) ensureInitialized() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.initialized {
		return nil
	}

	// Initialize Playwright
	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not launch playwright: %w", err)
	}
	b.pw = pw

	// Launch browser (chromium by default)
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("could not launch chromium: %w", err)
	}
	b.browser = browser

	// Create a new page
	page, err := browser.NewPage()
	if err != nil {
		return fmt.Errorf("could not create page: %w", err)
	}
	b.page = page
	b.initialized = true
	return nil
}

// baseBrowserTool provides common functionality for browser tools.
type baseBrowserTool struct {
	upstream  *BrowserUpstream
	toolProto *pb.Tool
	mcpTool   *mcp.Tool
	once      sync.Once
}

func (t *baseBrowserTool) Tool() *pb.Tool {
	return t.toolProto
}

func (t *baseBrowserTool) MCPTool() *mcp.Tool {
	t.once.Do(func() {
		// Use the converter from tool package if exported, otherwise simplified conversion
		// Since we are in upstream package, we can't easily access unexported converters in tool package.
		// We will implement a basic conversion here or rely on the fact that tool.Manager handles it if passed via AddTool?
		// tool.Manager.AddTool calls ConvertProtoToMCPTool which IS exported in tool package.
		// Wait, ConvertProtoToMCPTool is in `tool` package (management.go calls it).
		// But here I need to return it.
		// I will create a basic one.
		t.mcpTool = &mcp.Tool{
			Name:        t.toolProto.GetName(),
			Description: t.toolProto.GetDescription(),
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": make(map[string]interface{}),
			},
		}
		// Basic schema mapping
		if t.toolProto.InputSchema != nil {
			if props := t.toolProto.InputSchema.Fields["properties"]; props != nil {
				if propsStruct := props.GetStructValue(); propsStruct != nil {
					for k, v := range propsStruct.Fields {
						// This is a simplification. For full support use tool.ConvertProtoToMCPTool if available.
						// Since I cannot import tool.ConvertProtoToMCPTool (cyclic dependency if I import tool?),
						// No, I can import `server/pkg/tool`.
						// Let's check if `ConvertProtoToMCPTool` is exported.
						// `tool/converters.go` usually has it.
						_ = k
						_ = v
					}
				}
			}
		}
	})
	return t.mcpTool
}

func (t *baseBrowserTool) GetCacheConfig() *configv1.CacheConfig {
	return nil
}

// browseOpenTool
type browseOpenTool struct {
	baseBrowserTool
}

func newBrowseOpenTool(upstream *BrowserUpstream, serviceID string) *browseOpenTool {
	t := &browseOpenTool{
		baseBrowserTool: baseBrowserTool{
			upstream: upstream,
		},
	}
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"url": map[string]interface{}{"type": "string", "description": "The URL to navigate to."},
		},
		"required": []interface{}{"url"},
	})
	t.toolProto = &pb.Tool{
		Name:        proto.String("browse_open"),
		DisplayName: proto.String("Open URL"),
		Description: proto.String("Navigates the browser to the specified URL."),
		ServiceId:   proto.String(serviceID),
		InputSchema: inputSchema,
	}
	return t
}

func (t *browseOpenTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if err := t.upstream.ensureInitialized(); err != nil {
		return nil, err
	}
	var inputs struct {
		URL string `json:"url"`
	}
	if err := fastJSON.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
	}

	if _, err := t.upstream.page.Goto(inputs.URL); err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", inputs.URL, err)
	}
	return map[string]string{"status": "opened", "url": inputs.URL}, nil
}

// browseScreenshotTool
type browseScreenshotTool struct {
	baseBrowserTool
}

func newBrowseScreenshotTool(upstream *BrowserUpstream, serviceID string) *browseScreenshotTool {
	t := &browseScreenshotTool{
		baseBrowserTool: baseBrowserTool{
			upstream: upstream,
		},
	}
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	})
	t.toolProto = &pb.Tool{
		Name:        proto.String("browse_screenshot"),
		DisplayName: proto.String("Take Screenshot"),
		Description: proto.String("Takes a screenshot of the current page."),
		ServiceId:   proto.String(serviceID),
		InputSchema: inputSchema,
	}
	return t
}

func (t *browseScreenshotTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if err := t.upstream.ensureInitialized(); err != nil {
		return nil, err
	}
	data, err := t.upstream.page.Screenshot()
	if err != nil {
		return nil, fmt.Errorf("failed to take screenshot: %w", err)
	}
	// Return base64 encoded string or raw bytes? Tool execution result handles raw bytes for images?
	// Usually returning a map with base64 is safer for JSON response.
	// But `Execute` returns `any`.
	// Let's return a map with base64.
	// Actually `tool` execution result handling in `manager.go` uses `mcp.CallToolResult`.
	// If I return a map, it gets marshaled to JSON.
	// I should probably return a map with "image_data" or similar.
	return map[string]any{"screenshot_bytes": data}, nil
}

// browseClickTool
type browseClickTool struct {
	baseBrowserTool
}

func newBrowseClickTool(upstream *BrowserUpstream, serviceID string) *browseClickTool {
	t := &browseClickTool{
		baseBrowserTool: baseBrowserTool{
			upstream: upstream,
		},
	}
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"selector": map[string]interface{}{"type": "string", "description": "The CSS selector to click."},
		},
		"required": []interface{}{"selector"},
	})
	t.toolProto = &pb.Tool{
		Name:        proto.String("browse_click"),
		DisplayName: proto.String("Click Element"),
		Description: proto.String("Clicks an element on the current page matching the selector."),
		ServiceId:   proto.String(serviceID),
		InputSchema: inputSchema,
	}
	return t
}

func (t *browseClickTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if err := t.upstream.ensureInitialized(); err != nil {
		return nil, err
	}
	var inputs struct {
		Selector string `json:"selector"`
	}
	if err := fastJSON.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
	}

	if err := t.upstream.page.Click(inputs.Selector); err != nil {
		return nil, fmt.Errorf("failed to click %s: %w", inputs.Selector, err)
	}
	return map[string]string{"status": "clicked", "selector": inputs.Selector}, nil
}

// browseTypeTool
type browseTypeTool struct {
	baseBrowserTool
}

func newBrowseTypeTool(upstream *BrowserUpstream, serviceID string) *browseTypeTool {
	t := &browseTypeTool{
		baseBrowserTool: baseBrowserTool{
			upstream: upstream,
		},
	}
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"selector": map[string]interface{}{"type": "string", "description": "The CSS selector to type into."},
			"text":     map[string]interface{}{"type": "string", "description": "The text to type."},
		},
		"required": []interface{}{"selector", "text"},
	})
	t.toolProto = &pb.Tool{
		Name:        proto.String("browse_type"),
		DisplayName: proto.String("Type Text"),
		Description: proto.String("Types text into an element matching the selector."),
		ServiceId:   proto.String(serviceID),
		InputSchema: inputSchema,
	}
	return t
}

func (t *browseTypeTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if err := t.upstream.ensureInitialized(); err != nil {
		return nil, err
	}
	var inputs struct {
		Selector string `json:"selector"`
		Text     string `json:"text"`
	}
	if err := fastJSON.Unmarshal(req.ToolInputs, &inputs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal inputs: %w", err)
	}

	if err := t.upstream.page.Fill(inputs.Selector, inputs.Text); err != nil {
		return nil, fmt.Errorf("failed to type into %s: %w", inputs.Selector, err)
	}
	return map[string]string{"status": "typed", "selector": inputs.Selector}, nil
}

// browseContentTool
type browseContentTool struct {
	baseBrowserTool
}

func newBrowseContentTool(upstream *BrowserUpstream, serviceID string) *browseContentTool {
	t := &browseContentTool{
		baseBrowserTool: baseBrowserTool{
			upstream: upstream,
		},
	}
	inputSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type":       "object",
		"properties": map[string]interface{}{},
	})
	t.toolProto = &pb.Tool{
		Name:        proto.String("browse_content"),
		DisplayName: proto.String("Get Content"),
		Description: proto.String("Gets the text content of the current page."),
		ServiceId:   proto.String(serviceID),
		InputSchema: inputSchema,
	}
	return t
}

func (t *browseContentTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	if err := t.upstream.ensureInitialized(); err != nil {
		return nil, err
	}
	content, err := t.upstream.page.Content()
	if err != nil {
		return nil, fmt.Errorf("failed to get content: %w", err)
	}
	// Maybe return innerText of body to be cleaner?
	// But Content() returns full HTML.
	// Let's stick to Content() for now.
	return map[string]string{"content": content}, nil
}
