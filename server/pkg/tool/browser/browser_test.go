// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
)

func TestBrowserProvider(t *testing.T) {
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		t.Skipf("skipping test, could not install playwright: %v", err)
	}

	// Verify the browser can actually launch (system deps like libnspr4 may be missing).
	pw, err := playwright.Run()
	if err != nil {
		t.Skipf("skipping test, could not start playwright: %v", err)
	}
	launchBrowser, launchErr := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(true)})
	if launchErr == nil {
		_ = launchBrowser.Close()
	}
	_ = pw.Stop()
	if launchErr != nil {
		t.Skipf("skipping test, could not launch browser (missing system dependencies?): %v", launchErr)
	}

	// Use a local HTTP test server to avoid external network access.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><head><title>Test Page</title></head><body><h1>Hello from local test server</h1></body></html>")
	}))
	defer ts.Close()

	p := NewProvider()

	def := p.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	content, err := p.BrowsePage(context.Background(), ts.URL)
	assert.NoError(t, err)
	assert.Contains(t, content, "Hello from local test server")

	_, err = p.BrowsePage(context.Background(), "")
	assert.Error(t, err)
}
