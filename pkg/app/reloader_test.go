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

package app

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/pkg/bus"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigReloader(t *testing.T) {
	fs := afero.NewMemMapFs()
	configContent := `
upstream_services:
  - name: service1
    mcp_service:
      http_connection:
        http_address: http://localhost:8080
`
	err := afero.WriteFile(fs, "config.yaml", []byte(configContent), 0644)
	require.NoError(t, err)

	// Setup Bus
	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	bp, err := bus.NewBusProvider(messageBus)
	require.NoError(t, err)

	// Initial config
	initialConfig := &configv1.McpAnyServerConfig{}
	service1 := &configv1.UpstreamServiceConfig{}
	service1.SetName("service1")

	mcpService := &configv1.McpUpstreamService{}
	httpConn := &configv1.McpStreamableHttpConnection{}
	httpConn.SetHttpAddress("http://localhost:8080")
	mcpService.SetHttpConnection(httpConn)
	service1.SetMcpService(mcpService)

	initialConfig.SetUpstreamServices([]*configv1.UpstreamServiceConfig{service1})

	reloader := NewConfigReloader(
		fs,
		[]string{"config.yaml"},
		100*time.Millisecond,
		bp,
		initialConfig,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	reloader.Start(ctx)

	// Listen for registration requests
	reqBus := bus.GetBus[*bus.ServiceRegistrationRequest](bp, bus.ServiceRegistrationRequestTopic)
	reqChan := make(chan *bus.ServiceRegistrationRequest, 10)
	reqBus.Subscribe(ctx, "request", func(req *bus.ServiceRegistrationRequest) {
		reqChan <- req
	})

	// 1. Modify config to add a service
	newConfigContent := `
upstream_services:
  - name: service1
    mcp_service:
      http_connection:
        http_address: http://localhost:8080
  - name: service2
    mcp_service:
      http_connection:
        http_address: http://localhost:8081
`
	err = afero.WriteFile(fs, "config.yaml", []byte(newConfigContent), 0644)
	require.NoError(t, err)

	// Wait for reload
	timeout := time.After(1 * time.Second)
	found := false
	for !found {
		select {
		case req := <-reqChan:
			if req.Config.GetName() == "service2" {
				found = true
			}
		case <-timeout:
			t.Fatal("timeout waiting for service2")
		}
	}

	// 2. Modify config to remove service1
	newConfigContent2 := `
upstream_services:
  - name: service2
    mcp_service:
      http_connection:
        http_address: http://localhost:8081
`
	err = afero.WriteFile(fs, "config.yaml", []byte(newConfigContent2), 0644)
	require.NoError(t, err)

	// Wait for reload (service1 removal)
	timeout = time.After(1 * time.Second)
	found = false
	for !found {
		select {
		case req := <-reqChan:
			if req.Config.GetName() == "service1" {
				assert.True(t, req.Config.GetDisable())
				found = true
			}
		case <-timeout:
			t.Fatal("timeout waiting for service1 removal")
		}
	}
}
