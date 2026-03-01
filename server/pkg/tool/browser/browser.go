// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides a browser automation tool.
package browser

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
)

// Provider implements a basic browser automation tool.
//
// Summary: Tool provider for browsing web pages.
type Provider struct {
	pw *playwright.Playwright
	bw playwright.Browser
	mu sync.Mutex
}

// NewProvider creates a new Provider.
//
// Summary: Initializes a new browser provider.
//
// Returns:
//   - *Provider: The initialized provider.
func NewProvider() *Provider {
	return &Provider{}
}

func (b *Provider) initPlaywright() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.pw != nil && b.bw != nil {
		return nil
	}

	err := playwright.Install()
	if err != nil {
		return fmt.Errorf("could not install playwright: %w", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("could not start playwright: %w", err)
	}

	bw, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		_ = pw.Stop()
		return fmt.Errorf("could not launch browser: %w", err)
	}

	b.pw = pw
	b.bw = bw
	return nil
}

// Close gracefully closes the browser and Playwright instances.
//
// Summary: Closes the browser and stops Playwright.
//
// Returns:
//   - error: An error if closing fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Stops the Playwright browser instance.
func (b *Provider) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.bw != nil {
		_ = b.bw.Close()
	}
	if b.pw != nil {
		_ = b.pw.Stop()
	}
	return nil
}

func isSafeURL(targetURL string) error {
	u, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: %s, only http/https allowed", u.Scheme)
	}

	host := u.Hostname()

	// Check for local loopback or common internal names
	if host == "localhost" || host == "127.0.0.1" || host == "::1" || strings.HasSuffix(host, ".local") {
		return fmt.Errorf("URL points to an internal/local resource: %s", host)
	}

	// Resolve the IP addresses
	ips, err := net.LookupIP(host)
	if err != nil {
		return fmt.Errorf("could not resolve hostname: %w", err)
	}

	for _, ip := range ips {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() || ip.IsUnspecified() {
			return fmt.Errorf("URL resolves to an internal/local IP address: %s", ip.String())
		}
	}

	return nil
}

// BrowsePage simulates browsing a page
//
// Summary: Fetches the content of a web page using Playwright.
//
// Parameters:
//   - _: context.Context. Unused.
//   - targetURL: string. The URL to visit.
//
// Returns:
//   - string: The HTML content of the page.
//   - error: An error if the URL is empty or the page cannot be loaded.
//
// Errors:
//   - Returns "url is required" if targetURL is empty.
//   - Returns an error if the URL is pointing to a local or internal network.
//   - Returns an error if Playwright initialization fails.
//   - Returns an error if the page fails to load.
//
// Side Effects:
//   - Starts Playwright on first use.
//   - Navigates the browser to the specified URL.
func (b *Provider) BrowsePage(_ context.Context, targetURL string) (string, error) {
	if targetURL == "" {
		return "", fmt.Errorf("url is required")
	}

	if err := isSafeURL(targetURL); err != nil {
		return "", fmt.Errorf("security policy violation: %w", err)
	}

	if err := b.initPlaywright(); err != nil {
		return "", fmt.Errorf("failed to initialize browser: %w", err)
	}

	b.mu.Lock()
	bw := b.bw
	b.mu.Unlock()

	page, err := bw.NewPage()
	if err != nil {
		return "", fmt.Errorf("could not create page: %w", err)
	}
	defer page.Close()

	if _, err = page.Goto(targetURL); err != nil {
		return "", fmt.Errorf("could not goto: %w", err)
	}

	content, err := page.Content()
	if err != nil {
		return "", fmt.Errorf("could not get content: %w", err)
	}

	return content, nil
}

// ToolDefinition returns the MCP tool definition.
//
// Summary: Defines the metadata for the browse_page tool.
//
// Returns:
//   - map[string]interface{}: The JSON schema definition of the tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
