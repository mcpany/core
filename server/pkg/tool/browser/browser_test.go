// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrowserProvider(t *testing.T) {
	p := NewProvider()
	defer p.Close()

	def := p.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	content, err := p.BrowsePage(context.Background(), "https://example.com")
	assert.NoError(t, err)
	assert.True(t, strings.Contains(content, "example") || strings.Contains(content, "Example"))

	_, err = p.BrowsePage(context.Background(), "")
	assert.Error(t, err)
}

func TestBrowserProviderSSRF(t *testing.T) {
	p := NewProvider()
	defer p.Close()

	_, err := p.BrowsePage(context.Background(), "http://localhost:8080")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security policy violation")

	_, err = p.BrowsePage(context.Background(), "http://127.0.0.1:8080")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security policy violation")

	_, err = p.BrowsePage(context.Background(), "http://192.168.1.1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security policy violation")

	_, err = p.BrowsePage(context.Background(), "file:///etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "security policy violation")
}
