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

// PageFetcher fetches the visible text content of a URL.
// It is an interface so tests can inject a lightweight implementation without
// requiring a real browser installation.
//
// Summary. Represents a PageFetcher.
type PageFetcher interface {
	// FetchText retrieves the text content of a URL.
	//
	// Summary: Retrieves the text content of a URL.
	//
// Parameters.
	//   - ctx: context.Context. The context for the request.
	//   - url: string. The URL to visit.
	//
// Returns.
	//   - string: The text content of the page.
	//   - error: An error if the fetch fails.
	//
// Errors:
	//   - Returns error if any.
	//
// Side Effects:
	//   - None.
	FetchText(ctx context.Context, url string) (string, error)
}

// Provider implements a basic browser automation tool.
//
// Summary. Tool provider for browsing web pages.
type Provider struct {
	fetcher PageFetcher // nil → default playwrightFetcher
}

// NewProvider provides newprovider functionality.
//
// Summary: NewProvider.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
func NewProvider() *Provider {
	return &Provider{}
}

// BrowsePage provides browsepage functionality.
//
// Summary: BrowsePage.
//
// Parameters.
//   - ctx: The parameter.
//   - url: The parameter.
//   - error: The parameter.
//
// Returns.
//   - None.
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

// ToolDefinition provides tooldefinition functionality.
//
// Summary: ToolDefinition.
//
// Parameters.
//   - None.
//
// Returns.
//   - result: The result.
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

// FetchText provides fetchtext functionality.
//
// Summary: FetchText.
//
// Parameters.
//   - _: The parameter.
//   - url: The parameter.
//   - error: The parameter.
//
// Returns.
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
