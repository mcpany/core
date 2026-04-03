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

// PageFetcher represents the public PageFetcher entity.
//
// Summary: Defines the structured data model representing a fetcher.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// Provider represents the public Provider entity.
//
// Summary: Defines the structured data model representing a .
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Provider struct {
	fetcher PageFetcher // nil → default playwrightFetcher
}

// NewProvider serves as a public interface for interacting with NewProvider.
//
// Summary: Constructs and returns an initialized provider ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func NewProvider() *Provider {
	return &Provider{}
}

// BrowsePage serves as a public interface for interacting with BrowsePage.
//
// Summary: Browse the page appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// ToolDefinition serves as a public interface for interacting with ToolDefinition.
//
// Summary: Tool the definition appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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

// Run serves as a public interface for interacting with Run.
//
// Summary: Run the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (d *defaultPlaywrightRunner) Run() (playwrightImpl, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, err
	}
	return &realPlaywright{pw}, nil
}

type realPlaywright struct{ pw *playwright.Playwright }

// Stop serves as a public interface for interacting with Stop.
//
// Summary: Stop the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realPlaywright) Stop() error { return r.pw.Stop() }

// Chromium serves as a public interface for interacting with Chromium.
//
// Summary: Chromium the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realPlaywright) Chromium() playwrightBrowserType {
	return &realBrowserType{r.pw.Chromium}
}

type realBrowserType struct{ bt playwright.BrowserType }

// Launch serves as a public interface for interacting with Launch.
//
// Summary: Launch the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realBrowserType) Launch(options ...playwright.BrowserTypeLaunchOptions) (playwrightBrowser, error) {
	b, err := r.bt.Launch(options...)
	if err != nil {
		return nil, err
	}
	return &realBrowser{b}, nil
}

type realBrowser struct{ b playwright.Browser }

// Close serves as a public interface for interacting with Close.
//
// Summary: Close the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realBrowser) Close() error { return r.b.Close() }

// NewPage serves as a public interface for interacting with NewPage.
//
// Summary: Constructs and returns an initialized page ready for consumption.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realBrowser) NewPage(options ...playwright.BrowserNewPageOptions) (playwrightPage, error) {
	p, err := r.b.NewPage(options...)
	if err != nil {
		return nil, err
	}
	return &realPage{p}, nil
}

type realPage struct{ p playwright.Page }

// Goto serves as a public interface for interacting with Goto.
//
// Summary: Goto the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realPage) Goto(url string, options ...playwright.PageGotoOptions) (playwright.Response, error) {
	return r.p.Goto(url, options...)
}

// Locator serves as a public interface for interacting with Locator.
//
// Summary: Locator the  appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the successfully computed domain model or execution state.
//
// Errors:
//   - No explicit errors are thrown by this operation.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realPage) Locator(selector string, options ...playwright.PageLocatorOptions) playwrightLocator {
	return &realLocator{r.p.Locator(selector, options...)}
}

type realLocator struct{ l playwright.Locator }

// TextContent serves as a public interface for interacting with TextContent.
//
// Summary: Text the content appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
func (r *realLocator) TextContent(options ...playwright.LocatorTextContentOptions) (string, error) {
	return r.l.TextContent(options...)
}

// playwrightFetcher is the production PageFetcher that uses playwright-go.
type playwrightFetcher struct {
	runner playwrightRunner
}

// FetchText serves as a public interface for interacting with FetchText.
//
// Summary: Fetch the text appropriately based on current system conditions.
//
// Parameters:
//   - Refer to the function signature for strongly-typed input arguments.
//
// Returns:
//   - Returns the expected domain model and an error upon failure.
//
// Errors:
//   - Propagates exceptions from underlying I/O or validation layers.
//
// Side Effects:
//   - May safely mutate local state without unintended external side effects.
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
