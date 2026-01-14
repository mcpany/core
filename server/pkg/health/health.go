// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package health provides health check functionality.
package health

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/alexliesenfeld/health"
	"github.com/coder/websocket"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/command"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/metrics"
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
	// GetAddress returns the address of the service.
	GetAddress() string
	// GetHealthCheck returns the HTTP health check configuration for the service.
	GetHealthCheck() *configv1.HttpHealthCheck
}

// NewChecker creates a new health checker for the given upstream service.
// It determines the type of service (HTTP, gRPC, etc.) and creates an appropriate
// health check strategy.
//
// Parameters:
//   - uc: The configuration of the upstream service to check.
//
// Returns:
//   - A health.Checker instance that can be used to monitor the service's health.
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
		check = websocketCheck(serviceName, uc.GetWebsocketService())
	case configv1.UpstreamServiceConfig_WebrtcService_case:
		check = webrtcCheck(serviceName, uc.GetWebrtcService())
	case configv1.UpstreamServiceConfig_McpService_case:
		check = mcpCheck(serviceName, uc.GetMcpService())
	case configv1.UpstreamServiceConfig_FilesystemService_case:
		check = filesystemCheck(serviceName, uc.GetFilesystemService())
	default:
		return nil
	}

	opts := []health.CheckerOption{
		health.WithStatusListener(func(_ context.Context, state health.CheckerState) {
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
		// Cache the health check result for a short duration to avoid spamming the upstream
		// if IsHealthy is called frequently (e.g. by the pool).
		health.WithCacheDuration(1 * time.Second),
	}

	return health.NewChecker(opts...)
}

func httpCheckFunc(ctx context.Context, _ string, hc *configv1.HttpHealthCheck) error {
	if hc == nil {
		return nil
	}

	client := &http.Client{
		Timeout: lo.Ternary(hc.GetTimeout() != nil, hc.GetTimeout().AsDuration(), 5*time.Second),
	}

	method := lo.Ternary(hc.GetMethod() != "", hc.GetMethod(), http.MethodGet)
	req, err := http.NewRequestWithContext(ctx, method, hc.GetUrl(), nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != int(hc.GetExpectedCode()) {
		return fmt.Errorf("health check failed with status code: %d", resp.StatusCode)
	}

	if hc.GetExpectedResponseBodyContains() != "" {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read health check response body: %w", err)
		}
		if !strings.Contains(string(body), hc.GetExpectedResponseBodyContains()) {
			return fmt.Errorf("health check response body does not contain expected string")
		}
	}
	return nil
}

func httpCheck(name string, c HTTPServiceWithHealthCheck) health.Check {
	return health.Check{
		Name:    name,
		Timeout: 5 * time.Second,
		Check: func(ctx context.Context) error {
			return httpCheckFunc(ctx, c.GetAddress(), c.GetHealthCheck())
		},
	}
}

func webrtcCheck(name string, c *configv1.WebrtcUpstreamService) health.Check {
	return health.Check{
		Name:    name,
		Timeout: 5 * time.Second,
		Check: func(ctx context.Context) error {
			// For WebRTC, the health check is primarily concerned with the signaling
			// server, which is typically an HTTP or WebSocket endpoint.
			if hc := c.GetHealthCheck(); hc != nil {
				if httpCheck := hc.GetHttp(); httpCheck != nil {
					return httpCheckFunc(ctx, c.GetAddress(), httpCheck)
				}
				if wsCheck := hc.GetWebsocket(); wsCheck != nil {
					return websocketCheckFunc(ctx, c.GetAddress(), wsCheck)
				}
			}
			return checkConnection(ctx, c.GetAddress())
		},
	}
}

func websocketCheck(name string, c *configv1.WebsocketUpstreamService) health.Check {
	return health.Check{
		Name:    name,
		Timeout: 5 * time.Second,
		Check: func(ctx context.Context) error {
			return websocketCheckFunc(ctx, c.GetAddress(), c.GetHealthCheck())
		},
	}
}

func websocketCheckFunc(ctx context.Context, address string, hc *configv1.WebsocketHealthCheck) error {
	if hc == nil {
		return checkConnection(ctx, address)
	}

	healthCheckURL := hc.GetUrl()
	if healthCheckURL == "" {
		healthCheckURL = address
	}
	// Address/URL should start with ws:// or wss://
	// If it doesn't, assume ws:// if it looks like an address, but usually URL field should handle it.
	// We'll trust the URL field mostly, but if it came from address it might lack scheme.
	if !strings.HasPrefix(healthCheckURL, "ws://") && !strings.HasPrefix(healthCheckURL, "wss://") {
		healthCheckURL = "ws://" + healthCheckURL
	}

	timeout := lo.Ternary(hc.GetTimeout() != nil, hc.GetTimeout().AsDuration(), 5*time.Second)

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conn, resp, err := websocket.Dial(ctx, healthCheckURL, nil)
	if resp != nil {
		defer func() {
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
		}()
	}
	if err != nil {
		logging.GetLogger().Error("WebSocket health check failed", "url", healthCheckURL, "error", err)
		return fmt.Errorf("WebSocket health check failed: %w", err)
	}
	defer func() { _ = conn.Close(websocket.StatusNormalClosure, "") }()

	if hc.GetMessage() != "" {
		err = conn.Write(ctx, websocket.MessageText, []byte(hc.GetMessage()))
		if err != nil {
			logging.GetLogger().Error("Failed to write to websocket", "error", err)
			return fmt.Errorf("failed to write message to websocket: %w", err)
		}
	}

	if hc.GetExpectedResponseContains() != "" {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			logging.GetLogger().Error("Failed to read from websocket", "error", err)
			return fmt.Errorf("failed to read message from websocket: %w", err)
		}
		if !strings.Contains(string(msg), hc.GetExpectedResponseContains()) {
			logging.GetLogger().Error("Websocket response mismatch", "expected", hc.GetExpectedResponseContains(), "actual", string(msg))
			return fmt.Errorf(
				"websocket health check response did not contain expected string: %s",
				hc.GetExpectedResponseContains(),
			)
		}
	}

	return nil
}

