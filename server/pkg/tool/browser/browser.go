// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides a browser automation tool.
package browser

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// Provider implements a basic browser automation tool. Summary: Tool provider for browsing web pages.
//
// Summary: Provider implements a basic browser automation tool. Summary: Tool provider for browsing web pages.
//
// Fields:
//   - Contains the configuration and state properties required for Provider functionality.
type Provider struct {
}

// NewProvider creates a new Provider. Summary: Initializes a new browser provider. Returns: - *Provider: The initialized provider.
//
// Summary: NewProvider creates a new Provider. Summary: Initializes a new browser provider. Returns: - *Provider: The initialized provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - (*Provider): The resulting Provider object containing the requested data.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewProvider() *Provider {
	return &Provider{}
}

// BrowsePage simulates browsing a page using playwright-go. Summary: Fetches the content of a web page. Parameters: - ctx: context.Context. The context for the request. - url: string. The URL to visit. Returns: - string: The text content of the page. - error: An error if the URL is empty or the browser fails. Errors: - Returns "url is required" if url is empty. - Returns "failed to start playwright" or "failed to launch browser" if the browser fails to start.
//
// Summary: BrowsePage simulates browsing a page using playwright-go. Summary: Fetches the content of a web page. Parameters: - ctx: context.Context. The context for the request. - url: string. The URL to visit. Returns: - string: The text content of the page. - error: An error if the URL is empty or the browser fails. Errors: - Returns "url is required" if url is empty. - Returns "failed to start playwright" or "failed to launch browser" if the browser fails to start.
//
// Parameters:
//   - _ (context.Context): The _ parameter used in the operation.
//   - url (string): The url parameter used in the operation.
//
// Returns:
//   - (string): A string value representing the operation's result.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
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
	content = strings.TrimSpace(content)

	return content, nil
}

// ToolDefinition returns the MCP tool definition. Summary: Defines the metadata for the browse_page tool. Returns: - map[string]interface{}: The JSON schema definition of the tool.
//
// Summary: ToolDefinition returns the MCP tool definition. Summary: Defines the metadata for the browse_page tool. Returns: - map[string]interface{}: The JSON schema definition of the tool.
//
// Parameters:
//   - None.
//
// Returns:
//   - (map[string]interface): A string value representing the operation's result.
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
