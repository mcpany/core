// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package connectivity provides tools to verify network connectivity to upstream services.
package connectivity

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

	// Import SQL drivers.
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

// Result represents the outcome of a connectivity check.
type Result struct {
	ServiceName string
	Type        string
	Target      string
	Status      bool
	Latency     time.Duration
	Error       error
}

// Check performs connectivity checks for all services in the configuration.
func Check(ctx context.Context, cfg *configv1.McpAnyServerConfig) []Result {
	services := cfg.GetUpstreamServices()
	results := make([]Result, 0, len(services))

	for _, service := range services {
		results = append(results, checkService(ctx, service))
	}

	return results
}

func checkService(ctx context.Context, service *configv1.UpstreamServiceConfig) Result {
	res := Result{
		ServiceName: service.GetName(),
		Status:      true,
	}

	start := time.Now()
	defer func() {
		res.Latency = time.Since(start)
	}()

	var err error
	var target string
	var svcType string

	switch {
	case service.GetHttpService() != nil:
		svcType = "HTTP"
		target = service.GetHttpService().GetAddress()
		err = checkHTTP(ctx, target)
	case service.GetWebsocketService() != nil:
		svcType = "WebSocket"
		target = service.GetWebsocketService().GetAddress()
		err = checkTCP(ctx, target) // WebSocket uses TCP
	case service.GetGrpcService() != nil:
		svcType = "gRPC"
		target = service.GetGrpcService().GetAddress()
		err = checkTCP(ctx, target)
	case service.GetOpenapiService() != nil:
		svcType = "OpenAPI"
		s := service.GetOpenapiService()
		switch {
		case s.GetAddress() != "":
			target = s.GetAddress()
			err = checkHTTP(ctx, target)
		case s.GetSpecUrl() != "":
			target = s.GetSpecUrl()
			err = checkHTTP(ctx, target)
		default:
			// Spec content, nothing to check network-wise
			target = "local-spec"
		}
	case service.GetSqlService() != nil:
		svcType = "SQL"
		s := service.GetSqlService()
		target = fmt.Sprintf("%s://%s", s.GetDriver(), "redacted")
		err = checkSQL(ctx, s.GetDriver(), s.GetDsn())
	case service.GetGraphqlService() != nil:
		svcType = "GraphQL"
		target = service.GetGraphqlService().GetAddress()
		err = checkHTTP(ctx, target)
	case service.GetMcpService() != nil:
		svcType = "MCP"
		m := service.GetMcpService()
		if m.GetHttpConnection() != nil {
			target = m.GetHttpConnection().GetHttpAddress()
			err = checkHTTP(ctx, target)
		} else {
			target = "stdio/bundle"
			// Stdio/Bundle already validated by static validator
		}
	case service.GetWebrtcService() != nil:
		svcType = "WebRTC"
		target = service.GetWebrtcService().GetAddress()
		err = checkHTTP(ctx, target) // Usually signaled over HTTP/WS
	case service.GetCommandLineService() != nil:
		svcType = "CMD"
		target = service.GetCommandLineService().GetCommand()
		// Validated by static validator
	default:
		svcType = "Unknown"
		target = "unknown"
	}

	res.Type = svcType
	res.Target = target
	if err != nil {
		res.Status = false
		res.Error = err
	}

	return res
}

func checkTCP(ctx context.Context, address string) error {
	// Handle scheme if present
	if strings.Contains(address, "://") {
		u, err := url.Parse(address)
		if err != nil {
			return err
		}
		address = u.Host
		switch u.Scheme {
		case "https", "wss":
			// If default ports are missing
			if !strings.Contains(address, ":") {
				address += ":443"
			}
		case "http", "ws":
			if !strings.Contains(address, ":") {
				address += ":80"
			}
		}
	}

	d := net.Dialer{Timeout: 5 * time.Second}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

func checkHTTP(ctx context.Context, address string) error {
	client := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "HEAD", address, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		// Fallback to GET if HEAD fails (some servers block HEAD)
		req, err = http.NewRequestWithContext(ctx, "GET", address, nil)
		if err != nil {
			return err
		}
		resp, err = client.Do(req)
		if err != nil {
			return err
		}
	}
	_ = resp.Body.Close()
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}
	return nil
}

func checkSQL(ctx context.Context, driver, dsn string) error {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return db.PingContext(ctx)
}
