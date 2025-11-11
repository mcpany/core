/*
 * Copyright 2025 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"context"
	"net/http"
	"net/http/httptest"
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
			w.Write([]byte(`{"services": [{"name": "service1", "version": "2.0"}]}`))
		case "/collection2":
			w.Write([]byte(`{"services": [{"name": "service3", "version": "1.0"}]}`))
		case "/collection3":
			w.Write([]byte(`{"services": [{"name": "service1", "version": "3.0"}]}`))
		case "/collection-invalid-semver":
			w.Write([]byte(`{"version": "invalid", "services": [{"name": "service1", "version": "1.0"}]}`))
		case "/collection-yaml":
			w.Header().Set("Content-Type", "application/x-yaml")
			w.Write([]byte(`
services:
- name: service1
  version: "2.0"
`))
		case "/collection-authed":
			if r.Header.Get("Authorization") != "Bearer my-secret-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Write([]byte(`{"services": [{"name": "service1", "version": "4.0"}]}`))
		case "/collection-apikey":
			if r.Header.Get("X-API-Key") != "my-api-key" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Write([]byte(`{"services": [{"name": "service1", "version": "5.0"}]}`))
		case "/collection-basicauth":
			user, pass, ok := r.BasicAuth()
			if !ok || user != "testuser" || pass != "testpass" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Write([]byte(`{"services": [{"name": "service1", "version": "6.0"}]}`))
		case "/collection-no-content-type":
			w.Header().Del("Content-Type")
			w.Write([]byte(`
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
			manager := NewUpstreamServiceManager()
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
