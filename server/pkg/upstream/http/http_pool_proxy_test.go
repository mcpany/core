// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package http //nolint:revive,nolintlint // Package name 'http' is intentional for this directory structure.

import (
	"net/http"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHTTPPool_ProxyConfiguration(t *testing.T) {
	t.Run("with proxy", func(t *testing.T) {
		proxyURL := "http://proxy.example.com:8080"
		config := configv1.UpstreamServiceConfig_builder{
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("https://api.example.com"),
				ProxyConfig: configv1.ProxyConfig_builder{
					Url: proto.String(proxyURL),
				}.Build(),
			}.Build(),
		}.Build()

		p, err := NewHTTPPool(1, 1, 10, config)
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		hp, ok := p.(*httpPool)
		require.True(t, ok, "Should be httpPool")

		req, _ := http.NewRequest("GET", "https://api.example.com", nil)
		proxy, err := hp.transport.Proxy(req)
		require.NoError(t, err)
		require.NotNil(t, proxy)
		assert.Equal(t, proxyURL, proxy.String())
	})

	t.Run("with authenticated proxy", func(t *testing.T) {
		proxyURL := "http://proxy.example.com:8080"
		username := "user"
		password := "pass"

		config := configv1.UpstreamServiceConfig_builder{
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("https://api.example.com"),
				ProxyConfig: configv1.ProxyConfig_builder{
					Url: proto.String(proxyURL),
					Username: proto.String(username),
					Password: proto.String(password),
				}.Build(),
			}.Build(),
		}.Build()

		p, err := NewHTTPPool(1, 1, 10, config)
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		hp, ok := p.(*httpPool)
		require.True(t, ok)

		req, _ := http.NewRequest("GET", "https://api.example.com", nil)
		proxy, err := hp.transport.Proxy(req)
		require.NoError(t, err)
		require.NotNil(t, proxy)

		assert.Equal(t, "http://user:pass@proxy.example.com:8080", proxy.String())
	})

	t.Run("without proxy", func(t *testing.T) {
		config := configv1.UpstreamServiceConfig_builder{
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("https://api.example.com"),
			}.Build(),
		}.Build()

		p, err := NewHTTPPool(1, 1, 10, config)
		require.NoError(t, err)
		defer func() { _ = p.Close() }()

		hp, ok := p.(*httpPool)
		require.True(t, ok)

		// Should default to ProxyFromEnvironment (which is a function)
		assert.NotNil(t, hp.transport.Proxy)
	})

    t.Run("invalid proxy url", func(t *testing.T) {
		proxyURL := "::invalid::"
		config := configv1.UpstreamServiceConfig_builder{
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("https://api.example.com"),
				ProxyConfig: configv1.ProxyConfig_builder{
					Url: proto.String(proxyURL),
				}.Build(),
			}.Build(),
		}.Build()

		_, err := NewHTTPPool(1, 1, 10, config)
		require.Error(t, err)
        assert.Contains(t, err.Error(), "invalid proxy URL")
    })
}
