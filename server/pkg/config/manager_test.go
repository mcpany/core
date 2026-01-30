// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpstreamServiceManager_LoadAndMergeServices(t *testing.T) {
	// Local service definitions
	localService1 := &configv1.UpstreamServiceConfig{}
	localService1.SetName("service1")
	localService1.SetVersion("1.0")
	localService2 := &configv1.UpstreamServiceConfig{}
	localService2.SetName("service2")
	localService2.SetVersion("1.0")

	// Create a mock HTTP server to serve the remote collections
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/collection1":
			_, _ = w.Write([]byte(`{"services": [{"name": "service1", "version": "2.0"}]}`))
		case "/collection2":
			_, _ = w.Write([]byte(`{"services": [{"name": "service3", "version": "1.0"}]}`))
		case "/collection3":
			_, _ = w.Write([]byte(`{"services": [{"name": "service1", "version": "3.0"}]}`))
		case "/collection-invalid-semver":
			_, _ = w.Write([]byte(`{"version": "invalid", "services": [{"name": "service1", "version": "1.0"}]}`))
		case "/collection-yaml":
			w.Header().Set("Content-Type", "application/x-yaml")
			_, _ = w.Write([]byte(`
services:
- name: service1
  version: "2.0"
`))
		case "/collection-authed":
			if r.Header.Get("Authorization") != "Bearer my-secret-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"services": [{"name": "service1", "version": "4.0"}]}`))
		case "/collection-apikey":
			if r.Header.Get("X-API-Key") != "my-api-key" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"services": [{"name": "service1", "version": "5.0"}]}`))
		case "/collection-basicauth":
			user, pass, ok := r.BasicAuth()
			if !ok || user != "testuser" || pass != "testpass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			_, _ = w.Write([]byte(`{"services": [{"name": "service1", "version": "6.0"}]}`))
		case "/collection-no-content-type":
			w.Header().Del("Content-Type")
			_, _ = w.Write([]byte(`
services:
- name: service1
  version: "7.0"
