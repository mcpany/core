/*
 * Copyright 2024 Author(s) of MCP Any
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"github.com/mcpany/core/proto/config/v1"
)

// Checker defines the interface for a health checker.
type Checker interface {
	Start()
	Stop()
}

// NewChecker creates a new health checker based on the provided configuration.
func NewChecker(serviceName string, healthCheckConfig *config.HealthCheck, serviceRegistry HealthCheckStatusUpdater) (Checker, error) {
	if healthCheckConfig == nil {
		return nil, fmt.Errorf("health check config cannot be nil")
	}

	interval := time.Duration(healthCheckConfig.GetIntervalSeconds()) * time.Second
	if interval == 0 {
		interval = 15 * time.Second
	}

	switch c := healthCheckConfig.GetCheck().(type) {
	case *config.HealthCheck_HttpHealthCheck:
		return NewHTTPChecker(serviceName, c.HttpHealthCheck, interval, serviceRegistry), nil
	case *config.HealthCheck_GrpcHealthCheck:
		return NewGRPCChecker(serviceName, c.GrpcHealthCheck, interval, serviceRegistry), nil
	default:
		return nil, fmt.Errorf("unknown health check type: %T", c)
	}
}

// HealthCheckStatusUpdater defines the interface for updating the health check
// status of a service.
type HealthCheckStatusUpdater interface {
	UpdateHealthCheckStatus(serviceID string, isHealthy bool)
}

type httpChecker struct {
	serviceName     string
	config          *config.HTTPHealthCheck
	interval        time.Duration
	serviceRegistry HealthCheckStatusUpdater
	stopChan        chan struct{}
}

// NewHTTPChecker creates a new HTTP health checker.
func NewHTTPChecker(serviceName string, config *config.HTTPHealthCheck, interval time.Duration, serviceRegistry HealthCheckStatusUpdater) *httpChecker {
	return &httpChecker{
		serviceName:     serviceName,
		config:          config,
		interval:        interval,
		serviceRegistry: serviceRegistry,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the health checking process for the HTTP service.
func (c *httpChecker) Start() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.check()
		case <-c.stopChan:
			return
		}
	}
}

// Stop terminates the health checking process.
func (c *httpChecker) Stop() {
	close(c.stopChan)
}

func (c *httpChecker) check() {
	resp, err := http.Get(c.config.GetAddress())
	if err != nil {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, false)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, true)
	} else {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, false)
	}
}

type grpcChecker struct {
	serviceName     string
	config          *config.GRPCHealthCheck
	interval        time.Duration
	serviceRegistry HealthCheckStatusUpdater
	stopChan        chan struct{}
	conn            *grpc.ClientConn
}

// NewGRPCChecker creates a new gRPC health checker.
func NewGRPCChecker(serviceName string, config *config.GRPCHealthCheck, interval time.Duration, serviceRegistry HealthCheckStatusUpdater) *grpcChecker {
	return &grpcChecker{
		serviceName:     serviceName,
		config:          config,
		interval:        interval,
		serviceRegistry: serviceRegistry,
		stopChan:        make(chan struct{}),
	}
}

// Start begins the health checking process for the gRPC service.
func (c *grpcChecker) Start() {
	var err error
	c.conn, err = grpc.Dial(c.config.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, false)
		return
	}

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.check()
		case <-c.stopChan:
			return
		}
	}
}

// Stop terminates the health checking process.
func (c *grpcChecker) Stop() {
	if c.conn != nil {
		c.conn.Close()
	}
	close(c.stopChan)
}

func (c *grpcChecker) check() {
	if c.conn == nil {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, false)
		return
	}

	client := healthpb.NewHealthClient(c.conn)
	resp, err := client.Check(context.Background(), &healthpb.HealthCheckRequest{})
	if err != nil {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, false)
		return
	}

	if resp.GetStatus() == healthpb.HealthCheckResponse_SERVING {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, true)
	} else {
		c.serviceRegistry.UpdateHealthCheckStatus(c.serviceName, false)
	}
}
