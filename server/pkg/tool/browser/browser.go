// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides a browser automation tool.
package browser

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mcpany/core/server/pkg/util"
)

// Provider implements a basic browser automation tool.
type Provider struct {
	client *http.Client
}

// Option defines a functional option for the Provider.
type Option func(*Provider)

// WithClient sets a custom HTTP client for the Provider.
func WithClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

// NewProvider creates a new Provider.
//
// Returns the result.
func NewProvider(opts ...Option) *Provider {
	// Use SafeDialer to prevent SSRF (blocks private/loopback IPs)
	safeDialer := util.NewSafeDialer()

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           safeDialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	p := &Provider{
		client: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
	}

	for _, opt := range opts {
		opt(p)
	}
	return p
}

// BrowsePage visits a webpage and returns its content as Markdown.
func (b *Provider) BrowsePage(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is required")
	}

	// Ensure URL has a scheme
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent to mimic a real browser to avoid some 403s
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; McpAnyBot/1.0; +https://github.com/mcpany/core)")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("HTTP request failed with status code %d", resp.StatusCode)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read page content: %w", err)
	}

	converter := md.NewConverter("", true, nil)
	markdown, err := converter.ConvertString(string(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to convert HTML to markdown: %w", err)
	}

	return markdown, nil
}

// ToolDefinition returns the MCP tool definition.
//
// Returns the result.
func (b *Provider) ToolDefinition() map[string]interface{} {
	return map[string]interface{}{
		"name":        "browse_page",
		"description": "Visit a webpage and return its content",
		"inputSchema": map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":        "string",
					"description": "The URL to visit",
				},
			},
			"required": []string{"url"},
		},
	}
}
