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
	"google.golang.org/protobuf/proto"
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
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{localService1, localService2},
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection1"),
						HttpUrl:  proto.String(server.URL + "/collection1"),
						Priority: proto.Int32(-1),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "2.0",
				"service2": "1.0",
			},
		},
		{
			name: "local and remote services, local has higher priority",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{localService1, localService2},
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection1"),
						HttpUrl:  proto.String(server.URL + "/collection1"),
						Priority: proto.Int32(1),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "1.0",
				"service2": "1.0",
			},
		},
		{
			name: "multiple remote collections, mixed priorities",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{localService1, localService2},
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection1"),
						HttpUrl:  proto.String(server.URL + "/collection1"),
						Priority: proto.Int32(1),
					}).Build(),
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection2"),
						HttpUrl:  proto.String(server.URL + "/collection2"),
						Priority: proto.Int32(-1),
					}).Build(),
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection3"),
						HttpUrl:  proto.String(server.URL + "/collection3"),
						Priority: proto.Int32(-2),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "3.0",
				"service2": "1.0",
				"service3": "1.0",
			},
		},
		{
			name: "same priority, first one wins",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServices: []*configv1.UpstreamServiceConfig{localService1},
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection1"),
						HttpUrl:  proto.String(server.URL + "/collection1"),
						Priority: proto.Int32(0),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "1.0",
			},
		},
		{
			name: "invalid semver",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection-invalid-semver"),
						HttpUrl:  proto.String(server.URL + "/collection-invalid-semver"),
						Priority: proto.Int32(0),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{},
			expectLoadError:                 false, // The manager logs a warning but doesn't return an error
		},
		{
			name: "yaml content type",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:     proto.String("collection-yaml"),
						HttpUrl:  proto.String(server.URL + "/collection-yaml"),
						Priority: proto.Int32(0),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "2.0",
			},
		},
		{
			name: "authenticated collection",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:    proto.String("collection-authed"),
						HttpUrl: proto.String(server.URL + "/collection-authed"),
						Authentication: (&configv1.UpstreamAuthentication_builder{
							BearerToken: (&configv1.UpstreamBearerTokenAuth_builder{
								Token: (&configv1.SecretValue_builder{PlainText: proto.String("my-secret-token")}).Build(),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "4.0",
			},
		},
		{
			name: "api key authenticated collection",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:    proto.String("collection-apikey"),
						HttpUrl: proto.String(server.URL + "/collection-apikey"),
						Authentication: (&configv1.UpstreamAuthentication_builder{
							ApiKey: (&configv1.UpstreamAPIKeyAuth_builder{
								HeaderName: proto.String("X-API-Key"),
								ApiKey:     (&configv1.SecretValue_builder{PlainText: proto.String("my-api-key")}).Build(),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "5.0",
			},
		},
		{
			name: "basic auth authenticated collection",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:    proto.String("collection-basicauth"),
						HttpUrl: proto.String(server.URL + "/collection-basicauth"),
						Authentication: (&configv1.UpstreamAuthentication_builder{
							BasicAuth: (&configv1.UpstreamBasicAuth_builder{
								Username: proto.String("testuser"),
								Password: (&configv1.SecretValue_builder{PlainText: proto.String("testpass")}).Build(),
							}).Build(),
						}).Build(),
					}).Build(),
				},
			}).Build(),
			expectedServiceNamesAndVersions: map[string]string{
				"service1": "6.0",
			},
		},
		{
			name: "no content type assumes yaml",
			initialConfig: (configv1.McpAnyServerConfig_builder{
				UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
					(configv1.UpstreamServiceCollection_builder{
						Name:    proto.String("collection-no-content-type"),
						HttpUrl: proto.String(server.URL + "/collection-no-content-type"),
					}).Build(),
				},
			}).Build(),
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
	config := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			ProfileDefinitions: []*configv1.ProfileDefinition{
				{
					Name: proto.String("dev"),
					ServiceConfig: map[string]*configv1.ProfileServiceConfig{
						"dev-service": {Enabled: proto.Bool(true)},
						"prod-service": {Enabled: proto.Bool(false)},
					},
				},
				{
					Name: proto.String("prod"),
					ServiceConfig: map[string]*configv1.ProfileServiceConfig{
						"dev-service": {Enabled: proto.Bool(false)},
						"prod-service": {Enabled: proto.Bool(true)},
					},
				},
			},
		},
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name:     proto.String("dev-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://dev")},
				},
			},
			{
				Name:     proto.String("prod-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://prod")},
				},
			},
			{
				Name:     proto.String("common-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://common")},
				},
			},
		},
	}

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
