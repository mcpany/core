// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides a browser automation tool.
package browser

import (
	"context"
	"fmt"
// PageFetcher fetches the visible text content of a URL.
// It is an interface so tests can inject a lightweight implementation without
// requiring a real browser installation.
//
// Summary: Represents a PageFetcher.
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
type PageFetcher interface {
	// FetchText retrieves the text content of a URL.
	//
	// Summary: Retrieves the text content of a URL.
	//
	// Parameters:
	//   - ctx: context.Context. The context for the request.
// Provider implements a basic browser automation tool.
// BrowsePage fetches the text content of the given URL.
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
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Errors:
//   - Returns "url is required" if url is empty.
//   - Returns "failed to start playwright" or "failed to launch browser" if the browser fails to start.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (b *Provider) BrowsePage(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("url is required")
	}
	f := b.fetcher
	if f == nil {
		f = &playwrightFetcher{}
	}
// ToolDefinition returns the MCP tool definition.
//
// Summary: Defines the metadata for the browse_page tool.
//
// Returns:
//   - map[string]interface{}: The JSON schema definition of the tool.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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

// playwrightRunner abstracts playwright execution so we can inject a mock for hermetic testing.
type playwrightRunner interface {
	Run() (playwrightImpl, error)
}

type playwrightImpl interface {
	Stop() error
	Chromium() playwrightBrowserType
}

type playwrightBrowserType interface {
	Launch(options ...playwright.BrowserTypeLaunchOptions) (playwrightBrowser, error)
}

type playwrightBrowser interface {
	Close() error
// Run starts the playwright instance.
//
// Summary: Starts playwright.
//
// Parameters:
//
// Stop stops the playwright instance.
//
// Summary: Stops playwright.
//
// Returns:
//   - error: An error if stopping fails.
//
// Errors:
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - Returns error if any.
// Launch launches the browser.
//
// Summary: Launches the browser.
//
// Parameters:
//   - options: ...playwright.BrowserTypeLaunchOptions. Options to launch the browser.
//
// Returns:
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// NewPage creates a new page in the browser.
//
// Summary: Creates a new page.
//
// Parameters:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - options: ...playwright.BrowserNewPageOptions. Options for the new page.
//
// Returns:
// Goto navigates to a URL.
//
// Summary: Navigates to a URL.
//
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
// Parameters:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - url: string. The URL to navigate to.
//   - options: ...playwright.PageGotoOptions. Options for navigation.
// Locator creates a new locator.
//
// TextContent retrieves the text content of the locator.
//
// Summary: Retrieves text content.
//
// Parameters:
// Returns:
//   - execution result or state changes.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - options: ...playwright.LocatorTextContentOptions. Options for retrieval.
//
// Returns:
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
//   - string: The text content.
//   - error: An error if retrieval fails.
//
// Errors:
//   - Returns error if any.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (r *realLocator) TextContent(options ...playwright.LocatorTextContentOptions) (string, error) {
	return r.l.TextContent(options...)
}

// playwrightFetcher is the production PageFetcher that uses playwright-go.
type playwrightFetcher struct {
	runner playwrightRunner
}

// FetchText fetches the text content of a URL using playwright.
//
// Summary: Fetches the text content of a URL using playwright.
//
// Parameters:
//   - ctx: context.Context. The context for the request.
//   - url: string. The URL to visit.
//
// Returns:
//   - string: The text content of the page.
//   - error: An error if the fetch fails.
//
// Errors:
//   - Returns error if any.
//
// Side Effects:
//   - None.
func (f *playwrightFetcher) FetchText(_ context.Context, url string) (string, error) {
	r := f.runner
	if r == nil {
		r = &defaultPlaywrightRunner{}
	}

	pw, err := r.Run()
	if err != nil {
		return "", fmt.Errorf("could not start playwright: %w", err)
	}
	defer func() {
		if err := pw.Stop(); err != nil {
			log.Printf("could not stop playwright: %v", err)
		}
	}()

	browser, err := pw.Chromium().Launch(playwright.BrowserTypeLaunchOptions{
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
