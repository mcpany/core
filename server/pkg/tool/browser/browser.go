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
// Summary: Represents a PageFetcher.
// PageFetcher fetches the visible text content of a URL.
// It is an interface so tests can inject a lightweight implementation without
// requiring a real browser installation.
// Summary: Represents a PageFetcher.
	// FetchText retrieves the text content of a URL.
	// Summary: Retrieves the text content of a URL.
	// Parameters:
	//   - ctx: context.Context. The context for the request.
	//   - url: string. The URL to visit.
	// Returns:
	//   - string: The text content of the page.
	//   - error: An error if the fetch fails.
	// Errors:
	//   - Returns error if any.
	// Side Effects:
	//   - None.
	FetchText(ctx context.Context, url string) (string, error)
}

// Provider implements a basic browser automation tool.
// Summary: Tool provider for browsing web pages.
// Provider implements a basic browser automation tool.
// Summary: Tool provider for browsing web pages.
	fetcher PageFetcher // nil → default playwrightFetcher
}

// NewProvider creates a new Provider.
// Summary: Initializes a new browser provider.
// Returns:
//   - *Provider: The initialized provider.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// NewProvider creates a new Provider.
// Summary: Initializes a new browser provider.
// Returns:
//   - *Provider: The initialized provider.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
	return &Provider{}
}

// BrowsePage fetches the text content of the given URL.
// Summary: Fetches the content of a web page.
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
// Side Effects:
//   - None.
// BrowsePage fetches the text content of the given URL.
// Summary: Fetches the content of a web page.
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
// Side Effects:
//   - None.
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

// ToolDefinition returns the MCP tool definition.
// Summary: Defines the metadata for the browse_page tool.
// Returns:
//   - map[string]interface{}: The JSON schema definition of the tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// ToolDefinition returns the MCP tool definition.
// Summary: Defines the metadata for the browse_page tool.
// Returns:
//   - map[string]interface{}: The JSON schema definition of the tool.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
// Parameters:
//   - None.
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
// Summary: Fetches the text content of a URL using playwright.
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
// FetchText fetches the text content of a URL using playwright.
// Summary: Fetches the text content of a URL using playwright.
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