func grpcCheck(name string, c *configv1.GrpcUpstreamService) health.Check {
	return health.Check{
		Name:    name,
		Timeout: 5 * time.Second,
		Check: func(ctx context.Context) error {
			if c.GetHealthCheck() == nil {
				return checkConnection(ctx, c.GetAddress())
			}

			conn, err := grpc.NewClient(
				c.GetAddress(),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			if err != nil {
				return fmt.Errorf("failed to connect to gRPC service: %w", err)
			}
			defer func() { _ = conn.Close() }()

			healthClient := healthpb.NewHealthClient(conn)
			resp, err := healthClient.Check(
				ctx,
				&healthpb.HealthCheckRequest{Service: c.GetHealthCheck().GetService()},
			)
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

			healthCheck := c.GetHealthCheck()
			executor := command.NewExecutor(c.GetContainerEnvironment())
			var args []string
			if healthCheck.GetMethod() != "" {
				args = append(args, healthCheck.GetMethod())
			}
			if healthCheck.GetPrompt() != "" {
				args = append(args, healthCheck.GetPrompt())
			}

			stdout, _, exitCodeChan, err := executor.Execute(
				ctx,
				c.GetCommand(),
				args,
				c.GetWorkingDirectory(),
				nil,
			)
			if err != nil {
				return fmt.Errorf("failed to execute health check command: %w", err)
			}

			var stdoutBuf bytes.Buffer
			_, _ = io.Copy(&stdoutBuf, stdout)
			exitCode := <-exitCodeChan

			if exitCode != 0 {
				return fmt.Errorf("health check command failed with exit code: %d", exitCode)
			}

			if healthCheck.GetExpectedResponseContains() != "" &&
				!strings.Contains(stdoutBuf.String(), healthCheck.GetExpectedResponseContains()) {
				return fmt.Errorf(
					"health check response did not contain expected string: %s",
					healthCheck.GetExpectedResponseContains(),
				)
			}

			return nil
		},
	}
}

func mcpCheck(name string, c *configv1.McpUpstreamService) health.Check {
	return health.Check{
		Name: name,
		Check: func(ctx context.Context) error {
			if conn := c.GetHttpConnection(); conn != nil {
				return checkConnection(ctx, conn.GetHttpAddress())
			}
			if c.GetStdioConnection() != nil {
				return nil // Assume healthy
			}
			return fmt.Errorf("no connection configured for MCP service")
		},
	}
}

func filesystemCheck(name string, c *configv1.FilesystemUpstreamService) health.Check {
	return health.Check{
		Name: name,
		Check: func(_ context.Context) error {
			// Basic check: Ensure root paths exist (for local OS FS)
			// Only check local paths if it's explicitly OsFs or not specified (default)
			// For remote filesystems (S3, GCS), we would need client-specific checks,
			// which are harder to implement here without the provider instance.
			// So we focus on Local FS for now which is the most common use case for "health" of FS.
			isLocal := false
			switch c.FilesystemType.(type) {
			case *configv1.FilesystemUpstreamService_Os:
				isLocal = true
			case nil:
				// Fallback behavior matches provider logic
				isLocal = true
			}

			if isLocal {
				for virtualPath, localPath := range c.GetRootPaths() {
					if _, err := os.Stat(localPath); err != nil {
						return fmt.Errorf("root path check failed for %s (%s): %w", virtualPath, localPath, err)
					}
				}
			}
			// For other types, assume healthy for now or add specific checks later.
			return nil
		},
	}
}

func checkConnection(ctx context.Context, address string) error {
	var target string
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err != nil {
			return fmt.Errorf("failed to parse address %s: %w", address, err)
		}
		host := u.Hostname()
		port := u.Port()
		if port == "" {
			if u.Scheme == "https" {
				port = "443"
			} else {
				port = "80"
			}
		}
		target = net.JoinHostPort(host, port)
	} else {
		// If no scheme, try to parse as host:port. If no port, assume 80.
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			// If SplitHostPort fails, it means no port was specified.
			// Assume it's just a hostname and default to port 80.
			host = address
			port = "80"
		}
		target = net.JoinHostPort(host, port)
	}

	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", target)
	if err != nil {
		logging.GetLogger().Error("checkConnection failed", "address", target, "error", err)
		return fmt.Errorf("failed to connect to address %s: %w", target, err)
	}
	defer func() { _ = conn.Close() }()
	return nil
}
