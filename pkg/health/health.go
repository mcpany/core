// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/mcpany/core/pkg/logging"
	"github.com/mcpany/core/pkg/metrics"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

const (
	healthStatusGauge = "mcp_any_health_check_status"
)

// HTTPServiceWithHealthCheck is an interface for services that have an address and an HTTP health check.
type HTTPServiceWithHealthCheck interface {
	GetAddress() string
	GetHealthCheck() *configv1.HttpHealthCheck
}

// NewChecker creates a new health checker for the given upstream service.
func NewChecker(uc *configv1.UpstreamServiceConfig) health.Checker {
	if uc == nil {
		return nil
	}

	var check health.Check
	serviceName := uc.GetName()

	switch uc.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_HttpService_case:
		check = httpCheck(serviceName, uc.GetHttpService())
	case configv1.UpstreamServiceConfig_GrpcService_case:
		check = grpcCheck(serviceName, uc.GetGrpcService())
	case configv1.UpstreamServiceConfig_OpenapiService_case:
		check = httpCheck(serviceName, uc.GetOpenapiService())
	case configv1.UpstreamServiceConfig_CommandLineService_case:
		check = commandLineCheck(serviceName, uc.GetCommandLineService())
	case configv1.UpstreamServiceConfig_WebsocketService_case:
		check = connectionCheck(serviceName, uc.GetWebsocketService().GetAddress())
	case configv1.UpstreamServiceConfig_WebrtcService_case:
		check = connectionCheck(serviceName, uc.GetWebrtcService().GetAddress())
	case configv1.UpstreamServiceConfig_McpService_case:
		check = mcpCheck(serviceName, uc.GetMcpService())
	default:
		return nil
	}

	opts := []health.CheckerOption{
		health.WithStatusListener(func(ctx context.Context, state health.CheckerState) {
			status := float32(0.0)
			if state.Status == health.StatusUp {
				status = 1.0
			}
			metrics.SetGauge(healthStatusGauge, status, serviceName)
			logging.GetLogger().Info("health status changed", "service", serviceName, "status", state.Status)
		}),
		// Using synchronous checks for now to simplify the implementation and ensure
		// tests are reliable. Periodic checks can be re-introduced later if needed,
		// likely controlled by a configuration option.
		health.WithCheck(check),
	}

	return health.NewChecker(opts...)
}

func httpCheck(name string, c HTTPServiceWithHealthCheck) health.Check {
	return health.Check{
		Name:    name,
		Timeout: 5 * time.Second,
		Check: func(ctx context.Context) error {
			if c.GetHealthCheck() == nil {
				return checkConnection(c.GetAddress())
			}

			client := &http.Client{
				Timeout: lo.Ternary(c.GetHealthCheck().GetTimeout() != nil, c.GetHealthCheck().GetTimeout().AsDuration(), 5*time.Second),
			}

			req, err := http.NewRequestWithContext(ctx, "GET", c.GetHealthCheck().GetUrl(), nil)
			if err != nil {
				return fmt.Errorf("failed to create health check request: %w", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				return fmt.Errorf("health check failed: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != int(c.GetHealthCheck().GetExpectedCode()) {
				return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
			}
			return nil
		},
	}
}

func grpcCheck(name string, c *configv1.GrpcUpstreamService) health.Check {
	return health.Check{
		Name:    name,
		Timeout: 5 * time.Second,
		Check: func(ctx context.Context) error {
			if c.GetHealthCheck() == nil {
				return checkConnection(c.GetAddress())
			}

			conn, err := grpc.DialContext(ctx, c.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return fmt.Errorf("failed to connect to gRPC service: %w", err)
			}
			defer conn.Close()

			healthClient := healthpb.NewHealthClient(conn)
			resp, err := healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: c.GetHealthCheck().GetService()})
			if err != nil {
				return fmt.Errorf("gRPC health check failed: %w", err)
			}

			if resp.Status != healthpb.HealthCheckResponse_SERVING {
				return fmt.Errorf("gRPC service is not serving, status: %s", resp.Status)
			}
			return nil
		},
	}
}

func commandLineCheck(name string, c *configv1.CommandLineUpstreamService) health.Check {
	return health.Check{
		Name: name,
		Check: func(ctx context.Context) error {
			// For command line services, we assume it's healthy if not otherwise configured.
			// A more sophisticated check would involve running a specific command and checking the output.
			if c.GetHealthCheck() == nil {
				return nil
			}
			// NOTE: StdioHealthCheck is not implemented yet.
			return nil
		},
	}
}

func connectionCheck(name, address string) health.Check {
	return health.Check{
		Name:    name,
		Timeout: 5 * time.Second,
		Check: func(ctx context.Context) error {
			return checkConnection(address)
		},
	}
}

func mcpCheck(name string, c *configv1.McpUpstreamService) health.Check {
	return health.Check{
		Name: name,
		Check: func(ctx context.Context) error {
			if conn := c.GetHttpConnection(); conn != nil {
				return checkConnection(conn.GetHttpAddress())
			}
			if c.GetStdioConnection() != nil {
				return nil // Assume healthy
			}
			return fmt.Errorf("no connection configured for MCP service")
		},
	}
}

func checkConnection(address string) error {
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to address %s: %w", address, err)
	}
	defer conn.Close()
	return nil
}
