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

package health

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/mcpxy/core/pkg/logging"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// HTTPServiceWithHealthCheck is an interface for services that have an address and an HTTP health check.
type HTTPServiceWithHealthCheck interface {
	GetAddress() string
	GetHealthCheck() *configv1.HttpHealthCheck
}

// Check checks the health of the upstream service based on the provided configuration.
func Check(ctx context.Context, uc *configv1.UpstreamServiceConfig) bool {
	if uc == nil {
		return false
	}

	if c := uc.GetHttpService(); c != nil {
		return checkHTTPHealth(ctx, c)
	}
	if c := uc.GetGrpcService(); c != nil {
		return checkGRPCHealth(ctx, c)
	}
	if c := uc.GetOpenapiService(); c != nil {
		return checkOpenAPIHealth(ctx, c)
	}
	if c := uc.GetCommandLineService(); c != nil {
		return checkCommandLineHealth(ctx, c)
	}
	if c := uc.GetWebsocketService(); c != nil {
		return checkWebSocketHealth(ctx, c)
	}
	if c := uc.GetWebrtcService(); c != nil {
		return checkWebRTCHealth(ctx, c)
	}
	if c := uc.GetMcpService(); c != nil {
		return checkMCPHealth(ctx, c)
	}

	return false
}

func checkHTTPHealth(ctx context.Context, c HTTPServiceWithHealthCheck) bool {
	if c.GetHealthCheck() == nil {
		return checkConnection(ctx, c.GetAddress())
	}

	client := &http.Client{
		Timeout: lo.Ternary(c.GetHealthCheck().GetTimeout() != nil, c.GetHealthCheck().GetTimeout().AsDuration(), 5*time.Second),
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.GetHealthCheck().GetUrl(), nil)
	if err != nil {
		logging.GetLogger().Warn("Failed to create health check request", "error", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		logging.GetLogger().Warn("Health check failed", "error", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != int(c.GetHealthCheck().GetExpectedCode()) {
		logging.GetLogger().Warn("Health check failed", "status_code", resp.StatusCode)
		return false
	}

	logging.GetLogger().Info("Health check successful")
	return true
}

func checkGRPCHealth(ctx context.Context, c *configv1.GrpcUpstreamService) bool {
	if c.GetHealthCheck() == nil {
		return checkConnection(ctx, c.GetAddress())
	}

	conn, err := grpc.DialContext(ctx, c.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logging.GetLogger().Warn("Failed to connect to gRPC service", "error", err)
		return false
	}
	defer conn.Close()

	healthClient := healthpb.NewHealthClient(conn)
	resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: c.GetHealthCheck().GetService()})
	if err != nil {
		logging.GetLogger().Warn("gRPC health check failed", "error", err)
		return false
	}

	if resp.Status != healthpb.HealthCheckResponse_SERVING {
		logging.GetLogger().Warn("gRPC service is not serving", "status", resp.Status)
		return false
	}

	logging.GetLogger().Info("gRPC health check successful")
	return true
}

func checkOpenAPIHealth(ctx context.Context, c *configv1.OpenapiUpstreamService) bool {
	return checkHTTPHealth(ctx, c)
}

func checkCommandLineHealth(ctx context.Context, c *configv1.CommandLineUpstreamService) bool {
	// For command line services, we assume it's healthy if the command can be executed.
	// A more sophisticated check would involve running a specific command and checking the output.
	return true
}

func checkWebSocketHealth(ctx context.Context, c *configv1.WebsocketUpstreamService) bool {
	return checkConnection(ctx, c.GetAddress())
}

func checkWebRTCHealth(ctx context.Context, c *configv1.WebrtcUpstreamService) bool {
	// For WebRTC, a health check would typically involve the signaling server.
	return checkConnection(ctx, c.GetAddress())
}

func checkMCPHealth(ctx context.Context, c *configv1.McpUpstreamService) bool {
	if conn := c.GetHttpConnection(); conn != nil {
		return checkConnection(ctx, conn.GetHttpAddress())
	}
	if c.GetStdioConnection() != nil {
		return true // Assume healthy
	}
	return false
}

func checkConnection(ctx context.Context, address string) bool {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		logging.GetLogger().Warn("Failed to connect to address", "address", address, "error", err)
		return false
	}
	defer conn.Close()
	logging.GetLogger().Info("Connection successful", "address", address)
	return true
}
