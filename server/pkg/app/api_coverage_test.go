package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func setupCoverageTestApp() (*Application, storage.Storage) {
	bp, _ := bus.NewProvider(nil)
	app := &Application{
		PromptManager:   prompt.NewManager(),
		ToolManager:     tool.NewManager(bp),
		ResourceManager: resource.NewManager(),
		busProvider:     bp,
	}
	store := memory.NewStore()
	return app, store
}

func TestHandleServices(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleServices(store)

	// Test POST Success
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
		Id: proto.String(uuid.New().String()),
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d: %s", w.Code, w.Body.String())
	}

	// Test GET
	req = httptest.NewRequest(http.MethodGet, "/services", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test POST Missing Name
	svc.Name = proto.String("")
	body, _ = opts.Marshal(svc)
	req = httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", w.Code)
	}
}

func TestHandleServiceDetail(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleServiceDetail(store)

	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
		Id: proto.String(uuid.New().String()),
	}
	_ = store.SaveService(context.Background(), svc)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/services/test-service", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test GET Not Found
	req = httptest.NewRequest(http.MethodGet, "/services/non-existent", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", w.Code)
	}

	// Test PUT
	svc.ServiceConfig = &configv1.UpstreamServiceConfig_HttpService{
		HttpService: &configv1.HttpUpstreamService{
			Address: proto.String("http://updated:8080"),
		},
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(svc)
	req = httptest.NewRequest(http.MethodPut, "/services/test-service", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test DELETE
	req = httptest.NewRequest(http.MethodDelete, "/services/test-service", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 No Content, got %d", w.Code)
	}
}

func TestHandleServiceStatus(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleServiceDetail(store)

	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: proto.String("http://localhost:8080"),
			},
		},
	}
	_ = store.SaveService(context.Background(), svc)

	req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleSettings(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleSettings(store)

	// Test POST
	settings := &configv1.GlobalSettings{
		AllowedIps: []string{"127.0.0.1"},
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(settings)
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test GET
	req = httptest.NewRequest(http.MethodGet, "/settings", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleTools(t *testing.T) {
	app, _ := setupCoverageTestApp()
	handler := app.handleTools()

	// Add a dummy tool
	// Requires mocking or registering a tool. ToolManager.RegisterTool
	// But RegisterTool requires a serviceID.
	// Let's just check empty list response.
	req := httptest.NewRequest(http.MethodGet, "/tools", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandlePrompts(t *testing.T) {
	app, _ := setupCoverageTestApp()
	handler := app.handlePrompts()

	req := httptest.NewRequest(http.MethodGet, "/prompts", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleResources(t *testing.T) {
	app, _ := setupCoverageTestApp()
	handler := app.handleResources()

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleSecrets(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleSecrets(store)

	// Test POST
	secret := &configv1.Secret{
		Name:  proto.String("my-secret"),
		Value: proto.String("super-secret"),
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(secret)
	req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test GET
	req = httptest.NewRequest(http.MethodGet, "/secrets", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	// Verify redacted
	if bytes.Contains(w.Body.Bytes(), []byte("super-secret")) {
		t.Error("Secret value should be redacted")
	}
}

func TestHandleSecretDetail(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleSecretDetail(store)

	secret := &configv1.Secret{
		Id:    proto.String("sec-123"),
		Name:  proto.String("my-secret"),
		Value: proto.String("super-secret"),
	}
	_ = store.SaveSecret(context.Background(), secret)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/secrets/sec-123", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	if bytes.Contains(w.Body.Bytes(), []byte("super-secret")) {
		t.Error("Secret value should be redacted")
	}

	// Test DELETE
	req = httptest.NewRequest(http.MethodDelete, "/secrets/sec-123", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 No Content, got %d", w.Code)
	}
}

func TestHandleProfiles(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleProfiles(store)

	profile := &configv1.ProfileDefinition{
		Name: proto.String("dev"),
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/profiles", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleProfileDetail(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleProfileDetail(store)

	profile := &configv1.ProfileDefinition{
		Name: proto.String("dev"),
	}
	_ = store.SaveProfile(context.Background(), profile)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/profiles/dev", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test PUT
	// profile.Id field doesn't exist in ProfileDefinition usually?
	// Ah, ProfileDefinition has name. ID is likely not there or inferred.
	// Let's check proto.
	// ProfileDefinition: name, selector, required_roles, parent_profile_ids, service_config, secrets.
	// No ID field. Name is the key.

	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(profile)
	req = httptest.NewRequest(http.MethodPut, "/profiles/dev", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test Export
	req = httptest.NewRequest(http.MethodGet, "/profiles/dev/export", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test DELETE
	req = httptest.NewRequest(http.MethodDelete, "/profiles/dev", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 No Content, got %d", w.Code)
	}
}

func TestHandleCollections(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleCollections(store)

	collection := &configv1.UpstreamServiceCollectionShare{
		Name: proto.String("col1"),
	}
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(collection)
	req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201 Created, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/collections", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleCollectionDetail(t *testing.T) {
	app, store := setupCoverageTestApp()
	handler := app.handleCollectionDetail(store)

	collection := &configv1.UpstreamServiceCollectionShare{
		Name: proto.String("col1"),
		Services: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("svc1"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: proto.String("http://foo"),
					},
				},
			},
		},
	}
	_ = store.SaveServiceCollection(context.Background(), collection)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/collections/col1", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test PUT
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(collection)
	req = httptest.NewRequest(http.MethodPut, "/collections/col1", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test Export
	req = httptest.NewRequest(http.MethodGet, "/collections/col1/export", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test Apply
	req = httptest.NewRequest(http.MethodPost, "/collections/col1/apply", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test DELETE
	req = httptest.NewRequest(http.MethodDelete, "/collections/col1", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected 204 No Content, got %d", w.Code)
	}
}

func TestHandleExecute(t *testing.T) {
	app, _ := setupCoverageTestApp()
	handler := app.handleExecute()

	// 1. Invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/execute", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400, got %d", w.Code)
	}

	// 2. Unknown Tool
	execReq := &tool.ExecutionRequest{
		ToolName: "unknown",
	}
	body, _ := json.Marshal(execReq)
	req = httptest.NewRequest(http.MethodPost, "/execute", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	// ExecuteTool returns error for unknown tool, handler returns 500 or error message?
	// api.go:375: http.Error(w, err.Error(), http.StatusInternalServerError)
	// tool/manager.go: ExecuteTool returns error if not found.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}
