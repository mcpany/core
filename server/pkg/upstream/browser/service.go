// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package browser provides the browser upstream implementation.
package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/chromedp/chromedp"
)

// Service manages the browser session.
type Service struct {
	allocCtx   context.Context
	cancel     context.CancelFunc
	browserCtx context.Context
}

// NewService creates a new browser service.
func NewService(ctx context.Context, endpoint string, headless bool, userAgent string, width, height int) (*Service, error) {
	var opts []chromedp.ExecAllocatorOption

	if endpoint != "" {
		allocCtx, cancel := chromedp.NewRemoteAllocator(ctx, endpoint)

		// Initialize the browser context immediately to avoid race conditions later.
		bCtx, _ := chromedp.NewContext(allocCtx)

		// Run a dummy action to ensure connection?
		// Not strictly necessary, but good to fail fast if endpoint is bad.
		// However, NewContext doesn't connect yet. Run does.
		// Let's just return the context. The first action will trigger connection.

		return &Service{
			allocCtx:   allocCtx,
			cancel:     cancel,
			browserCtx: bCtx,
		}, nil
	}

	opts = append(opts, chromedp.DefaultExecAllocatorOptions[:]...)
	if !headless {
		opts = append(opts, chromedp.Flag("headless", false))
	}
	if userAgent != "" {
		opts = append(opts, chromedp.UserAgent(userAgent))
	}
	if width > 0 && height > 0 {
		opts = append(opts, chromedp.WindowSize(width, height))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)

	// Create a browser context.
	// We create one context for the service to keep the browser alive.
	bCtx, _ := chromedp.NewContext(allocCtx)
	if err := chromedp.Run(bCtx); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	return &Service{
		allocCtx:   allocCtx,
		cancel:     cancel,
		browserCtx: bCtx,
	}, nil
}

// Close cleans up the browser resources.
func (s *Service) Close() {
	if s.cancel != nil {
		s.cancel()
	}
}

// GetContext returns a new context for executing tasks.
func (s *Service) GetContext(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	// browserCtx is now guaranteed to be initialized in NewService.
	// Create a child context from the persistent browser context.
	// This ensures all actions share the same browser session/target unless we explicitly create new targets.
	// For "Single User" mode, this is correct.

	// Note: If we wanted to support multiple tabs, we would use NewContext(s.browserCtx) here.
	// But `chromedp` default behavior with Run(ctx) where ctx is same target queue actions?
	// If we want a NEW target (tab) per tool call? No, usually we want to control the SAME tab to navigate then click.
	// So we should reuse the context?
	// If we reuse `s.browserCtx`, we can't attach a per-request timeout easily without cancelling the parent?
	// Actually `chromedp.Run` takes a context.
	// If we do `ctx, cancel = context.WithTimeout(s.browserCtx, timeout)`, cancelling this `ctx`
	// does not kill the browser, only the actions associated with it?
	// Yes, `s.browserCtx` is the parent.

	if timeout > 0 {
		return context.WithTimeout(s.browserCtx, timeout)
	}
	return context.WithCancel(s.browserCtx)
}

// Navigate navigates to a URL.
func (s *Service) Navigate(ctx context.Context, url string) error {
	return chromedp.Run(ctx, chromedp.Navigate(url))
}

// Screenshot takes a screenshot.
func (s *Service) Screenshot(ctx context.Context, sel string, fullPage bool) ([]byte, error) {
	var buf []byte
	var tasks []chromedp.Action
	if fullPage {
		tasks = append(tasks, chromedp.FullScreenshot(&buf, 100))
	} else {
		if sel == "" {
			tasks = append(tasks, chromedp.CaptureScreenshot(&buf))
		} else {
			tasks = append(tasks, chromedp.Screenshot(sel, &buf, chromedp.NodeVisible))
		}
	}
	if err := chromedp.Run(ctx, tasks...); err != nil {
		return nil, err
	}
	return buf, nil
}

// GetContent gets the text content of a selector or the whole page.
func (s *Service) GetContent(ctx context.Context, sel string, html bool) (string, error) {
	var res string
	var tasks []chromedp.Action
	if sel == "" {
		sel = "body"
	}

	if html {
		tasks = append(tasks, chromedp.OuterHTML(sel, &res))
	} else {
		tasks = append(tasks, chromedp.Text(sel, &res))
	}

	if err := chromedp.Run(ctx, tasks...); err != nil {
		return "", err
	}
	return res, nil
}

// Click clicks on a selector.
func (s *Service) Click(ctx context.Context, sel string) error {
	return chromedp.Run(ctx, chromedp.Click(sel))
}

// Type types text into a selector.
func (s *Service) Type(ctx context.Context, sel, text string) error {
	return chromedp.Run(ctx, chromedp.SendKeys(sel, text))
}

// Evaluate evaluates JavaScript.
func (s *Service) Evaluate(ctx context.Context, expression string) (any, error) {
	var res any
	if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &res)); err != nil {
		return nil, err
	}
	return res, nil
}
