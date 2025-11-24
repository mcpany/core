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

package serviceregistry

import (
	"context"
	"testing"

	"fmt"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	"github.com/mcpany/core/pkg/tool"
	"github.com/mcpany/core/pkg/upstream"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
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

	// Register the service
	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)
	assert.Equal(t, "mock-service-key", serviceID)

	// Unregister the service
	err = registry.UnregisterService(context.Background(), serviceID)
	require.NoError(t, err)

	// Verify the service is gone
	_, ok := registry.GetServiceConfig(serviceID)
	assert.False(t, ok, "Service should be unregistered")

	// Try to unregister a non-existent service
	err = registry.UnregisterService(context.Background(), "non-existent-service")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestServiceRegistry_GetAllServices(t *testing.T) {
	i := 0
	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			i++
			return &mockUpstream{
				registerFunc: func() (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
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
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("test-service-2")

	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err)
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.NoError(t, err)

	services, err := registry.GetAllServices()
	require.NoError(t, err)
	assert.Len(t, services, 2)
}

func TestServiceRegistry_ServiceInfo(t *testing.T) {
	registry := New(nil, nil, nil, nil, nil)
	serviceID := "test-service"
	serviceInfo := &tool.ServiceInfo{
		Name: "Test Service",
	}

	registry.AddServiceInfo(serviceID, serviceInfo)

	retrievedInfo, ok := registry.GetServiceInfo(serviceID)
	require.True(t, ok)
	assert.Equal(t, serviceInfo, retrievedInfo)

	_, ok = registry.GetServiceInfo("non-existent-service")
	assert.False(t, ok)
}

func TestServiceRegistry_RegisterService_OAuth2(t *testing.T) {
	f := &mockFactory{}
	tm := &mockToolManager{}
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("oauth2-service")
	httpService := &configv1.HttpUpstreamService{}
	httpService.SetAddress("http://localhost")
	serviceConfig.SetHttpService(httpService)

	authConfig := configv1.AuthenticationConfig_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			IssuerUrl: proto.String("https://accounts.google.com"),
			Audience:  proto.String("test-audience"),
		}.Build(),
	}.Build()
	serviceConfig.SetAuthentication(authConfig)

	serviceID, _, _, err := registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err)

	_, ok := am.GetAuthenticator(serviceID)
	assert.True(t, ok, "OAuth2 authenticator should have been added")
}

type mockPromptManager struct {
	prompt.PromptManagerInterface
	ClearPromptsForServiceFunc func(serviceID string)
}

func (m *mockPromptManager) ClearPromptsForService(serviceID string) {
	if m.ClearPromptsForServiceFunc != nil {
		m.ClearPromptsForServiceFunc(serviceID)
	}
}

type mockResourceManager struct {
	resource.ResourceManagerInterface
	ClearResourcesForServiceFunc func(serviceID string)
}

func (m *mockResourceManager) ClearResourcesForService(serviceID string) {
	if m.ClearResourcesForServiceFunc != nil {
		m.ClearResourcesForServiceFunc(serviceID)
	}
}

func TestServiceRegistry_RegisterService_DuplicateName_Cleanup(t *testing.T) {
	var toolClearCount, promptClearCount, resourceClearCount int

	tm := &mockToolManager{
		ClearToolsForServiceFunc: func(serviceID string) {
			toolClearCount++
		},
	}
	prm := &mockPromptManager{
		ClearPromptsForServiceFunc: func(serviceID string) {
			promptClearCount++
		},
	}
	rm := &mockResourceManager{
		ClearResourcesForServiceFunc: func(serviceID string) {
			resourceClearCount++
		},
	}

	f := &mockFactory{
		newUpstreamFunc: func() (upstream.Upstream, error) {
			return &mockUpstream{
				registerFunc: func() (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
					return "duplicate-service", nil, nil, nil
				},
			}, nil
		},
	}
	registry := New(f, tm, prm, rm, auth.NewAuthManager())

	serviceConfig1 := &configv1.UpstreamServiceConfig{}
	serviceConfig1.SetName("test-service")
	_, _, _, err := registry.RegisterService(context.Background(), serviceConfig1)
	require.NoError(t, err)

	// Attempt to register another service with the same name
	serviceConfig2 := &configv1.UpstreamServiceConfig{}
	serviceConfig2.SetName("test-service")
	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig2)
	require.Error(t, err)

	assert.Equal(t, 1, toolClearCount, "ClearToolsForService should be called once for the failed registration")
	assert.Equal(t, 1, promptClearCount, "ClearPromptsForService should be called once for the failed registration")
	assert.Equal(t, 1, resourceClearCount, "ClearResourcesForService should be called once for the failed registration")
}
