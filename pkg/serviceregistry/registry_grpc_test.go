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
	"net"
	"testing"

	"github.com/mcpany/core/pkg/auth"
	"github.com/mcpany/core/pkg/prompt"
	"github.com/mcpany/core/pkg/resource"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func TestServiceRegistry_RegisterService_gRPC_Reflection(t *testing.T) {
	// Create a mock gRPC server with reflection
	lis, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	defer lis.Close()

	s := grpc.NewServer()
	go s.Serve(lis)
	defer s.Stop()

	reflection.Register(s)

	f := &mockFactory{}
	tm := &mockToolManager{}
	prm := prompt.NewPromptManager()
	rm := resource.NewResourceManager()
	am := auth.NewAuthManager()
	registry := New(f, tm, prm, rm, am)

	serviceConfig := &configv1.UpstreamServiceConfig{}
	serviceConfig.SetName("test-grpc-service")
	grpcService := &configv1.GrpcUpstreamService{}
	grpcService.SetAddress(lis.Addr().String())
	grpcService.SetUseReflection(true)
	serviceConfig.SetGrpcService(grpcService)

	_, _, _, err = registry.RegisterService(context.Background(), serviceConfig)
	require.NoError(t, err, "Service registration with gRPC reflection should succeed")
}
