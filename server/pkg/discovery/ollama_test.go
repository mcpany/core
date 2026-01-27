// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOllamaProvider_Name(t *testing.T) {
	provider := &OllamaProvider{}
	assert.Equal(t, "ollama", provider.Name())
}

func TestOllamaProvider_Discover(t *testing.T) {
	tests := []struct {
		name        string
		setupServer func(t *testing.T) *httptest.Server
		wantErr     bool
		wantSvcs    int
	}{
		{
			name: "ollama discovered successfully",
			setupServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/api/tags", r.URL.Path)
					assert.Equal(t, "GET", r.Method)
					w.WriteHeader(http.StatusOK)
					// We can return valid JSON if the implementation parses it, but currently it just checks for 200 OK.
					// Providing empty JSON object just in case logic changes to parse it later, though currently unused.
					_, _ = w.Write([]byte(`{}`))
				}))
			},
			wantErr:  false,
			wantSvcs: 1,
		},
		{
			name: "ollama unreachable",
			setupServer: func(t *testing.T) *httptest.Server {
				// Create a server and close it immediately to simulate unreachable
				s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				s.Close()
				return s
			},
			wantErr:  true,
			wantSvcs: 0,
		},
		{
			name: "ollama returns error status",
			setupServer: func(t *testing.T) *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			wantErr:  true,
			wantSvcs: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.setupServer(t)

			// We only want to defer Close if the server is not already closed by the setup function (for unreachable test)
			// A cleaner way is to let the setup function handle lifecycle, but we need the URL.
			// The unreachable test closes it to make it unreachable.
			// httptest.Server.Close() panics if called twice?
			// Documentation says "Close shuts down the server...". It doesn't explicitly say it panics on double close,
			// but it's better to be safe.
			// However, checking if it's closed is not straightforward on httptest.Server struct.
			// We can use a flag in the test case struct or just let the "unreachable" case be special.

			if tt.name != "ollama unreachable" {
				defer server.Close()
			}

			provider := &OllamaProvider{
				Endpoint: server.URL,
			}

			svcs, err := provider.Discover(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Len(t, svcs, tt.wantSvcs)
				if tt.wantSvcs > 0 {
					svc := svcs[0]
					assert.Equal(t, "Local Ollama", svc.GetName())
					assert.Equal(t, "v1", svc.GetVersion())
					assert.Equal(t, server.URL+"/v1", svc.GetHttpService().GetAddress())
					assert.Contains(t, svc.GetTags(), "local-llm")
				}
			}
		})
	}
}
