/*
 * Copyright 2025 Author(s) of MCP-XY
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

package mcpserver

import (
	"context"
	"testing"

	"github.com/mcpxy/core/pkg/auth"
	"github.com/mcpxy/core/pkg/bus"
	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	"github.com/mcpxy/core/pkg/serviceregistry"
	"github.com/mcpxy/core/pkg/tool"
	"github.com/mcpxy/core/pkg/upstream/factory"
	"github.com/mcpxy/core/pkg/worker"
	v1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRegistrationServer_RegisterService(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup bus and worker
	busProvider := bus.NewBusProvider()

	// Setup components
	poolManager := pool.NewManager()
	upstreamFactory := factory.NewUpstreamServiceFactory(poolManager)
	toolManager := tool.NewToolManager(busProvider)
	promptManager := prompt.NewPromptManager()
	resourceManager := resource.NewResourceManager()
	authManager := auth.NewAuthManager()
	serviceRegistry := serviceregistry.New(upstreamFactory, toolManager, promptManager, resourceManager, authManager)
	registrationWorker := worker.NewServiceRegistrationWorker(busProvider, serviceRegistry)
	registrationWorker.Start(ctx)

	// Setup server
	registrationServer, err := NewRegistrationServer(busProvider)
	require.NoError(t, err)

	t.Run("successful registration", func(t *testing.T) {
		serviceName := "testservice"
		config := &configv1.UpstreamServiceConfig{}
		config.SetName(serviceName)
		config.SetHttpService(&configv1.HttpUpstreamService{})

		req := &v1.RegisterServiceRequest{}
		req.SetConfig(config)

		resp, err := registrationServer.RegisterService(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Contains(t, resp.GetMessage(), "registered successfully")

		// Verify that the service info was added to the tool manager
		serviceInfo, ok := toolManager.GetServiceInfo("testservice")
		require.True(t, ok)
		require.NotNil(t, serviceInfo)
		assert.Equal(t, "testservice", serviceInfo.Name)
	})

	t.Run("missing config", func(t *testing.T) {
		req := &v1.RegisterServiceRequest{}
		_, err := registrationServer.RegisterService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	t.Run("missing config name", func(t *testing.T) {
		config := &configv1.UpstreamServiceConfig{}
		req := &v1.RegisterServiceRequest{}
		req.SetConfig(config)
		_, err := registrationServer.RegisterService(ctx, req)
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}
