// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// CheckConnectivity validates the connectivity of upstream services defined in the configuration.
// It performs a "dry run" connection to ensure the services are reachable.
func CheckConnectivity(ctx context.Context, config *configv1.McpAnyServerConfig) []ValidationError {
	var validationErrors []ValidationError

	// We use a timeout context for all checks to ensure we don't block startup for too long.
	// Individual checks also have their own timeouts.
	// 5 seconds total for all checks might be too short if there are many,
	// but we run them sequentially here for simplicity.
	// Ideally, this should be parallelized if we have many services.
	// For now, let's keep it simple and sequential but with short timeouts per service.

	for _, service := range config.GetUpstreamServices() {
		// Skip disabled services
		if service.GetDisable() {
			continue
		}

		if err := checkServiceConnectivity(ctx, service); err != nil {
			validationErrors = append(validationErrors, ValidationError{
				ServiceName: service.GetName(),
				Err:         fmt.Errorf("connectivity check failed: %w", err),
			})
		}
	}

	return validationErrors
}

func checkServiceConnectivity(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	// 2 seconds timeout per service check
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if httpService := service.GetHttpService(); httpService != nil {
		return checkURL(checkCtx, httpService.GetAddress())
	} else if graphqlService := service.GetGraphqlService(); graphqlService != nil {
		return checkURL(checkCtx, graphqlService.GetAddress())
	} else if openapiService := service.GetOpenapiService(); openapiService != nil {
		// OpenApi can have Address or SpecURL
		if openapiService.GetAddress() != "" {
			return checkURL(checkCtx, openapiService.GetAddress())
		}
		if openapiService.GetSpecUrl() != "" {
			return checkURL(checkCtx, openapiService.GetSpecUrl())
		}
	} else if grpcService := service.GetGrpcService(); grpcService != nil {
		return checkGRPC(checkCtx, grpcService.GetAddress())
	} else if webrtcService := service.GetWebrtcService(); webrtcService != nil {
		return checkURL(checkCtx, webrtcService.GetAddress())
	} else if mcpService := service.GetMcpService(); mcpService != nil {
		if httpConn := mcpService.GetHttpConnection(); httpConn != nil {
			return checkURL(checkCtx, httpConn.GetHttpAddress())
		}
		// Stdio checks are done via file existence validation
	}

	// SQL and others omitted for now to avoid dependency complexity or because they require specific drivers.
	return nil
}

func checkURL(ctx context.Context, address string) error {
	// Parse URL to ensure it's valid
	_, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	// Create a request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use a custom transport to disable keep-alives and set timeouts
	transport := &http.Transport{
		DisableKeepAlives: true,
		DialContext: (&net.Dialer{
			Timeout:   2 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout: 2 * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true}, // We accept any cert for connectivity check
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		// If HEAD fails (some servers block it), try GET
		req.Method = http.MethodGet
		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("unreachable: %w", err)
		}
	}
	defer resp.Body.Close()

	// Any status code indicates connectivity (even 404, 401, 500)
	// We only care that we reached *something*.
	return nil
}

func checkGRPC(ctx context.Context, address string) error {
	// Try to dial the gRPC server
	// We use WithBlock to wait for connection
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()), // Try insecure first?
		grpc.WithBlock(),
	}

	// Check if address implies TLS (not standard in gRPC addresses, usually handled by creds)
	// But without knowing if upstream is TLS or not, this is tricky.
	// Most internal gRPC is insecure, public is secure.
	// If we fail with insecure, should we try secure?
	// For now, let's just try to dial. If it fails, it fails.
	// Actually, `grpc.Dial` might succeed even if handshake fails later?
	// WithBlock ensures we wait for handshake.

	conn, err := grpc.DialContext(ctx, address, opts...)
	if err != nil {
		// Try with TLS if insecure failed?
		// But error from DialContext with Block usually means network unreachable or handshake fail.
		// If it was "protocol error", maybe we mismatch TLS.
		// Let's retry with TLS credentials if first attempt fails and error suggests it?
		// Or just try TLS if the first one failed.

		// Reset context for retry
		// Actually, we can't easily reuse the timeout context if it expired.
		if ctx.Err() != nil {
			return fmt.Errorf("timeout connecting to %s", address)
		}

		// Retry with TLS
		tlsOpts := []grpc.DialOption{
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})),
			grpc.WithBlock(),
		}
		connTLS, errTLS := grpc.DialContext(ctx, address, tlsOpts...)
		if errTLS != nil {
			// Return original error or combined?
			return fmt.Errorf("failed to connect (tried insecure and tls): %v / %v", err, errTLS)
		}
		conn = connTLS
	}

	if conn != nil {
		conn.Close()
	}
	return nil
}
