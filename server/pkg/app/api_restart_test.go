// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"errors"
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
	unregistered  []string
	unregisterErr error
}

func (m *SpyServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	m.unregistered = append(m.unregistered, serviceName)
	return m.unregisterErr
}

type MockStoreWithGet struct {
	MockServiceStore
	service *configv1.UpstreamServiceConfig
	getErr  error
}

func (m *MockStoreWithGet) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if m.service != nil && m.service.GetName() == name {
		return m.service, nil
	}
	return nil, nil
}

type MockStoreWithLoadError struct {
	MockStoreWithGet
}

func (m *MockStoreWithLoadError) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	return nil, errors.New("load error")
}

func TestHandleServiceRestart_Success(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.configPaths = []string{}

	registry := &SpyServiceRegistry{}
	app.ServiceRegistry = registry

	store := &MockStoreWithGet{
		service: &configv1.UpstreamServiceConfig{Name: proto.String("test-service")},
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

func TestHandleServiceRestart_MethodNotAllowed(t *testing.T) {
	app := NewApplication()
	req := httptest.NewRequest(http.MethodGet, "/services/test/restart", nil)
	w := httptest.NewRecorder()
	app.handleServiceRestart(w, req, "test", &MockStoreWithGet{})
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleServiceRestart_StorageError(t *testing.T) {
	app := NewApplication()
	store := &MockStoreWithGet{getErr: errors.New("db error")}
	req := httptest.NewRequest(http.MethodPost, "/services/test/restart", nil)
	w := httptest.NewRecorder()
	app.handleServiceRestart(w, req, "test", store)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleServiceRestart_NotFound(t *testing.T) {
	app := NewApplication()
	store := &MockStoreWithGet{service: nil}
	req := httptest.NewRequest(http.MethodPost, "/services/test/restart", nil)
	w := httptest.NewRecorder()
	app.handleServiceRestart(w, req, "test", store)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandleServiceRestart_UnregisterError(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.configPaths = []string{} // ensure load succeeds

	registry := &SpyServiceRegistry{unregisterErr: errors.New("failed")}
	app.ServiceRegistry = registry

	store := &MockStoreWithGet{
		service: &configv1.UpstreamServiceConfig{Name: proto.String("test-service")},
	}
	app.Storage = store

	req := httptest.NewRequest(http.MethodPost, "/services/test-service/restart", nil)
	rr := httptest.NewRecorder()

	app.handleServiceRestart(rr, req, "test-service", store)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Len(t, registry.unregistered, 1)
}

func TestHandleServiceRestart_ReloadError(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.configPaths = []string{}

	registry := &SpyServiceRegistry{}
	app.ServiceRegistry = registry

	store := &MockStoreWithLoadError{
		MockStoreWithGet: MockStoreWithGet{
			service: &configv1.UpstreamServiceConfig{Name: proto.String("test-service")},
		},
	}
	app.Storage = store // Inject into app for ReloadConfig

	req := httptest.NewRequest(http.MethodPost, "/services/test-service/restart", nil)
	rr := httptest.NewRecorder()

	app.handleServiceRestart(rr, req, "test-service", store)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "load error")
}
