// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"testing"
	"time"

	pb_admin "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager(t *testing.T) {
	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Start(ctx)
	defer m.Stop()

	// Register a service that should fail connection
	cfg := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:0"), // Should fail
			},
		},
	}

	m.RegisterService("test-service", cfg)

	// Allow some time for background check
	assert.Eventually(t, func() bool {
		status := m.GetStatus("test-service")
		return status != nil && status.Status == pb_admin.ServiceStatus_SERVICE_STATUS_UNHEALTHY
	}, 2*time.Second, 100*time.Millisecond)

	status := m.GetStatus("test-service")
	assert.Contains(t, status.LastError, "failed to connect")

	m.UnregisterService("test-service")
	assert.Nil(t, m.GetStatus("test-service"))
}
