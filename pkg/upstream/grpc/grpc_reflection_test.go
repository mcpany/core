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

package grpc

import (
	"context"
	"testing"

	"github.com/mcpxy/core/pkg/pool"
	"github.com/mcpxy/core/pkg/prompt"
	"github.com/mcpxy/core/pkg/resource"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestGRPCUpstream_Register_Reflection(t *testing.T) {
	var promptManager prompt.PromptManagerInterface
	var resourceManager resource.ResourceManagerInterface

	server, addr := startMockServer(t)
	defer server.Stop()

	t.Run("successful registration with reflection", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := (configv1.GrpcUpstreamService_builder{
			Address:       &addr,
			UseReflection: proto.Bool(true),
		}).Build()

		serviceConfig := (configv1.UpstreamServiceConfig_builder{
			Name:        proto.String("calculator-service-reflection"),
			GrpcService: grpcService,
		}).Build()

		_, discoveredTools, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.NoError(t, err)
		assert.NotEmpty(t, discoveredTools)
		// We expect 2 tools from the annotations + 2 from the descriptors, and 1 for reflection
		assert.Len(t, tm.ListTools(), 5)
	})

	t.Run("reflection disabled", func(t *testing.T) {
		poolManager := pool.NewManager()
		upstream := NewGRPCUpstream(poolManager)
		tm := NewMockToolManager()

		grpcService := (configv1.GrpcUpstreamService_builder{
			Address:       &addr,
			UseReflection: proto.Bool(false),
		}).Build()

		serviceConfig := (configv1.UpstreamServiceConfig_builder{
			Name:        proto.String("calculator-service-no-reflection"),
			GrpcService: grpcService,
		}).Build()

		_, _, err := upstream.Register(context.Background(), serviceConfig, tm, promptManager, resourceManager, false)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse proto definitions")
	})
}
