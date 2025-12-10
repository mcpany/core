// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"fmt"
	"time"

	"github.com/mcpany/core/pkg/upstream"
	"github.com/mcpany/core/proto/config/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// GrpcChecker implements the Checkable interface for gRPC services.
type GrpcChecker struct {
	serviceID   string
	config      *config.GrpcHealthCheck
	upstreamSvc upstream.Upstream
}

// NewGrpcChecker creates a new GrpcChecker.
func NewGrpcChecker(serviceID string, cfg *config.GrpcHealthCheck, upstreamSvc upstream.Upstream) (*GrpcChecker, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gRPC health check config is nil for service %s", serviceID)
	}
	if upstreamSvc.Address() == "" {
		return nil, fmt.Errorf("gRPC health check address is not specified for service %s", serviceID)
	}

	return &GrpcChecker{
		serviceID:   serviceID,
		config:      cfg,
		upstreamSvc: upstreamSvc,
	}, nil
}

// ID returns the unique identifier of the service.
func (c *GrpcChecker) ID() string {
	return c.serviceID
}

// Interval returns the duration between health checks.
func (c *GrpcChecker) Interval() time.Duration {
	if c.config.Interval != nil {
		return c.config.Interval.AsDuration()
	}
	// Return a default interval if not specified.
	return 15 * time.Second
}

// HealthCheck performs the health check for the gRPC service.
// It uses the standard gRPC health checking protocol.
func (c *GrpcChecker) HealthCheck(ctx context.Context) error {
	dialOpts := []grpc.DialOption{
		grpc.WithBlock(), // Block until the connection is established.
	}
	if c.config.Insecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		// TODO: Implement TLS from upstreamSvc.TLSConfig()
		// For now, defaulting to insecure if TLS is not explicitly handled.
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	conn, err := grpc.DialContext(ctx, c.upstreamSvc.Address(), dialOpts...)
	if err != nil {
		return fmt.Errorf("failed to dial gRPC service %s for health check: %w", c.serviceID, err)
	}
	defer conn.Close()

	healthClient := grpc_health_v1.NewHealthClient(conn)
	req := &grpc_health_v1.HealthCheckRequest{
		Service: c.config.Service,
	}

	resp, err := healthClient.Check(ctx, req)
	if err != nil {
		return fmt.Errorf("gRPC health check RPC failed for %s: %w", c.serviceID, err)
	}

	if resp.GetStatus() != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("gRPC service %s is not in serving state: got %s", c.serviceID, resp.GetStatus())
	}

	return nil
}
