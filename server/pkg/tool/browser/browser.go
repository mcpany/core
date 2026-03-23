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
// Summary: Represents a PageFetcher.
type PageFetcher interface {
	// FetchText retrieves the text content of a URL.
	//
	// Summary: Retrieves the text content of a URL.
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
	FetchText(ctx context.Context, url string) (string, error)
}

// Provider implements a basic browser automation tool.
//
// Summary: Tool provider for browsing web pages.
type Provider struct {
	fetcher PageFetcher // nil → default playwrightFetcher
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

// ToolDefinition returns the MCP tool definition.
//
// Summary: Defines the metadata for the browse_page tool.
//
// Returns:
//   - map[string]interface{}: The JSON schema definition of the tool.
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
	NewPage(options ...playwright.BrowserNewPageOptions) (playwrightPage, error)
}

type playwrightPage interface {
	Goto(url string, options ...playwright.PageGotoOptions) (playwright.Response, error)
	Locator(selector string, options ...playwright.PageLocatorOptions) playwrightLocator
}

type playwrightLocator interface {
	TextContent(options ...playwright.LocatorTextContentOptions) (string, error)
}

// defaultPlaywrightRunner uses actual playwright-go
type defaultPlaywrightRunner struct{}

func (d *defaultPlaywrightRunner) Run() (playwrightImpl, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	return &realPlaywright{pw}, nil
}

type realPlaywright struct{ pw *playwright.Playwright }

func (r *realPlaywright) Stop() error { return r.pw.Stop() }
func (r *realPlaywright) Chromium() playwrightBrowserType {
	return &realBrowserType{r.pw.Chromium}
}

type realBrowserType struct{ bt playwright.BrowserType }

func (r *realBrowserType) Launch(options ...playwright.BrowserTypeLaunchOptions) (playwrightBrowser, error) {
	b, err := r.bt.Launch(options...)
	if err != nil {
		return nil, err
	}
	return &realBrowser{b}, nil
}

type realBrowser struct{ b playwright.Browser }

func (r *realBrowser) Close() error { return r.b.Close() }
func (r *realBrowser) NewPage(options ...playwright.BrowserNewPageOptions) (playwrightPage, error) {
	p, err := r.b.NewPage(options...)
	if err != nil {
		return nil, err
	}
	return &realPage{p}, nil
}

type realPage struct{ p playwright.Page }

func (r *realPage) Goto(url string, options ...playwright.PageGotoOptions) (playwright.Response, error) {
	return r.p.Goto(url, options...)
}
func (r *realPage) Locator(selector string, options ...playwright.PageLocatorOptions) playwrightLocator {
	return &realLocator{r.p.Locator(selector, options...)}
}

type realLocator struct{ l playwright.Locator }

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
