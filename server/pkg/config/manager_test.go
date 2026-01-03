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
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamServiceManager_LoadAndMergeServices(t *testing.T) {
	// Local service definitions
	localService1 := configv1.UpstreamServiceConfig_builder{}.Build()
	localService1.SetName("service1")
	localService1.SetVersion("1.0")
	localService2 := configv1.UpstreamServiceConfig_builder{}.Build()
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

func TestLoadAndMergeServices_Profiles(t *testing.T) {
	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name:     proto.String("default-service"),
				Profiles: []*configv1.Profile{}, // Should default to "default"
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name:     proto.String("dev-service"),
				Profiles: []*configv1.Profile{configv1.Profile_builder{Name: "dev"}.Build()},
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name:     proto.String("prod-service"),
				Profiles: []*configv1.Profile{configv1.Profile_builder{Name: "prod"}.Build()},
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name:     proto.String("mixed-service"),
				Profiles: []*configv1.Profile{
					configv1.Profile_builder{Name: "dev"}.Build(),
					configv1.Profile_builder{Name: "prod"}.Build(),
				},
			}.Build(),
		},
	}.Build()

	tests := []struct {
		name            string
		enabledProfiles []string
		expectedNames   []string
	}{
		{
			name:            "Default Profile",
			enabledProfiles: nil, // Defaults to "default"
			expectedNames:   []string{"default-service"},
		},
		{
			name:            "Dev Profile",
			enabledProfiles: []string{"dev"},
			expectedNames:   []string{"dev-service", "mixed-service"},
		},
		{
			name:            "Prod Profile",
			enabledProfiles: []string{"prod"},
			expectedNames:   []string{"mixed-service", "prod-service"},
		},
		{
			name:            "Multiple Profiles",
			enabledProfiles: []string{"dev", "default"},
			expectedNames:   []string{"default-service", "dev-service", "mixed-service"},
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

func TestUnmarshalServices(t *testing.T) {
	manager := NewUpstreamServiceManager(nil)

	t.Run("unmarshal single service from JSON", func(t *testing.T) {
		jsonData := `{"name": "single-service", "version": "1.0"}`
		var services []*configv1.UpstreamServiceConfig
		err := manager.unmarshalServices([]byte(jsonData), &services, "application/json")
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "single-service", services[0].GetName())
	})

	t.Run("unmarshal service list from protobuf", func(t *testing.T) {
		service := configv1.UpstreamServiceConfig_builder{}.Build()
		service.SetName("proto-service")
		service.SetVersion("1.0")

		serviceList := configv1.UpstreamServiceCollectionShare_builder{}.Build()
		serviceList.SetServices([]*configv1.UpstreamServiceConfig{service})
		protoData, err := prototext.Marshal(serviceList)
		require.NoError(t, err)

		var services []*configv1.UpstreamServiceConfig
		err = manager.unmarshalServices(protoData, &services, "application/protobuf")
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "proto-service", services[0].GetName())
	})

	t.Run("unmarshal service list with text/plain content type", func(t *testing.T) {
		service := configv1.UpstreamServiceConfig_builder{}.Build()
		service.SetName("text-service")
		service.SetVersion("1.0")

		serviceList := configv1.UpstreamServiceCollectionShare_builder{}.Build()
		serviceList.SetServices([]*configv1.UpstreamServiceConfig{service})

		protoData, err := prototext.Marshal(serviceList)
		require.NoError(t, err)

		var services []*configv1.UpstreamServiceConfig
		err = manager.unmarshalServices(protoData, &services, "text/plain")
		require.NoError(t, err)
		assert.Len(t, services, 1)
		assert.Equal(t, "text-service", services[0].GetName())
	})

	t.Run("unmarshal invalid protobuf", func(t *testing.T) {
		var services []*configv1.UpstreamServiceConfig
		err := manager.unmarshalServices([]byte("invalid-proto"), &services, "application/protobuf")
		assert.Error(t, err)
	})

	t.Run("unmarshal invalid yaml", func(t *testing.T) {
		var services []*configv1.UpstreamServiceConfig
		err := manager.unmarshalServices([]byte("services: - name: foo\n- bar"), &services, "application/x-yaml")
		assert.Error(t, err)
	})

	t.Run("unmarshal invalid json", func(t *testing.T) {
		var services []*configv1.UpstreamServiceConfig
		err := manager.unmarshalServices([]byte(`{"services": "not-a-list"}`), &services, "application/json")
		assert.Error(t, err)
	})

	t.Run("unmarshal empty json", func(t *testing.T) {
		var services []*configv1.UpstreamServiceConfig
		err := manager.unmarshalServices([]byte(`{}`), &services, "application/json")
		require.NoError(t, err)
		assert.Len(t, services, 0)
	})
}

