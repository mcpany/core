// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
)

const schemeTCP = "tcp"

// RunDiagnostics performs runtime connectivity and health checks on the configured services.
// It tries to connect to the upstream services to ensure they are reachable.
//
// Parameters:
//   ctx: Context for cancellation.
//   config: The server configuration to diagnose.
//
// Returns:
//   []ValidationError: A list of actionable errors if any services are unreachable.
func RunDiagnostics(ctx context.Context, config *configv1.McpAnyServerConfig) []ValidationError {
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		errors []ValidationError
	)

	// We use a shorter timeout for diagnostics to not block startup too long.
	// 2 seconds should be enough for local/LAN services.
	diagCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	for _, service := range config.GetUpstreamServices() {
		if service.GetDisable() {
			continue
		}

		wg.Add(1)
		go func(s *configv1.UpstreamServiceConfig) {
			defer wg.Done()
			if err := diagnoseService(diagCtx, s); err != nil {
				mu.Lock()
				errors = append(errors, ValidationError{
					ServiceName: s.GetName(),
					Err:         err,
				})
				mu.Unlock()
			}
		}(service)
	}

	wg.Wait()
	return errors
}

func diagnoseService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	// Only checking HTTP and TCP-based services for now.
	// We can extend this to others later.

	var target string
	var checkType string // "http" or "tcp"

	if httpSvc := service.GetHttpService(); httpSvc != nil {
		target = httpSvc.GetAddress()
		checkType = schemeHTTP
	} else if wsSvc := service.GetWebsocketService(); wsSvc != nil {
		target = wsSvc.GetAddress()
		checkType = schemeTCP // WS uses TCP
	} else if grpcSvc := service.GetGrpcService(); grpcSvc != nil {
		target = grpcSvc.GetAddress()
		checkType = schemeTCP
	} else if graphQLSvc := service.GetGraphqlService(); graphQLSvc != nil {
		target = graphQLSvc.GetAddress()
		checkType = schemeHTTP
	} else if openapiSvc := service.GetOpenapiService(); openapiSvc != nil {
		// OpenAPIService might rely on spec_url or address
		if openapiSvc.GetAddress() != "" {
			target = openapiSvc.GetAddress()
			checkType = schemeHTTP
		} else if openapiSvc.GetSpecUrl() != "" {
			target = openapiSvc.GetSpecUrl()
			checkType = schemeHTTP
		}
	} else if mcpSvc := service.GetMcpService(); mcpSvc != nil {
		if httpConn := mcpSvc.GetHttpConnection(); httpConn != nil {
			target = httpConn.GetHttpAddress()
			checkType = schemeHTTP
		}
	} else if webrtcSvc := service.GetWebrtcService(); webrtcSvc != nil {
		target = webrtcSvc.GetAddress()
		checkType = schemeHTTP // Signaling usually HTTP/WS
	}

	if target == "" {
		return nil
	}

	return checkReachability(ctx, checkType, target)
}

func checkReachability(ctx context.Context, checkType, target string) error {
	log := logging.GetLogger().With("component", "diagnostics")

	u, err := url.Parse(target)
	if err != nil {
		// If simple host:port for TCP/gRPC
		if checkType == schemeTCP {
			// Try parsing as host:port directly if URL parse fails or scheme missing?
			// But grpc service address usually doesn't have scheme in config?
			// Let's assume it might be just "localhost:50051".
			u = &url.URL{Host: target}
		} else {
			return fmt.Errorf("invalid url format: %v", err)
		}
	}

	host := u.Host
	if host == "" {
		// If parsing failed to extract host (e.g. "localhost:50051" without scheme parses as Path), try to assume it's host:port
		if checkType == schemeTCP {
			host = target
		} else {
			return nil // Cannot check
		}
	}

	// If no port is specified, infer from scheme
	if _, _, err := net.SplitHostPort(host); err != nil {
		// Missing port
		port := "80"
		if u.Scheme == schemeHTTPS || u.Scheme == "wss" {
			port = "443"
		}
		host = net.JoinHostPort(host, port)
	}

	log.Debug("Checking connectivity", "target", host, "type", checkType)

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", host)
	if err != nil {
		return &ActionableError{
			Err:        err,
			Suggestion: fmt.Sprintf("Verify that the service at '%s' is running and reachable from the MCP Any server.\n       If using Docker, ensure you are using the correct hostname (e.g., 'host.docker.internal' instead of 'localhost').", target),
		}
	}
	_ = conn.Close()
	return nil
}
