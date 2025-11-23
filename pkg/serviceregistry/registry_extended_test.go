// Copyright 2025 Author(s) of MCP Any
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-jose/go-jose/v4"
	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/bus"
	"github.com/mcpany/core/pkg/pool"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/pkg/upstream/factory"
	busv1 "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistry_UnregisterService(t *testing.T) {
	pm := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(pm)
	bp, err := bus.NewBusProvider(&busv1.MessageBus{})
	require.NoError(t, err)
	tm := tool.NewToolManager(bp)
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost")
	serviceConfig.SetHttpService(httpService)

	// Register a service first
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Successful unregistration
	err = registry.UnregisterService(context.Background(), serviceID)
	assert.NoError(t, err)

	// Verify the service is gone
	_, ok := registry.GetServiceConfig(serviceID)
	assert.False(t, ok)

	// Unregister non-existent service
	err = registry.UnregisterService(context.Background(), "non-existent")
	assert.Error(t, err)
}

func TestServiceRegistry_GetAllServices(t *testing.T) {
	i := 0
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func() (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					i++
					return fmt.Sprintf("mock-service-key-%d", i), nil, nil, nil
				},
			}, nil
		},
	}
	bp, err := bus.NewBusProvider(&busv1.MessageBus{})
	require.NoError(t, err)
	tm := tool.NewToolManager(bp)
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	serviceConfig1.SetName("test-service-1")
	httpService1 := &configv1.HttpUpstreamService{}
	httpService1.SetAddress("http://localhost")
	serviceConfig1.SetHttpService(httpService1)

	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("test-service-2")
	httpService2 := &configv1.HttpUpstreamService{}
	httpService2.SetAddress("http://localhost")
	serviceConfig2.SetHttpService(httpService2)

	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err)
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.NoError(t, err)

	services, err := registry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, services, 2)
}

func TestServiceRegistry_RegisterService_OAuth2(t *testing.T) {
	// Create a mock OIDC provider
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwk := jose.JSONWebKey{
		Key:       &privateKey.PublicKey,
		Algorithm: string(jose.RS256),
		Use:       "sig",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(struct {
			Issuer  string `json:"issuer"`
			JWKSURI string `json:"jwks_uri"`
		}{
			Issuer:  "http://" + r.Host,
			JWKSURI: "http://" + r.Host + "/jwks",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jose.JSONWebKeySet{
			Keys: []jose.JSONWebKey{jwk},
		})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	pm := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(pm)
	bp, err := bus.NewBusProvider(&busv1.MessageBus{})
	require.NoError(t, err)
	tm := tool.NewToolManager(bp)
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost")
	serviceConfig.SetHttpService(httpService)

	authConfig := &configv1.AuthenticationConfig{}
	oauth2Config := &configv1.OAuth2Auth{}
	oauth2Config.SetIssuerUrl(server.URL) // Use the mock server's URL
	oauth2Config.SetAudience("test-audience")
	authConfig.SetOauth2(oauth2Config)
	serviceConfig.SetAuthentication(authConfig)

	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	// Check authenticator
	_, ok := am.GetAuthenticator(serviceID)
	assert.True(t, ok, "Authenticator should have been added")
}

func TestServiceRegistry_RegisterService_OAuth2_Error(t *testing.T) {
	pm := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(pm)
	bp, err := bus.NewBusProvider(&busv1.MessageBus{})
	require.NoError(t, err)
	tm := tool.NewToolManager(bp)
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-service")
	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost")
	serviceConfig.SetHttpService(httpService)

	authConfig := &configv1.AuthenticationConfig{}
	oauth2Config := &configv1.OAuth2Auth{}
	oauth2Config.SetIssuerUrl("http://invalid-url") // Invalid URL
	authConfig.SetOauth2(oauth2Config)
	serviceConfig.SetAuthentication(authConfig)

	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
}

func TestServiceRegistry_RegisterDuplicateService_RestoresOriginalTools(t *testing.T) {
	pm := pool.NewManager()
	f := factory.NewUpstreamServiceFactory(pm)
	bp, err := bus.NewBusProvider(&busv1.MessageBus{})
	require.NoError(t, err)
	tm := tool.NewToolManager(bp)
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	// Define the original service with one tool
	originalServiceConfig := &configv1.UpstreamServiceConfig{}
	originalServiceConfig.SetName("test-service")
	originalHttpService := &configv1.HttpUpstreamService{}
	originalHttpService.SetAddress("http://localhost")
	originalHttpService.SetCalls(map[string]*configv1.HttpCallDefinition{
		"original_tool": {
			Id:     func() *string { s := "original_tool"; return &s }(),
			Method: configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
		},
	})
	originalHttpService.SetTools([]*configv1.ToolDefinition{
		{
			Name:   func() *string { s := "original_tool"; return &s }(),
			CallId: func() *string { s := "original_tool"; return &s }(),
		},
	})
	originalServiceConfig.SetHttpService(originalHttpService)

	// Register the original service
	serviceID, _, _, err := registry.RegisterService(context.Background(), originalServiceConfig)
	require.NoError(t, err)
	assert.Equal(t, "test-service", serviceID)

	// Verify the original tool is registered
	originalTools := tm.ListTools()
	require.Len(t, originalTools, 1)
	assert.Equal(t, "original_tool", originalTools[0].Tool().GetName())

	// Define a duplicate service with the same name but a different tool
	duplicateServiceConfig := &configv1.UpstreamServiceConfig{}
	duplicateServiceConfig.SetName("test-service")
	duplicateHttpService := &configv1.HttpUpstreamService{}
	duplicateHttpService.SetAddress("http://anotherhost")
	duplicateHttpService.SetCalls(map[string]*configv1.HttpCallDefinition{
		"duplicate_tool": {
			Id:     func() *string { s := "duplicate_tool"; return &s }(),
			Method: configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
		},
	})
	duplicateHttpService.SetTools([]*configv1.ToolDefinition{
		{
			Name:   func() *string { s := "duplicate_tool"; return &s }(),
			CallId: func() *string { s := "duplicate_tool"; return &s }(),
		},
	})
	duplicateServiceConfig.SetHttpService(duplicateHttpService)

	// Attempt to register the duplicate service
	_, _, _, err = registry.RegisterService(context.Background(), duplicateServiceConfig)

	// Verify that the registration failed as expected
	require.Error(t, err)
	assert.Contains(t, err.Error(), `service with name "test-service" already registered`)

	// Verify the original tool is still present and the duplicate tool was not added
	finalTools := tm.ListTools()
	require.Len(t, finalTools, 1, "The number of tools for the original service should not have changed")
	assert.Equal(t, "original_tool", finalTools[0].Tool().GetName(), "The original tool should still be registered")
}
