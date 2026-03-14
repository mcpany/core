// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides a browser automation tool.
// Provider implements a basic browser automation tool.
//
// Summary: Tool provider for browsing web pages.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// NewProvider creates a new Provider.
//
// Summary: Initializes a new browser provider.
//
// Returns:
//   - *Provider: The initialized provider.
//
//
// Errors:
//   - An error if it fails.
//
// Side Effects:
//   - None.
// BrowsePage simulates browsing a page using playwright-go.
//
// Summary: Fetches the content of a web page.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - url: string. The URL to visit.
//
// Returns:
//   - string: The text content of the page.
//   - error: An error if the URL is empty or the browser fails.
//
// Errors:
//   - Returns "url is required" if url is empty.
//   - Returns "failed to start playwright" or "failed to launch browser" if the browser fails to start.
//
//
// Side Effects:
//   - None.
package browser

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

type Provider struct {
}

func NewProvider() *Provider {
	return &Provider{}
}

func (b *Provider) BrowsePage(_ context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is required")
	}

	pw, err := playwright.Run()
	if err != nil {
		return "", fmt.Errorf("could not start playwright: %w", err)
	}
	defer func() {
		if err := pw.Stop(); err != nil {
			log.Printf("could not stop playwright: %v", err)
		}
	}()

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(true),
	})
	if err != nil {
		return "", fmt.Errorf("could not launch browser: %w", err)
	}
	defer func() {
		if err := browser.Close(); err != nil {
			log.Printf("could not close browser: %v", err)
		}
	}()

	page, err := browser.NewPage()
	if err != nil {
		return "", fmt.Errorf("could not create page: %w", err)
	}

	if _, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); err != nil {
		return "", fmt.Errorf("could not goto: %w", err)
	}

	content, err := page.Locator("body").TextContent()
	if err != nil {
		return "", fmt.Errorf("could not extract text content: %w", err)
	}

	// Clean up content slightly
	// ToolDefinition returns the MCP tool definition.
	//
	// Summary: Defines the metadata for the browse_page tool.
	//
	// Returns:
	//   - map[string]interface{}: The JSON schema definition of the tool.
	//
	//
	// Errors:
	//   - An error if it fails.
	//
	// Side Effects:
	//   - None.
	content = strings.TrimSpace(content)

	return content, nil
}

func (b *Provider) ToolDefinition() map[string]interface{} {
	return map[string]interface{}{
		"name":		"browse_page",
		"description":	"Visit a webpage and return its content",
		"inputSchema": map[string]interface{}{
			"type":	"object",
			"properties": map[string]interface{}{
				"url": map[string]interface{}{
					"type":		"string",
					"description":	"The URL to visit",
				},
			},
			"required":	[]string{"url"},
		},
	}
}
