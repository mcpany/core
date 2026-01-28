// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"context"
	"testing"

	config "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

type mockProvider struct {
	services []*config.UpstreamServiceConfig
}

func (m *mockProvider) GetAllServices() ([]*config.UpstreamServiceConfig, error) {
	return m.services, nil
}

func TestRecorder(t *testing.T) {
	svc1 := config.UpstreamServiceConfig_builder{
		Name: proto.String("service-1"),
	}.Build()
	svc2 := config.UpstreamServiceConfig_builder{
		Name:      proto.String("service-2"),
		LastError: proto.String("failed"),
	}.Build()

	services := []*config.UpstreamServiceConfig{svc1, svc2}
	provider := &mockProvider{services: services}
	recorder := NewRecorder(provider)

	// Test recording
	recorder.record(context.Background())

	history := recorder.GetHistory()
	assert.Len(t, history, 1)
	assert.Equal(t, int32(2), history[0].TotalServices)
	assert.Equal(t, int32(1), history[0].HealthyServices)
	assert.Equal(t, 50.0, history[0].UptimePercentage)

	// Test max points
	recorder.maxPoints = 2
	recorder.record(context.Background())
	recorder.record(context.Background()) // Should trigger shift

	history = recorder.GetHistory()
	assert.Len(t, history, 2)
}
