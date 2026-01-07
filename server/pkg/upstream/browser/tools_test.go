// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetTools(t *testing.T) {
	tools := getTools()
	assert.NotEmpty(t, tools)

	toolMap := make(map[string]browserToolDef)
	for _, t := range tools {
		toolMap[t.Name] = t
	}

	assert.Contains(t, toolMap, "navigate")
	assert.Contains(t, toolMap, "screenshot")
	assert.Contains(t, toolMap, "get_content")
	assert.Contains(t, toolMap, "click")
	assert.Contains(t, toolMap, "type")
	assert.Contains(t, toolMap, "evaluate")

	// Verify schemas
	nav := toolMap["navigate"]
	assert.Contains(t, nav.Input, "url")

	ss := toolMap["screenshot"]
	assert.Contains(t, ss.Input, "selector")
	assert.Contains(t, ss.Input, "full_page")

	gc := toolMap["get_content"]
	assert.Contains(t, gc.Input, "selector")
	assert.Contains(t, gc.Input, "html")
}

func TestBrowserCallable_Call(t *testing.T) {
	// We can't easily mock the Service struct because it uses chromedp internally.
	// But we can test that the wrapper passes arguments correctly if we mocked the handler,
	// but the handler is part of the struct.
	// So we mostly rely on integration tests or simply verifying the wiring code compiles and looks correct.
	// Since we don't have a mocked chromedp interface in our code (we use the library directly),
	// unit testing the handlers without a real browser is hard.
	// However, we can test the `getTools` definition validity as above.
}

// NOTE: Integration tests requiring a real browser should be tagged or skipped in CI if browser is missing.
func TestService_Integration_SkipIfNoBrowser(t *testing.T) {
	// This test attempts to launch a browser. If it fails, we skip.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	svc, err := NewService(ctx, "", true, "", 0, 0)
	if err != nil {
		t.Skipf("Skipping integration test: failed to launch browser: %v", err)
	}
	defer svc.Close()

	// Navigate
	if err := svc.Navigate(ctx, "data:text/html,<html><body><h1>Hello</h1></body></html>"); err != nil {
		t.Fatalf("Failed to navigate: %v", err)
	}

	// Get Content
	content, err := svc.GetContent(ctx, "h1", false)
	assert.NoError(t, err)
	assert.Equal(t, "Hello", content)
}
