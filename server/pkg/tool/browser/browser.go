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

// PageFetcher pageFetcher represents a page fetcher.
//
// Summary: PageFetcher represents a page fetcher.
type PageFetcher interface {
	// FetchText retrieves the text content of a URL.
	//
	// Summary: Retrieves the text content of a URL.
	//
	// Parameters: - None.
	//   - ctx: context.Context. The context for the request.
	//   - url: string. The URL to visit.
	//
	// Returns: - None.
	//   - string: The text content of the page.
	//   - error: An error if the fetch fails.
	//
	// Errors: - None.
	//   - Returns error if any.
	//
	// Side Effects: - None.
	//   - None.
	FetchText(ctx context.Context, url string) (string, error)
}

// Provider provider represents a provider.
//
// Summary: Provider represents a provider.
type Provider struct {
	fetcher PageFetcher // nil → default playwrightFetcher
}

// NewProvider creates a new provider.
//
// Summary: Creates a new provider.
//
// Parameters:
//   - None.
//
// Returns:
//   - *Provider: The result.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewProvider() *Provider {
	return &Provider{}
}

// BrowsePage fetches the text content of the given URL.
//
// Summary: Fetches the content of a web page.
//
// Parameters: - None.
//   - ctx: context.Context. The context for the request.
//   - url: string. The URL to visit.
//
// Returns: - None.
//   - string: The text content of the page.
//   - error: An error if the URL is empty or the browser fails.
//
// Errors: - None.
//   - Returns "url is required" if url is empty.
//   - Returns "failed to start playwright" or "failed to launch browser" if the browser fails to start.
func (b *Provider) BrowsePage(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	f := b.fetcher
	if f == nil {
		f = &playwrightFetcher{}
	}
	content, err := f.FetchText(ctx, url)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(content), nil
}

// ToolDefinition toolDefinition tool definition.
//
// Summary: ToolDefinition tool definition.
//
// Parameters:
//   - None.
//
// Returns:
//   - map[string]interface: The result.
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

// playwrightFetcher is the production PageFetcher that uses playwright-go.
type playwrightFetcher struct{}

// FetchText fetches the text content of a URL using playwright.
//
// Summary: Fetches the text content of a URL using playwright.
//
// Parameters: - None.
//   - ctx: context.Context. The context for the request.
//   - url: string. The URL to visit.
//
// Returns: - None.
//   - string: The text content of the page.
//   - error: An error if the fetch fails.
//
// Errors: - None.
//   - Returns error if any.
//
// Side Effects: - None.
//   - None.
func (f *playwrightFetcher) FetchText(_ context.Context, url string) (string, error) {
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

	return content, nil
}
