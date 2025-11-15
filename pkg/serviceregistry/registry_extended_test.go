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
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceRegistry_UnregisterService(t *testing.T) {
	f := &mockFactory{}
	tm := &mockToolManager{}
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
	tm := &mockToolManager{}
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

	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig1)
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

	f := &mockFactory{}
	tm := &mockToolManager{}
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
	f := &mockFactory{}
	tm := &mockToolManager{}
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

	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.Error(t, err)
}

func TestServiceRegistry_ServiceInfo(t *testing.T) {
	f := &mockFactory{}
	tm := &mockToolManager{}
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

	// Test GetServiceInfo for a service that doesn't have info yet
	_, ok := registry.GetServiceInfo(serviceID)
	assert.False(t, ok, "should return false for a service with no info")

	// Add service info
	serviceInfo := &tool.ServiceInfo{
		Name:   "test-service",
		Config: serviceConfig,
	}

	registry.AddServiceInfo(serviceID, serviceInfo)

	// Get the service info and verify it
	retrievedInfo, ok := registry.GetServiceInfo(serviceID)
	require.True(t, ok, "should return true after adding service info")
	assert.Equal(t, serviceInfo.Name, retrievedInfo.Name)
	assert.Equal(t, serviceInfo.Config, retrievedInfo.Config)

	// Test GetServiceInfo for a non-existent service
	_, ok = registry.GetServiceInfo("non-existent-id")
	assert.False(t, ok, "should return false for a non-existent service")
}
