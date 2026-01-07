// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowserProvider(t *testing.T) {
	p := NewProvider()

	def := p.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	content, err := p.BrowsePage(context.Background(), "https://example.com")
	assert.NoError(t, err)
	assert.Contains(t, content, "Mock Content")
	assert.Contains(t, content, "example.com")

	_, err = p.BrowsePage(context.Background(), "")
	assert.Error(t, err)
}
