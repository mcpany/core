// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides a browser automation tool.
package browser

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/mcpany/core/server/pkg/util"
)

// Provider implements a basic browser automation tool.
type Provider struct {
	client *http.Client
}

// Option defines a functional option for Provider.
type Option func(*Provider)

// WithClient sets a custom HTTP client for the provider.
func WithClient(client *http.Client) Option {
	return func(p *Provider) {
		p.client = client
	}
}

// NewProvider creates a new Provider.
//
// Returns the result.
func NewProvider(opts ...Option) *Provider {
	// Create a custom transport to prevent SSRF
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			// Resolve IPs using context-aware lookup
			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}

			// Check IPs and dial the first safe one
			for _, ipAddr := range ips {
				if util.IsPrivateIP(ipAddr.IP) {
					continue
				}

				// Safe IP found, dial specific IP to prevent DNS rebinding
				conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ipAddr.IP.String(), port))
				if err == nil {
					return conn, nil
				}
				// If dial fails, try next IP (standard behavior)
			}

			// If we exhausted all IPs or they were all private
			return nil, fmt.Errorf("no safe public IP found for %s", host)
		},
	}

	p := &Provider{
		client: &http.Client{
			Transport: transport,
			Timeout:   30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// BrowsePage simulates browsing a page.
// It fetches the URL and converts the HTML content to Markdown.
func (b *Provider) BrowsePage(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set a User-Agent to avoid being blocked by some sites
	req.Header.Set("User-Agent", "MCP-Any-Browser/1.0")

	resp, err := b.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("request failed with status code: %d", resp.StatusCode)
	}

	// Limit response body to 5MB to prevent memory exhaustion
	limitedBody := io.LimitReader(resp.Body, 5*1024*1024)

	// Read body
	bodyBytes, err := io.ReadAll(limitedBody)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Create converter
	converter := md.NewConverter("", true, nil)

	// Convert to markdown
	markdown, err := converter.ConvertString(string(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to convert html to markdown: %w", err)
	}

	// Add metadata header
	result := fmt.Sprintf("# Browsed: %s\n\n%s", url, markdown)
	return result, nil
}

// ToolDefinition returns the MCP tool definition.
//
// Returns the result.
func (b *Provider) ToolDefinition() map[string]interface{} {
	return map[string]interface{}{
		"name":        "browse_page",
		"description": "Visit a webpage and return its content as Markdown",
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
