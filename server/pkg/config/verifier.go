// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"google.golang.org/protobuf/proto"
)

// VerifyServices performs connectivity checks on the configured services.
// It updates the ConfigError field of any service that fails the check.
func VerifyServices(ctx context.Context, config *configv1.McpAnyServerConfig) {
	// Use SafeHTTPClient to respect global security settings
	client := util.NewSafeHTTPClient()
	// Set a short timeout for verification to avoid blocking startup significantly
	client.Timeout = 2 * time.Second

	for _, service := range config.GetUpstreamServices() {
		// Skip disabled services
		if service.GetDisable() {
			continue
		}

		// Skip if already has config error
		if service.GetConfigError() != "" {
			continue
		}

		var err error
		switch {
		case service.GetHttpService() != nil:
			err = verifyURL(ctx, client, service.GetHttpService().GetAddress())
		case service.GetGraphqlService() != nil:
			err = verifyURL(ctx, client, service.GetGraphqlService().GetAddress())
		case service.GetOpenapiService() != nil:
			if addr := service.GetOpenapiService().GetAddress(); addr != "" {
				err = verifyURL(ctx, client, addr)
			}
		case service.GetMcpService() != nil:
			if httpConn := service.GetMcpService().GetHttpConnection(); httpConn != nil {
				err = verifyURL(ctx, client, httpConn.GetHttpAddress())
			}
		case service.GetGrpcService() != nil:
			err = verifyTCP(ctx, service.GetGrpcService().GetAddress())
		case service.GetWebsocketService() != nil:
			// WebSocket URLs are usually ws:// or wss://
			err = verifyWebSocket(ctx, service.GetWebsocketService().GetAddress())
		case service.GetWebrtcService() != nil:
			// WebRTC signaling is usually HTTP
			err = verifyURL(ctx, client, service.GetWebrtcService().GetAddress())
		}

		if err != nil {
			service.ConfigError = proto.String(fmt.Sprintf("Connectivity check failed: %v", err))
		}
	}
}

func verifyURL(ctx context.Context, client *http.Client, address string) error {
	// Parse URL to ensure it's valid before requesting
	u, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	if u.Scheme == "" {
		return fmt.Errorf("missing scheme in url")
	}

	// Basic Head request
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, address, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func verifyTCP(ctx context.Context, address string) error {
	// We use SafeDialContext which uses the default SafeDialer (strict mode).
	// To respect environment variables (like MCPANY_ALLOW_LOOPBACK_RESOURCES),
	// we need to construct a SafeDialer similar to how NewSafeHTTPClient does.
	// This ensures our verification logic matches the runtime security policies.
	dialer := util.NewSafeDialer()
	// Note: We are relying on NewSafeHTTPClient logic pattern here.
	// Ideally util package should expose a NewConfiguredSafeDialer().
	// For now, we manually check the same env vars as NewSafeHTTPClient.
	if isEnvTrue("MCPANY_ALLOW_LOOPBACK_RESOURCES") {
		dialer.AllowLoopback = true
	}
	if isEnvTrue("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") {
		dialer.AllowPrivate = true
	}

	// Add timeout logic by wrapping context
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	conn.Close()
	return nil
}

func isEnvTrue(key string) bool {
	return os.Getenv(key) == util.TrueStr
}

func verifyWebSocket(ctx context.Context, address string) error {
	u, err := url.Parse(address)
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}
	port := u.Port()
	if port == "" {
		if u.Scheme == "wss" {
			port = "443"
		} else {
			port = "80"
		}
	}
	host := u.Hostname()
	return verifyTCP(ctx, net.JoinHostPort(host, port))
}
