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

package framework

import (
	"context"
	"fmt"
	"net"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	"github.com/mcpany/core/tests/integration"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type mockRegistrationService struct {
	apiv1.UnimplementedRegistrationServiceServer
	t *testing.T
}

func (s *mockRegistrationService) RegisterService(ctx context.Context, req *apiv1.RegisterServiceRequest) (*apiv1.RegisterServiceResponse, error) {
	s.t.Logf("mockRegistrationService.RegisterService called with name: %s", req.GetConfig().GetName())
	resp := &apiv1.RegisterServiceResponse{}
	resp.SetMessage("Service registered successfully")
	return resp, nil
}

func setupTestServer(t *testing.T) (apiv1.RegistrationServiceClient, func()) {
	l, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	apiv1.RegisterRegistrationServiceServer(grpcServer, &mockRegistrationService{t: t})

	go func() {
		// This will return an error when the listener is closed.
		_ = grpcServer.Serve(l)
	}()

	conn, err := grpc.Dial(l.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)

	client := apiv1.NewRegistrationServiceClient(conn)
	return client, func() {
		conn.Close()
		grpcServer.Stop()
		l.Close()
	}
}

func TestBuildAndRegisterServices(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		buildFunc        func(t *testing.T) *integration.ManagedProcess
		registerFunc     func(t *testing.T, client apiv1.RegistrationServiceClient, endpoint string)
		requiresUpstream bool
		serviceType      string
	}{
		{
			name:             "GRPCWeatherServer",
			buildFunc:        BuildGRPCWeatherServer,
			registerFunc:     RegisterGRPCWeatherService,
			requiresUpstream: true,
			serviceType:      "grpc",
		},
		{
			name:             "GRPCAuthedWeatherServer",
			buildFunc:        BuildGRPCAuthedWeatherServer,
			registerFunc:     RegisterGRPCAuthedWeatherService,
			requiresUpstream: true,
			serviceType:      "grpc",
		},
		{
			name:             "WebsocketWeatherServer",
			buildFunc:        BuildWebsocketWeatherServer,
			registerFunc:     RegisterWebsocketWeatherService,
			requiresUpstream: true,
			serviceType:      "websocket",
		},
		{
			name:             "WebrtcWeatherServer",
			buildFunc:        BuildWebrtcWeatherServer,
			registerFunc:     RegisterWebrtcWeatherService,
			requiresUpstream: true,
			serviceType:      "webrtc",
		},
		{
			name:             "StdioServer",
			buildFunc:        BuildStdioServer,
			registerFunc:     RegisterStdioService,
			requiresUpstream: false,
		},
		{
			name:             "StdioDockerServer",
			buildFunc:        BuildStdioDockerServer,
			registerFunc:     RegisterStdioDockerService,
			requiresUpstream: false,
		},
		{
			name:             "OpenAPIWeatherServer",
			buildFunc:        BuildOpenAPIWeatherServer,
			registerFunc:     RegisterOpenAPIWeatherService,
			requiresUpstream: true,
			serviceType:      "openapi",
		},
		{
			name:             "OpenAPIAuthedServer",
			buildFunc:        BuildOpenAPIAuthedServer,
			registerFunc:     RegisterOpenAPIAuthedService,
			requiresUpstream: true,
			serviceType:      "openapi",
		},
		{
			name:             "StreamableHTTPServer",
			buildFunc:        BuildStreamableHTTPServer,
			registerFunc:     RegisterStreamableHTTPService,
			requiresUpstream: true,
			serviceType:      "streamablehttp",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, cleanup := setupTestServer(t)
			defer cleanup()

			var proc *integration.ManagedProcess
			if tc.buildFunc != nil {
				proc = tc.buildFunc(t)
				if proc != nil {
					err := proc.Start()
					require.NoError(t, err, "Failed to start managed process")
					defer proc.Stop()
					if proc.Port != 0 {
						integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)
					}
				}
			}

			endpoint := ""
			if tc.requiresUpstream && proc != nil {
				if proc.Port != 0 {
					endpoint = fmt.Sprintf("localhost:%d", proc.Port)
					switch tc.serviceType {
					case "websocket":
						endpoint = fmt.Sprintf("ws://%s/echo", endpoint)
					case "webrtc":
						endpoint = fmt.Sprintf("http://%s/signal", endpoint)
					case "openapi", "streamablehttp":
						endpoint = fmt.Sprintf("http://%s", endpoint)
					}
				}
			}

			if tc.registerFunc != nil {
				tc.registerFunc(t, client, endpoint)
			}
		})
	}
}

func TestBuildAndRegisterHttpServices(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		buildFunc    func(t *testing.T) *integration.ManagedProcess
		registerFunc func(t *testing.T, client apiv1.RegistrationServiceClient, endpoint string)
	}{
		{
			name:         "HTTPEchoServer",
			buildFunc:    BuildHTTPEchoServer,
			registerFunc: RegisterHTTPEchoService,
		},
		{
			name:         "HTTPAuthedEchoServer",
			buildFunc:    BuildHTTPAuthedEchoServer,
			registerFunc: RegisterHTTPAuthedEchoService,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, cleanup := setupTestServer(t)
			defer cleanup()

			proc := tc.buildFunc(t)
			err := proc.Start()
			require.NoError(t, err, "Failed to start managed process")
			defer proc.Stop()

			integration.WaitForTCPPort(t, proc.Port, integration.ServiceStartupTimeout)

			endpoint := fmt.Sprintf("http://localhost:%d", proc.Port)
			tc.registerFunc(t, client, endpoint)
		})
	}
}
