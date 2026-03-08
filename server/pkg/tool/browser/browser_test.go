// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"os"
	"testing"

	"github.com/playwright-community/playwright-go"
	"github.com/stretchr/testify/assert"
)

func TestBrowserProvider(t *testing.T) {
	if os.Getenv("CI") != "" || os.Getenv("GITHUB_ACTIONS") != "" {
		t.Skip("Skipping browser test in CI environment without guaranteed playwright driver.")
	}

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

	p := NewProvider()

	def := p.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	content, err := p.BrowsePage(context.Background(), "https://example.com")
	assert.NoError(t, err)
	assert.Contains(t, content, "Example Domain")

	_, err = p.BrowsePage(context.Background(), "")
	assert.Error(t, err)
}
