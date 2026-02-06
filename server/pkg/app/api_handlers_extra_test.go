package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TestHandleListServices_ToolCountInjection tests the tool count injection logic in handleListServices.
// It covers the lines where tool count is calculated and injected into the response.
func TestHandleListServices_ToolCountInjection(t *testing.T) {
	app, store := setupApiTestApp()

	// Mock ServiceRegistry to return success
	mockRegistry := new(MockServiceRegistry)
	app.ServiceRegistry = mockRegistry

	// Add services to store (fallback) or registry
	svc1 := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("svc1"),
		Name: proto.String("Service 1"),
	}.Build()
	svc2 := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String("svc2"),
		Name: proto.String("Service 2"),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svc1, svc2}, nil)
	mockRegistry.On("GetServiceError", mock.Anything).Return("", false)

	// Add tools to ToolManager using TestMockTool (defined in api_test.go)
	app.ToolManager.AddTool(&TestMockTool{
		toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool1"), ServiceId: proto.String("svc1")}.Build(),
	})
	app.ToolManager.AddTool(&TestMockTool{
		toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool2"), ServiceId: proto.String("svc1")}.Build(),
	})
	app.ToolManager.AddTool(&TestMockTool{
		toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool3"), ServiceId: proto.String("svc2")}.Build(),
	})

	handler := app.handleServices(store)
	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var services []map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &services)
	require.NoError(t, err)

	require.Len(t, services, 2)

	for _, s := range services {
		id := s["id"].(string)
		if id == "svc1" {
			assert.Equal(t, float64(2), s["tool_count"])
		} else if id == "svc2" {
			assert.Equal(t, float64(1), s["tool_count"])
		}
	}
}

func TestHandleCreateService_Validation(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleServices(store)

	// Invalid URL
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("invalid-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("invalid-url"),
		}.Build(),
	}.Build()
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid service configuration")
}

func TestHandleCreateService_Unsafe_NonAdmin(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleServices(store)

	// Unsafe config (Filesystem)
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("unsafe-fs"),
		FilesystemService: configv1.FilesystemUpstreamService_builder{
			RootPaths: map[string]string{"/": "/"},
			Os:        configv1.OsFs_builder{}.Build(),
		}.Build(),
	}.Build()
	body, _ := protojson.Marshal(svc)

	// Request without admin role
	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	// No context with admin role
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestHandleServiceValidate_Connectivity(t *testing.T) {
	// Enable loopback for reachability checks
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")

	app := &Application{}
	handler := app.handleServiceValidate()

	t.Run("Valid HTTP", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer ts.Close()

		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("valid-http"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String(ts.URL),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["valid"].(bool))
	})

	t.Run("Invalid HTTP", func(t *testing.T) {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("invalid-http"),
			HttpService: configv1.HttpUpstreamService_builder{
				Address: proto.String("http://127.0.0.1:0"),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["valid"].(bool))
		assert.Contains(t, resp["details"], "reachability check failed")
	})

	t.Run("Valid Filesystem", func(t *testing.T) {
		tmpDir := t.TempDir()
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("valid-fs"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				RootPaths: map[string]string{"root": tmpDir},
				Os:        configv1.OsFs_builder{}.Build(),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["valid"].(bool))
	})

	t.Run("Invalid Filesystem", func(t *testing.T) {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("invalid-fs"),
			FilesystemService: configv1.FilesystemUpstreamService_builder{
				RootPaths: map[string]string{"root": "/non/existent/path"},
				Os:        configv1.OsFs_builder{}.Build(),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["valid"].(bool))
		assert.Contains(t, resp["details"], "Filesystem path check failed")
	})

	t.Run("Valid Command", func(t *testing.T) {
		// Use "sh" or "ls" which should exist
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("valid-cmd"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Command: proto.String("sh"),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.True(t, resp["valid"].(bool))
	})

	t.Run("Invalid Command", func(t *testing.T) {
		svc := configv1.UpstreamServiceConfig_builder{
			Name: proto.String("invalid-cmd"),
			CommandLineService: configv1.CommandLineUpstreamService_builder{
				Command: proto.String("non-existent-command-xyz"),
			}.Build(),
		}.Build()
		body, _ := protojson.Marshal(svc)
		req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		// Static validation catches non-existent commands
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]any
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.False(t, resp["valid"].(bool))
		assert.Contains(t, resp["details"], "Static validation failed")
	})
}

func TestHandleServiceDetail_Update_Unsafe_NonAdmin(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleServiceDetail(store)

	// Create initial safe service
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("my-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://example.com"),
		}.Build(),
	}.Build()
	store.SaveService(context.Background(), svc)

	// Update to unsafe service
	svc.SetCommandLineService(configv1.CommandLineUpstreamService_builder{
		Command: proto.String("sh"),
	}.Build())
	body, _ := protojson.Marshal(svc)

	req := httptest.NewRequest(http.MethodPut, "/services/my-service", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// MockStorage for failing SaveService
type MockStorage struct {
	storage.Storage // Embed interface to fallback
	failSave bool
}

func (m *MockStorage) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	if m.failSave {
		return errors.New("save failed")
	}
	return m.Storage.SaveService(ctx, service)
}

func (m *MockStorage) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	if m.failSave {
		return errors.New("save failed")
	}
	return m.Storage.SaveSecret(ctx, secret)
}

func TestHandleCreateService_SaveError(t *testing.T) {
	app, store := setupApiTestApp()
	mockStore := &MockStorage{Storage: store, failSave: true}
	handler := app.handleServices(mockStore)

	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080"),
		}.Build(),
	}.Build()
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleSecrets_SaveError(t *testing.T) {
	app, store := setupApiTestApp()
	mockStore := &MockStorage{Storage: store, failSave: true}
	handler := app.handleSecrets(mockStore)

	secret := configv1.Secret_builder{
		Name:  proto.String("my-secret"),
		Id:    proto.String("my-secret-id"),
		Value: proto.String("super-secret"),
	}.Build()
	body, _ := protojson.Marshal(secret)
	req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleServiceRestart_Error(t *testing.T) {
	// To test reload failure, we can corrupt the config file in the fs.
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	app.configPaths = []string{"config.yaml"}
	store := memory.NewStore()
	app.Storage = store // Ensure app knows about the store if used internally for reload

	// Initialize config file
	afero.WriteFile(fs, "config.yaml", []byte("upstream_services: []"), 0644)
	app.ReloadConfig(context.Background(), fs, []string{"config.yaml"})

	// Create service in store so we can restart it
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("svc1"),
	}.Build()
	store.SaveService(context.Background(), svc)

	// Now corrupt the config file so ReloadConfig fails
	afero.WriteFile(fs, "config.yaml", []byte("invalid_yaml: ["), 0644)

	handler := app.handleServiceDetail(store)
	req := httptest.NewRequest(http.MethodPost, "/services/svc1/restart", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Failed to restart service")
}
