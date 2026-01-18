package app

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// MockStoreError implements ServiceStore but returns errors
type MockStoreError struct {
	MockServiceStore
}

func (s *MockStoreError) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return errors.New("save failed")
}
func (s *MockStoreError) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return nil, errors.New("list failed")
}

func TestCreateService_Success(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	mockStore := &MockServiceStore{}

	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("new-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
	}
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := app.handleServices(mockStore)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
}

func TestCreateService_InvalidConfig(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	mockStore := &MockServiceStore{}

	// Missing Name
	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
	}
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := app.handleServices(mockStore)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateService_StoreError(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	mockStore := &MockStoreError{}

	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("new-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
	}
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	rr := httptest.NewRecorder()

	handler := app.handleServices(mockStore)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestListServices_Error(t *testing.T) {
	app := NewApplication()
	// No registry, so it falls back to store
	mockStore := &MockStoreError{}

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rr := httptest.NewRecorder()

	handler := app.handleServices(mockStore)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestListServices_Success_Fallback(t *testing.T) {
	app := NewApplication()
	// No registry, fallback to store
	mockStore := &MockServiceStore{
		services: []*configv1.UpstreamServiceConfig{
			{Name: proto.String("svc1")},
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rr := httptest.NewRecorder()

	handler := app.handleServices(mockStore)
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	require.Contains(t, rr.Body.String(), "svc1")
}
