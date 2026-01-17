// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ConnectivityError represents an error during a connectivity check.
type ConnectivityError struct {
	ServiceName string
	Err         error
}

func (e *ConnectivityError) Error() string {
	return fmt.Sprintf("service %q: %v", e.ServiceName, e.Err)
}

// CheckConnectivity performs active connectivity checks on the services defined in the configuration.
// It attempts to reach the upstream services to verify that the configuration is not only syntactically correct
// but also points to reachable targets.
func CheckConnectivity(ctx context.Context, config *configv1.McpAnyServerConfig) []ConnectivityError {
	var errors []ConnectivityError
	timeout := 5 * time.Second

	for _, service := range config.GetUpstreamServices() {
		if service.GetDisable() {
			continue
		}

		ctx, cancel := context.WithTimeout(ctx, timeout)
		var err error

		switch {
		case service.GetHttpService() != nil:
			err = checkHTTP(ctx, service.GetHttpService())
		case service.GetWebsocketService() != nil:
			err = checkWebSocket(ctx, service.GetWebsocketService())
		case service.GetGrpcService() != nil:
			err = checkGRPC(ctx, service.GetGrpcService())
		case service.GetSqlService() != nil:
			err = checkSQL(ctx, service.GetSqlService())
		case service.GetOpenapiService() != nil:
			err = checkOpenAPI(ctx, service.GetOpenapiService())
		case service.GetGraphqlService() != nil:
			err = checkGraphQL(ctx, service.GetGraphqlService())
		case service.GetWebrtcService() != nil:
			err = checkWebRTC(ctx, service.GetWebrtcService())
		}

		cancel()

		if err != nil {
			errors = append(errors, ConnectivityError{
				ServiceName: service.GetName(),
				Err:         err,
			})
		}
	}

	return errors
}

func checkHTTP(ctx context.Context, s *configv1.HttpUpstreamService) error {
	return checkURL(ctx, s.GetAddress())
}

func checkWebSocket(ctx context.Context, s *configv1.WebsocketUpstreamService) error {
	// WebSocket usually starts with HTTP handshake, so we can check the URL.
	// We might need to change scheme ws->http, wss->https for net/http client,
	// or use a dialer.
	u, err := url.Parse(s.GetAddress())
	if err != nil {
		return fmt.Errorf("invalid url: %w", err)
	}

	target := *u
	switch target.Scheme {
	case "ws":
		target.Scheme = "http"
	case "wss":
		target.Scheme = "https"
	}

	return checkURL(ctx, target.String())
}

func checkGRPC(ctx context.Context, s *configv1.GrpcUpstreamService) error {
	// gRPC often uses raw TCP, sometimes HTTP/2.
	// We can try to TCP dial the address.
	return checkTCP(ctx, s.GetAddress())
}

func checkOpenAPI(ctx context.Context, s *configv1.OpenapiUpstreamService) error {
	// OpenAPI service might have an address or spec_url.
	if s.GetAddress() != "" {
		if err := checkURL(ctx, s.GetAddress()); err != nil {
			return fmt.Errorf("address unreachable: %w", err)
		}
	}
	if s.GetSpecUrl() != "" {
		if err := checkURL(ctx, s.GetSpecUrl()); err != nil {
			return fmt.Errorf("spec_url unreachable: %w", err)
		}
	}
	return nil
}

func checkGraphQL(ctx context.Context, s *configv1.GraphQLUpstreamService) error {
	return checkURL(ctx, s.GetAddress())
}

func checkWebRTC(ctx context.Context, s *configv1.WebrtcUpstreamService) error {
	return checkURL(ctx, s.GetAddress())
}

func checkSQL(ctx context.Context, s *configv1.SqlUpstreamService) error {
	// We attempt to ping the database.
	// Note: This relies on the driver being registered in the binary that runs this code (e.g. main).
	db, err := sql.Open(s.GetDriver(), s.GetDsn())
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

func checkURL(ctx context.Context, address string) error {
	if address == "" {
		return nil
	}
	req, err := http.NewRequestWithContext(ctx, "HEAD", address, nil)
	if err != nil {
		return err
	}

	// Use a custom client to avoid redirects or long timeouts
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		// If HEAD fails (e.g. 405 Method Not Allowed), try GET
		req.Method = "GET"
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// We consider 5xx as "reachable but failing", which might be useful info,
	// but for "connectivity" check, if we got a response, it's connected.
	// However, connection refused or timeout is what we care about.
	return nil
}

func checkTCP(ctx context.Context, address string) error {
	if address == "" {
		return nil
	}

	// Handle scheme stripping if present
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err == nil {
			address = u.Host
		}
	}

	d := net.Dialer{}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}
