// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package browser

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// httpPageFetcher is a lightweight PageFetcher for tests that fetches pages
// via plain HTTP instead of a real browser. It reads the raw response body and
// uses that as the "text content", which is sufficient for the test server
// that returns a minimal HTML document.
type httpPageFetcher struct{}

func (f *httpPageFetcher) FetchText(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("could not create request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not fetch page: %w", err)
	}
	defer resp.Body.Close()
	var buf []byte
	tmp := make([]byte, 4096)
	for {
		n, readErr := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if readErr != nil {
			break
		}
	}
	return string(buf), nil
}

func TestBrowserProvider(t *testing.T) {
	// Use a local HTTP test server to avoid external network access.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html><head><title>Test Page</title></head><body><h1>Hello from local test server</h1></body></html>")
	}))
	defer ts.Close()

	// Inject the HTTP-based fetcher so no real browser / playwright installation
	// is required. The production code uses playwrightFetcher by default; this
	// override keeps the test hermetic and fast.
	p := &Provider{fetcher: &httpPageFetcher{}}

	def := p.ToolDefinition()
	assert.Equal(t, "browse_page", def["name"])

	content, err := p.BrowsePage(context.Background(), ts.URL)
	require.NoError(t, err)
	assert.Contains(t, content, "Hello from local test server")

	_, err = p.BrowsePage(context.Background(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "url is required")
}