func TestUpstreamServiceManager_ProfilesBehavior(t *testing.T) {
	// Create a config with 3 services:
	// 1. dev-service (profile: dev)
	// 2. prod-service (profile: prod)
	// 3. common-service (profile: dev, prod)
	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name:     proto.String("dev-service"),
				Profiles: []*configv1.Profile{configv1.Profile_builder{Name: "dev"}.Build()},
				HttpService: configv1.HttpUpstreamService_builder{Address: proto.String("http://dev")}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name:     proto.String("prod-service"),
				Profiles: []*configv1.Profile{configv1.Profile_builder{Name: "prod"}.Build()},
				HttpService: configv1.HttpUpstreamService_builder{Address: proto.String("http://prod")}.Build(),
			}.Build(),
			configv1.UpstreamServiceConfig_builder{
				Name:     proto.String("common-service"),
				Profiles: []*configv1.Profile{
					configv1.Profile_builder{Name: "dev"}.Build(),
					configv1.Profile_builder{Name: "prod"}.Build(),
				},
				HttpService: configv1.HttpUpstreamService_builder{Address: proto.String("http://common")}.Build(),
			}.Build(),
		},
	}.Build()

	// Case 1: Run with 'dev' profile
	mgrDev := NewUpstreamServiceManager([]string{"dev"})
	servicesDev, err := mgrDev.LoadAndMergeServices(context.Background(), cfg)
	require.NoError(t, err)

	require.Len(t, servicesDev, 2)
	require.Equal(t, "common-service", servicesDev[0].GetName()) // sorted
	require.Equal(t, "dev-service", servicesDev[1].GetName())

	// Case 2: Run with 'prod' profile
	mgrProd := NewUpstreamServiceManager([]string{"prod"})
	servicesProd, err := mgrProd.LoadAndMergeServices(context.Background(), cfg)
	require.NoError(t, err)

	require.Len(t, servicesProd, 2)
	require.Equal(t, "common-service", servicesProd[0].GetName())
	require.Equal(t, "prod-service", servicesProd[1].GetName())
}

func TestUpstreamServiceManager_LoadFromURLErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	mgr := NewUpstreamServiceManager(nil)
	mgr.httpClient = server.Client() // Use server client

	t.Run("404 error", func(t *testing.T) {
		config := configv1.McpAnyServerConfig_builder{
			UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
				configv1.UpstreamServiceCollection_builder{
					Name:    proto.String("404-collection"),
					HttpUrl: proto.String(server.URL + "/404"),
				}.Build(),
			},
		}.Build()
		// Errors are logged and suppressed for individual collections
		services, err := mgr.LoadAndMergeServices(context.Background(), config)
		assert.NoError(t, err)
		assert.Len(t, services, 0)
	})

	t.Run("connection error", func(t *testing.T) {
		// Use a closed port for connection refused
		config := configv1.McpAnyServerConfig_builder{
			UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
				configv1.UpstreamServiceCollection_builder{
					Name:    proto.String("connection-error"),
					HttpUrl: proto.String("http://127.0.0.1:0/invalid"),
				}.Build(),
			},
		}.Build()
		// Errors are logged and suppressed
		services, err := mgr.LoadAndMergeServices(context.Background(), config)
		assert.NoError(t, err)
		assert.Len(t, services, 0)
	})

	t.Run("auth error", func(t *testing.T) {
		config := configv1.McpAnyServerConfig_builder{
			UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
				configv1.UpstreamServiceCollection_builder{
					Name:    proto.String("auth-error"),
					HttpUrl: proto.String(server.URL + "/auth-error"),
					Authentication: configv1.UpstreamAuthentication_builder{
						ApiKey: configv1.UpstreamAPIKeyAuth_builder{
							HeaderName: proto.String("X-API-Key"),
							ApiKey: configv1.SecretValue_builder{
								PlainText: strPtr("file:nonexistent-secret"),
							}.Build(),
						}.Build(),
					}.Build(),
				}.Build(),
			},
		}.Build()
		// Should fail to apply auth -> Log warning -> 0 services
		services, err := mgr.LoadAndMergeServices(context.Background(), config)
		assert.NoError(t, err)
		assert.Len(t, services, 0)
	})

	t.Run("invalid url chars", func(t *testing.T) {
		config := configv1.McpAnyServerConfig_builder{
			UpstreamServiceCollections: []*configv1.UpstreamServiceCollection{
				configv1.UpstreamServiceCollection_builder{
					Name:    proto.String("invalid-url"),
					HttpUrl: proto.String("http://invalid\x7furl"),
				}.Build(),
			},
		}.Build()
		// Should fail to create request -> Log warning -> 0 services
		services, err := mgr.LoadAndMergeServices(context.Background(), config)
		assert.NoError(t, err)
		assert.Len(t, services, 0)
	})
}

func TestUpstreamServiceManager_ProfileMatching_ByID(t *testing.T) {
	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{
				Name: proto.String("id-only-service"),
				Profiles: []*configv1.Profile{
					configv1.Profile_builder{
						Id:   "prod-id",
						Name: "Production",
					}.Build(),
				},
				HttpService: configv1.HttpUpstreamService_builder{Address: proto.String("http://prod")}.Build(),
			}.Build(),
		},
	}.Build()

	// Case: Run with 'prod-id' profile.
	// Expected: The service should be loaded because its profile ID matches "prod-id".
	mgr := NewUpstreamServiceManager([]string{"prod-id"})
	services, err := mgr.LoadAndMergeServices(context.Background(), config)
	require.NoError(t, err)

	assert.Len(t, services, 1, "Service should be loaded when matched by profile ID")
	if len(services) > 0 {
		assert.Equal(t, "id-only-service", services[0].GetName())
	}
}
