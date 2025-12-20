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

func TestManager_RegisterService(t *testing.T) {
	m := NewManager()
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
	}

	m.RegisterService("test-id", config)

	state, ok := m.GetState("test-id")
	assert.True(t, ok)
	assert.Equal(t, config, state.Config)
	assert.Equal(t, pb_admin.ServiceStatus_SERVICE_STATUS_UNKNOWN, state.GetStatus())

	m.UnregisterService("test-id")
	_, ok = m.GetState("test-id")
	assert.False(t, ok)
}

func TestManager_CheckService(t *testing.T) {
	m := NewManager()
	config := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		// Use stdio service which always returns healthy in default checker
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Command: proto.String("echo"),
			},
		},
	}

	m.RegisterService("test-id", config)

	// Wait for background check to run
	time.Sleep(100 * time.Millisecond)

	state, ok := m.GetState("test-id")
	assert.True(t, ok)
	// commandLineCheck returns nil error if no health check configured, which maps to Up -> Healthy
	assert.Equal(t, pb_admin.ServiceStatus_SERVICE_STATUS_HEALTHY, state.GetStatus())
}

func TestManager_StartStop(t *testing.T) {
	m := NewManager()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.Start(ctx)
	// Just ensure no panic/race
	time.Sleep(10 * time.Millisecond)
	m.Stop()
}
