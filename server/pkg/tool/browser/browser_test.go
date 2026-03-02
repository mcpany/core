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
	if os.Getenv("CI") != "" {
		t.Skip("Skipping browser test in CI environment without guaranteed playwright driver.")
	}

	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{"chromium"},
	})
	if err != nil {
		t.Skipf("skipping test, could not install playwright: %v", err)
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
