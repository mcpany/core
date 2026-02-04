// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// SpyServiceRegistry captures calls
type SpyServiceRegistry struct {
	TestMockServiceRegistry
	unregistered []string
}

func (m *SpyServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	m.unregistered = append(m.unregistered, serviceName)
	return nil
}

type MockStoreWithGet struct {
	MockServiceStore
	service *configv1.UpstreamServiceConfig
}

func (m *MockStoreWithGet) GetDashboardLayout(ctx context.Context, userID string) (string, error) {
	return "", nil
}

func (m *MockStoreWithGet) SaveDashboardLayout(ctx context.Context, userID string, layoutJSON string) error {
	return nil
}

func (m *MockStoreWithGet) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	if m.service != nil && m.service.GetName() == name {
		return m.service, nil
	}
	return nil, nil
}

func TestHandleServiceRestart_Success(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.configPaths = []string{}

	registry := &SpyServiceRegistry{}
	app.ServiceRegistry = registry

	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	store := &MockStoreWithGet{
		service: svc,
	}

	req := httptest.NewRequest(http.MethodPost, "/services/test-service/restart", nil)
	rr := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.handleServiceRestart(w, r, "test-service", store)
	})

	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	// Verify UnregisterService was called
	require.Len(t, registry.unregistered, 1)
	assert.Equal(t, "test-service", registry.unregistered[0])
}
