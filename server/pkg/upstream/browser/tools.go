// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

type browserToolDef struct {
	Name        string
	Description string
	Input       map[string]interface{}
	Output      map[string]interface{}
	Handler     func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error)
}

func getTools() []browserToolDef {
	return []browserToolDef{
		{
			Name:        "navigate",
			Description: "Navigate to a URL.",
			Input: map[string]interface{}{
				"url": map[string]interface{}{"type": "string", "description": "The URL to navigate to."},
			},
			Output: map[string]interface{}{
				"success": map[string]interface{}{"type": "boolean"},
			},
			Handler: func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
				url, ok := args["url"].(string)
				if !ok {
					return nil, fmt.Errorf("url is required")
				}
				c, cancel := s.GetContext(ctx, timeout)
				defer cancel()
				if err := s.Navigate(c, url); err != nil {
					return nil, err
				}
				return map[string]interface{}{"success": true}, nil
			},
		},
		{
			Name:        "screenshot",
			Description: "Take a screenshot of the current page.",
			Input: map[string]interface{}{
				"selector":  map[string]interface{}{"type": "string", "description": "CSS selector to screenshot. If empty, takes viewport."},
				"full_page": map[string]interface{}{"type": "boolean", "description": "Whether to capture the full scrollable page."},
			},
			Output: map[string]interface{}{
				"image": map[string]interface{}{"type": "string", "description": "Base64 encoded PNG image."},
			},
			Handler: func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
				sel, _ := args["selector"].(string)
				fullPage, _ := args["full_page"].(bool)

				c, cancel := s.GetContext(ctx, timeout)
				defer cancel()

				buf, err := s.Screenshot(c, sel, fullPage)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"image": base64.StdEncoding.EncodeToString(buf)}, nil
			},
		},
		{
			Name:        "get_content",
			Description: "Get text or HTML content from the page.",
			Input: map[string]interface{}{
				"selector": map[string]interface{}{"type": "string", "description": "CSS selector. Default is body."},
				"html":     map[string]interface{}{"type": "boolean", "description": "Return HTML instead of text."},
			},
			Output: map[string]interface{}{
				"content": map[string]interface{}{"type": "string"},
			},
			Handler: func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
				sel, _ := args["selector"].(string)
				html, _ := args["html"].(bool)

				c, cancel := s.GetContext(ctx, timeout)
				defer cancel()

				content, err := s.GetContent(c, sel, html)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"content": content}, nil
			},
		},
		{
			Name:        "click",
			Description: "Click an element.",
			Input: map[string]interface{}{
				"selector": map[string]interface{}{"type": "string", "description": "CSS selector to click."},
			},
			Output: map[string]interface{}{
				"success": map[string]interface{}{"type": "boolean"},
			},
			Handler: func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
				sel, ok := args["selector"].(string)
				if !ok {
					return nil, fmt.Errorf("selector is required")
				}
				c, cancel := s.GetContext(ctx, timeout)
				defer cancel()

				if err := s.Click(c, sel); err != nil {
					return nil, err
				}
				return map[string]interface{}{"success": true}, nil
			},
		},
		{
			Name:        "type",
			Description: "Type text into an element.",
			Input: map[string]interface{}{
				"selector": map[string]interface{}{"type": "string", "description": "CSS selector."},
				"text":     map[string]interface{}{"type": "string", "description": "Text to type."},
			},
			Output: map[string]interface{}{
				"success": map[string]interface{}{"type": "boolean"},
			},
			Handler: func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
				sel, ok := args["selector"].(string)
				if !ok {
					return nil, fmt.Errorf("selector is required")
				}
				text, ok := args["text"].(string)
				if !ok {
					return nil, fmt.Errorf("text is required")
				}

				c, cancel := s.GetContext(ctx, timeout)
				defer cancel()

				if err := s.Type(c, sel, text); err != nil {
					return nil, err
				}
				return map[string]interface{}{"success": true}, nil
			},
		},
		{
			Name:        "evaluate",
			Description: "Evaluate JavaScript on the page.",
			Input: map[string]interface{}{
				"expression": map[string]interface{}{"type": "string", "description": "JavaScript expression."},
			},
			Output: map[string]interface{}{
				"result": map[string]interface{}{"type": "any"}, // Type is arbitrary
			},
			Handler: func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error) {
				expr, ok := args["expression"].(string)
				if !ok {
					return nil, fmt.Errorf("expression is required")
				}

				c, cancel := s.GetContext(ctx, timeout)
				defer cancel()

				res, err := s.Evaluate(c, expr)
				if err != nil {
					return nil, err
				}
				return map[string]interface{}{"result": res}, nil
			},
		},
	}
}

type browserCallable struct {
	handler func(ctx context.Context, s *Service, args map[string]interface{}, timeout time.Duration) (map[string]interface{}, error)
	service *Service
	timeout time.Duration
}

func (c *browserCallable) Call(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return c.handler(ctx, c.service, req.Arguments, c.timeout)
}

func createTools(serviceID string, s *Service, timeout time.Duration, toolManager tool.ManagerInterface, serviceConfig *configv1.UpstreamServiceConfig) ([]*configv1.ToolDefinition, error) {
	discoveredTools := make([]*configv1.ToolDefinition, 0)
	for _, t := range getTools() {
		inputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Input,
		})
		outputSchema, _ := structpb.NewStruct(map[string]interface{}{
			"type":       "object",
			"properties": t.Output,
		})

		toolDef := configv1.ToolDefinition_builder{
			Name:      proto.String(t.Name),
			ServiceId: proto.String(serviceID),
		}.Build()

		callable := &browserCallable{
			handler: t.Handler,
			service: s,
			timeout: timeout,
		}

		callableTool, err := tool.NewCallableTool(toolDef, serviceConfig, callable, inputSchema, outputSchema)
		if err != nil {
			return nil, err
		}

		if err := toolManager.AddTool(callableTool); err != nil {
			return nil, err
		}
		discoveredTools = append(discoveredTools, toolDef)
	}
	return discoveredTools, nil
}
