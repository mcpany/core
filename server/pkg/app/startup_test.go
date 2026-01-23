// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	buspkg "github.com/mcpany/core/server/pkg/bus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// manualBus implements bus.Bus manually to allow triggering events
type manualBus[T any] struct {
	subscribers []func(T)
	hold        bool // If true, Publish does not call subscribers
	pending     []T
}

func (m *manualBus[T]) Publish(ctx context.Context, topic string, msg T) error {
	if m.hold {
		m.pending = append(m.pending, msg)
		return nil
	}
	for _, sub := range m.subscribers {
		sub(msg)
	}
	return nil
}

func (m *manualBus[T]) Release() {
	m.hold = false
	for _, msg := range m.pending {
		for _, sub := range m.subscribers {
			sub(msg)
		}
	}
	m.pending = nil
}

func (m *manualBus[T]) Subscribe(ctx context.Context, topic string, handler func(T)) func() {
	m.subscribers = append(m.subscribers, handler)
	return func() {}
}

func (m *manualBus[T]) SubscribeOnce(ctx context.Context, topic string, handler func(T)) func() {
	return func() {}
}

func TestRun_WaitsForServiceRegistration(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configContent := `
upstream_services:
 - name: "test-service"
   http_service:
     address: "http://127.0.0.1:8080"
     tools:
       - name: "test-call"
         call_id: "test-call"
     calls:
        test-call:
          id: "test-call"
          endpoint_path: "/test"
          method: "HTTP_METHOD_POST"
`
	err := afero.WriteFile(fs, "/config.yaml", []byte(configContent), 0o644)
	require.NoError(t, err)

	reqBus := &manualBus[*buspkg.ServiceRegistrationRequest]{hold: true}
	resBus := &manualBus[*buspkg.ServiceRegistrationResult]{}

	// Mock bus.GetBus
	buspkg.GetBusHook = func(_ *buspkg.Provider, topic string) (any, error) {
		if topic == buspkg.ServiceRegistrationRequestTopic {
			return reqBus, nil
		}
		if topic == buspkg.ServiceRegistrationResultTopic {
			return resBus, nil
		}
		return nil, nil
	}
	defer func() { buspkg.GetBusHook = nil }()

	app := NewApplication()
	mockStore := new(MockStore)
	mockStore.On("Load", mock.Anything).Return((*configv1.McpAnyServerConfig)(nil), nil)
	mockStore.On("ListServices", mock.Anything).Return([]*configv1.UpstreamServiceConfig{}, nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return(&configv1.GlobalSettings{}, nil)
	mockStore.On("ListUsers", mock.Anything).Return([]*configv1.User{}, nil)
	mockStore.On("Close").Return(nil)
	app.Storage = mockStore

	// Run application in background
	go func() {
		_ = app.Run(RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     "127.0.0.1:0",
			GRPCPort:        "127.0.0.1:0",
			ConfigPaths:     []string{"/config.yaml"},
			APIKey:          "",
			ShutdownTimeout: 1 * time.Second,
		})
	}()

	// 1. Verify that startup BLOCKS (WaitForStartup times out)
	startupCtx, startupCancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer startupCancel()
	err = app.WaitForStartup(startupCtx)
	assert.Error(t, err, "WaitForStartup should timeout/error because Run is waiting for registration")
	assert.Equal(t, context.DeadlineExceeded, err)

	// 2. Release request to unblock worker (which will publish result)
	reqBus.Release()

	// 3. Verify startup COMPLETES
	startupCtx2, startupCancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer startupCancel2()
	err = app.WaitForStartup(startupCtx2)
	assert.NoError(t, err, "WaitForStartup should succeed after registration result is received")
}