`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	testCases := []struct {
		name                            string
		initialConfig                   *configv1.McpAnyServerConfig
		expectedServiceNamesAndVersions map[string]string
		expectLoadError                 bool
	}{
		{
			name: "local services only",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{localService1, localService2},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "1.0",
				"service2": "1.0",
			},
		},
		{
			name: "local and remote services, remote has higher priority",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{localService1, localService2})

				col := &configv1.Collection{}
				col.SetName("collection1")
				col.SetHttpUrl(server.URL + "/collection1")
				col.SetPriority(-1)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "2.0",
				"service2": "1.0",
			},
		},
		{
			name: "local and remote services, local has higher priority",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{localService1, localService2})

				col := &configv1.Collection{}
				col.SetName("collection1")
				col.SetHttpUrl(server.URL + "/collection1")
				col.SetPriority(1)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "1.0",
				"service2": "1.0",
			},
		},
		{
			name: "multiple remote collections, mixed priorities",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{localService1, localService2})

				col1 := &configv1.Collection{}
				col1.SetName("collection1")
				col1.SetHttpUrl(server.URL + "/collection1")
				col1.SetPriority(1)

				col2 := &configv1.Collection{}
				col2.SetName("collection2")
				col2.SetHttpUrl(server.URL + "/collection2")
				col2.SetPriority(-1)

				col3 := &configv1.Collection{}
				col3.SetName("collection3")
				col3.SetHttpUrl(server.URL + "/collection3")
				col3.SetPriority(-2)

				cfg.SetCollections([]*configv1.Collection{col1, col2, col3})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "3.0",
				"service2": "1.0",
				"service3": "1.0",
			},
		},
		{
			name: "same priority, first one wins",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{localService1})

				col := &configv1.Collection{}
				col.SetName("collection1")
				col.SetHttpUrl(server.URL + "/collection1")
				col.SetPriority(0)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "1.0",
			},
		},
		{
			name: "invalid semver",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}

				col := &configv1.Collection{}
				col.SetName("collection-invalid-semver")
				col.SetHttpUrl(server.URL + "/collection-invalid-semver")
				col.SetPriority(0)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{},
			expectLoadError:                 false, // The manager logs a warning but doesn't return an error
		},
		{
			name: "yaml content type",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}

				col := &configv1.Collection{}
				col.SetName("collection-yaml")
				col.SetHttpUrl(server.URL + "/collection-yaml")
				col.SetPriority(0)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "2.0",
			},
		},
		{
			name: "authenticated collection",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}

				col := &configv1.Collection{}
				col.SetName("collection-authed")
				col.SetHttpUrl(server.URL + "/collection-authed")

				auth := &configv1.Authentication{}
				bearer := &configv1.BearerTokenAuth{}
				val := &configv1.SecretValue{}
				val.SetPlainText("my-secret-token")
				bearer.SetToken(val)
				auth.SetBearerToken(bearer)
				col.SetAuthentication(auth)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "4.0",
			},
		},
		{
			name: "api key authenticated collection",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}

				col := &configv1.Collection{}
				col.SetName("collection-apikey")
				col.SetHttpUrl(server.URL + "/collection-apikey")

				auth := &configv1.Authentication{}
				apiKey := &configv1.APIKeyAuth{}
				apiKey.SetParamName("X-API-Key")
				val := &configv1.SecretValue{}
				val.SetPlainText("my-api-key")
				apiKey.SetValue(val)
				auth.SetApiKey(apiKey)
				col.SetAuthentication(auth)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "5.0",
			},
		},
		{
			name: "basic auth authenticated collection",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}

				col := &configv1.Collection{}
				col.SetName("collection-basicauth")
				col.SetHttpUrl(server.URL + "/collection-basicauth")

				auth := &configv1.Authentication{}
				basic := &configv1.BasicAuth{}
				basic.SetUsername("testuser")
				val := &configv1.SecretValue{}
				val.SetPlainText("testpass")
				basic.SetPassword(val)
				auth.SetBasicAuth(basic)
				col.SetAuthentication(auth)

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "6.0",
			},
		},
		{
			name: "no content type assumes yaml",
			initialConfig: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}

				col := &configv1.Collection{}
				col.SetName("collection-no-content-type")
				col.SetHttpUrl(server.URL + "/collection-no-content-type")

				cfg.SetCollections([]*configv1.Collection{col})
				return cfg
			}(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "7.0",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manager := NewUpstreamServiceManager(nil)
			manager.httpClient = &http.Client{}
			loadedServices, err := manager.LoadAndMergeServices(context.Background(), tc.initialConfig)

			if tc.expectLoadError {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, len(tc.expectedServiceNamesAndVersions), len(loadedServices))

			serviceMap := make(map[string]*configv1.UpstreamServiceConfig)
			for _, s := range loadedServices {
				serviceMap[s.GetName()] = s
			}

			for name, version := range tc.expectedServiceNamesAndVersions {
				s, ok := serviceMap[name]
				assert.True(t, ok, "expected service %s to be loaded", name)
				assert.Equal(t, version, s.GetVersion())
			}
		})
	}
}

func TestUpstreamServiceManager_Profiles_Overrides(t *testing.T) {
	// Config with 2 services and profile definitions
	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}

		gs := &configv1.GlobalSettings{}

		profDev := &configv1.ProfileDefinition{}
		profDev.SetName("dev")

		svcOverlapDev := &configv1.ProfileServiceConfig{}
		svcOverlapDev.SetEnabled(true)

		svcOverlapProd := &configv1.ProfileServiceConfig{}
		svcOverlapProd.SetEnabled(false)

		profDev.SetServiceConfig(map[string]*configv1.ProfileServiceConfig{
			"dev-service": svcOverlapDev,
			"prod-service": svcOverlapProd,
		})

		profProd := &configv1.ProfileDefinition{}
		profProd.SetName("prod")

		svcOverlapDev2 := &configv1.ProfileServiceConfig{}
		svcOverlapDev2.SetEnabled(false)

		svcOverlapProd2 := &configv1.ProfileServiceConfig{}
		svcOverlapProd2.SetEnabled(true)

		profProd.SetServiceConfig(map[string]*configv1.ProfileServiceConfig{
			"dev-service": svcOverlapDev2,
			"prod-service": svcOverlapProd2,
		})

		gs.SetProfileDefinitions([]*configv1.ProfileDefinition{profDev, profProd})
		cfg.SetGlobalSettings(gs)

		svc1 := &configv1.UpstreamServiceConfig{}
		svc1.SetName("dev-service")
		http1 := &configv1.HttpUpstreamService{}
		http1.SetAddress("http://dev")
		svc1.SetHttpService(http1)

		svc2 := &configv1.UpstreamServiceConfig{}
		svc2.SetName("prod-service")
		http2 := &configv1.HttpUpstreamService{}
		http2.SetAddress("http://prod")
		svc2.SetHttpService(http2)

		svc3 := &configv1.UpstreamServiceConfig{}
		svc3.SetName("common-service")
		http3 := &configv1.HttpUpstreamService{}
		http3.SetAddress("http://common")
		svc3.SetHttpService(http3)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc1, svc2, svc3})
		return cfg
	}()

	tests := []struct {
		name            string
		enabledProfiles []string
		expectedNames   []string
	}{
		{
			name:            "No Profile (common only?)",
			enabledProfiles: []string{}, // Default profile?
			// If no profile overrides, services are enabled by default (logic in manager.go:314 allowed := !isOverrideDisabled)
			expectedNames:   []string{"common-service", "dev-service", "prod-service"},
		},
		{
			name:            "Dev Profile",
			enabledProfiles: []string{"dev"},
			expectedNames:   []string{"common-service", "dev-service"}, // prod-service disabled by dev profile
		},
		{
			name:            "Prod Profile",
			enabledProfiles: []string{"prod"},
			expectedNames:   []string{"common-service", "prod-service"}, // dev-service disabled by prod profile
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewUpstreamServiceManager(tt.enabledProfiles)
			services, err := manager.LoadAndMergeServices(context.Background(), config)
			assert.NoError(t, err)

			var names []string
			for _, s := range services {
				names = append(names, s.GetName())
			}
			sort.Strings(names)
			sort.Strings(tt.expectedNames)
			assert.Equal(t, tt.expectedNames, names)
		})
	}
}
