// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// validateServiceConnection performs a network connectivity check for the given service.
// It attempts to establish a TCP connection to the configured address.
func validateServiceConnection(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	// Skip if disabled (though Validate runs before disable check usually, but good practice)
	if service.GetDisable() {
		return nil
	}

	var addressToCheck string

	switch service.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_HttpService_case:
		addressToCheck = service.GetHttpService().GetAddress()
	case configv1.UpstreamServiceConfig_WebsocketService_case:
		addressToCheck = service.GetWebsocketService().GetAddress()
	case configv1.UpstreamServiceConfig_GrpcService_case:
		addressToCheck = service.GetGrpcService().GetAddress()
	case configv1.UpstreamServiceConfig_OpenapiService_case:
		// Check Address if present
		if service.GetOpenapiService().GetAddress() != "" {
			addressToCheck = service.GetOpenapiService().GetAddress()
		} else if service.GetOpenapiService().GetSpecUrl() != "" {
			// Also check SpecURL if it's a network URL
			addressToCheck = service.GetOpenapiService().GetSpecUrl()
		}
	case configv1.UpstreamServiceConfig_GraphqlService_case:
		addressToCheck = service.GetGraphqlService().GetAddress()
	case configv1.UpstreamServiceConfig_WebrtcService_case:
		addressToCheck = service.GetWebrtcService().GetAddress()
	case configv1.UpstreamServiceConfig_McpService_case:
		mcp := service.GetMcpService()
		if mcp.WhichConnectionType() == configv1.McpUpstreamService_HttpConnection_case {
			addressToCheck = mcp.GetHttpConnection().GetHttpAddress()
		}
	case configv1.UpstreamServiceConfig_CommandLineService_case,
		configv1.UpstreamServiceConfig_SqlService_case:
		// Skip non-network or complex services
		return nil
	}

	if addressToCheck == "" {
		return nil
	}

	return checkAddressReachability(ctx, addressToCheck)
}

func checkAddressReachability(ctx context.Context, address string) error {
	host, port, err := parseHostPort(address)
	if err != nil {
		// If we can't parse it, maybe it's valid but we just don't understand it.
		// However, Validate() should have already caught invalid URLs.
		// So we return the error here to be strict.
		return fmt.Errorf("failed to parse address %q for connectivity check: %w", address, err)
	}

	target := net.JoinHostPort(host, port)
	timeout := 2 * time.Second

	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", target)
	if err != nil {
		return fmt.Errorf("connectivity check failed for %q: %w", target, err)
	}
	if conn != nil {
		_ = conn.Close()
	}

	return nil
}

func parseHostPort(address string) (string, string, error) {
	// Handle cases without scheme (common in gRPC)
	if !strings.Contains(address, "://") {
		// Assume it's host:port
		host, port, err := net.SplitHostPort(address)
		if err == nil {
			return host, port, nil
		}
		// If it fails, maybe it's just host (implies default port? No, net.Dial needs port)
		// Or maybe it has no port.
		// Let's try to parse as URL just in case by prepending dummy scheme
		// But gRPC often uses just "localhost:50051".
		return "", "", fmt.Errorf("invalid address format (expected host:port): %w", err)
	}

	u, err := url.Parse(address)
	if err != nil {
		return "", "", err
	}

	host := u.Hostname()
	port := u.Port()

	if port == "" {
		switch u.Scheme {
		case "http", "ws":
			port = "80"
		case "https", "wss":
			port = "443"
		default:
			return "", "", fmt.Errorf("unknown scheme %q, cannot determine default port", u.Scheme)
		}
	}

	return host, port, nil
}
