// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	// Import SQL drivers.
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

// DeepValidator handles "deep" validation checks (connectivity, auth).
type DeepValidator struct {
	Timeout time.Duration
}

// NewDeepValidator creates a new DeepValidator.
func NewDeepValidator(timeout time.Duration) *DeepValidator {
	return &DeepValidator{
		Timeout: timeout,
	}
}

// Validate performs deep validation on the configuration.
func (d *DeepValidator) Validate(ctx context.Context, config *configv1.McpAnyServerConfig) []ValidationError {
	var errors []ValidationError
	var mu sync.Mutex
	var wg sync.WaitGroup

	services := config.GetUpstreamServices()
	for _, service := range services {
		wg.Add(1)
		go func(svc *configv1.UpstreamServiceConfig) {
			defer wg.Done()
			if err := d.validateService(ctx, svc); err != nil {
				mu.Lock()
				errors = append(errors, ValidationError{
					ServiceName: svc.GetName(),
					Err:         err,
				})
				mu.Unlock()
			}
		}(service)
	}

	wg.Wait()
	return errors
}

func (d *DeepValidator) validateService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	// Create a child context with timeout for each service check
	ctx, cancel := context.WithTimeout(ctx, d.Timeout)
	defer cancel()

	if httpService := service.GetHttpService(); httpService != nil {
		return d.validateHTTP(ctx, httpService)
	} else if grpcService := service.GetGrpcService(); grpcService != nil {
		return d.validateGRPC(ctx, grpcService)
	} else if sqlService := service.GetSqlService(); sqlService != nil {
		return d.validateSQL(ctx, sqlService)
	}
	// Add other service types as needed (e.g. mcp http connection)

	return nil
}

func (d *DeepValidator) validateHTTP(ctx context.Context, s *configv1.HttpUpstreamService) error {
	req, err := http.NewRequestWithContext(ctx, "HEAD", s.GetAddress(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: d.Timeout}

	// Configure TLS if specified
	if tlsConfig := s.GetTlsConfig(); tlsConfig != nil {
		tlsConf, err := d.buildTLSConfig(tlsConfig)
		if err != nil {
			return fmt.Errorf("invalid TLS config: %w", err)
		}
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConf,
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		// If HEAD fails, try GET (some servers don't support HEAD)
		req.Method = "GET"
		resp, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to connect to %s: %w", s.GetAddress(), err)
		}
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("server returned server error: %d", resp.StatusCode)
	}

	return nil
}

func (d *DeepValidator) validateGRPC(ctx context.Context, s *configv1.GrpcUpstreamService) error {
	opts := []grpc.DialOption{grpc.WithBlock()} //nolint:staticcheck // Keep blocking dial for validation

	if tlsConfig := s.GetTlsConfig(); tlsConfig != nil {
		tlsConf, err := d.buildTLSConfig(tlsConfig)
		if err != nil {
			return fmt.Errorf("invalid TLS config: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConf)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Try to dial the gRPC server
	conn, err := grpc.DialContext(ctx, s.GetAddress(), opts...) //nolint:staticcheck // DialContext is deprecated but needed for blocking
	if err != nil {
		return fmt.Errorf("failed to dial gRPC server %s: %w", s.GetAddress(), err)
	}
	defer func() {
		_ = conn.Close()
	}()
	return nil
}

func (d *DeepValidator) validateSQL(ctx context.Context, s *configv1.SqlUpstreamService) error {
	db, err := sql.Open(s.GetDriver(), s.GetDsn())
	if err != nil {
		return fmt.Errorf("failed to open database connection: %w", err)
	}
	defer func() {
		_ = db.Close()
	}()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}
	return nil
}

func (d *DeepValidator) buildTLSConfig(cfg *configv1.TLSConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.GetInsecureSkipVerify(), //nolint:gosec // Intentionally allowing insecure skip verify if configured by user
		ServerName:         cfg.GetServerName(),
	}

	if cfg.GetCaCertPath() != "" {
		caCert, err := os.ReadFile(cfg.GetCaCertPath())
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	if cfg.GetClientCertPath() != "" && cfg.GetClientKeyPath() != "" {
		cert, err := tls.LoadX509KeyPair(cfg.GetClientCertPath(), cfg.GetClientKeyPath())
		if err != nil {
			return nil, fmt.Errorf("failed to load client cert/key: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	return tlsConfig, nil
}
