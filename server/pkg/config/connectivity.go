// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ConnectivityError encapsulates a connectivity error for a specific service.
type ConnectivityError struct {
	ServiceName string
	Err         error
}

// Error returns the formatted error message.
func (e *ConnectivityError) Error() string {
	return fmt.Sprintf("service %q connectivity check failed: %v", e.ServiceName, e.Err)
}

// CheckConnectivity performs network connectivity checks for the given services.
// It iterates through the list of upstream services and attempts to connect to
// the configured addresses to ensure they are reachable.
func CheckConnectivity(ctx context.Context, config *configv1.McpAnyServerConfig) []ConnectivityError {
	var connectivityErrors []ConnectivityError

	for _, service := range config.GetUpstreamServices() {
		if err := checkServiceConnectivity(ctx, service); err != nil {
			connectivityErrors = append(connectivityErrors, ConnectivityError{
				ServiceName: service.GetName(),
				Err:         err,
			})
		}
	}

	return connectivityErrors
}

func checkServiceConnectivity(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if httpService := service.GetHttpService(); httpService != nil {
		return checkHTTPConnectivity(ctx, httpService.GetAddress())
	} else if websocketService := service.GetWebsocketService(); websocketService != nil {
		return checkTCPConnectivity(ctx, websocketService.GetAddress(), "ws")
	} else if grpcService := service.GetGrpcService(); grpcService != nil {
		return checkTCPConnectivity(ctx, grpcService.GetAddress(), "tcp")
	} else if openapiService := service.GetOpenapiService(); openapiService != nil {
		// For OpenAPI, we check the address if present, or the spec_url if present
		if openapiService.GetAddress() != "" {
			return checkHTTPConnectivity(ctx, openapiService.GetAddress())
		}
		if openapiService.GetSpecUrl() != "" {
			return checkHTTPConnectivity(ctx, openapiService.GetSpecUrl())
		}
	} else if mcpService := service.GetMcpService(); mcpService != nil {
		if httpConn := mcpService.GetHttpConnection(); httpConn != nil {
			return checkHTTPConnectivity(ctx, httpConn.GetHttpAddress())
		}
		// Stdio and Bundle connections are local, so "connectivity" check is moot or already covered by validation (file existence).
	} else if graphqlService := service.GetGraphqlService(); graphqlService != nil {
		return checkHTTPConnectivity(ctx, graphqlService.GetAddress())
	}
	// SQL and others: might need specific drivers which we want to avoid importing if possible,
	// or we can try to parse host:port from DSN. For now, skipping.

	return nil
}

func checkHTTPConnectivity(ctx context.Context, address string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Use a client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 5xx errors indicate the service is reachable but failing.
	// 404 indicates reachable but path wrong.
	// We consider it "reachable" if we get any response.
	// However, if we get 404 on the root, maybe that's fine for some APIs?
	// But usually we want to know if the HOST is reachable.
	// If the user configured "https://api.github.com/invalid", 404 is a valid response from the server.
	// So connection is successful.
	return nil
}

func checkTCPConnectivity(ctx context.Context, address string, scheme string) error {
	// If address contains scheme (like ws://), strip it or parse it
	if scheme == "ws" {
		u, err := url.Parse(address)
		if err == nil {
			address = u.Host
		}
	}

	// If address doesn't have port, might be issue for Dial.
	// But Dial usually expects host:port.
	// If http/https scheme is in address, we need to handle it.
	if u, err := url.Parse(address); err == nil && u.Host != "" {
		address = u.Host
		if u.Port() == "" {
			if u.Scheme == "https" || u.Scheme == "wss" {
				address = net.JoinHostPort(address, "443")
			} else {
				address = net.JoinHostPort(address, "80")
			}
		}
	}

	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}
