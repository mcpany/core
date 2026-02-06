// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	pb "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	mcp_router_v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/mcpany/core/server/pkg/audit"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/middleware"
	"github.com/mcpany/core/server/pkg/prompt"
	"github.com/mcpany/core/server/pkg/resource"
	"github.com/mcpany/core/server/pkg/skill"
	"github.com/mcpany/core/server/pkg/storage"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func setupApiTestApp() (*Application, storage.Storage) {
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

func TestIsUnsafeConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *configv1.UpstreamServiceConfig
		isUnsafe bool
	}{
		{
			name: "Safe HTTP Service",
			config: configv1.UpstreamServiceConfig_builder{
				HttpService: configv1.HttpUpstreamService_builder{}.Build(),
			}.Build(),
			isUnsafe: false,
		},
		{
			name: "Unsafe Command Line Service",
			config: configv1.UpstreamServiceConfig_builder{
				CommandLineService: configv1.CommandLineUpstreamService_builder{}.Build(),
			}.Build(),
			isUnsafe: true,
		},
		{
			name: "Unsafe MCP Stdio Service",
			config: configv1.UpstreamServiceConfig_builder{
				McpService: configv1.McpUpstreamService_builder{
					StdioConnection: configv1.McpStdioConnection_builder{}.Build(),
				}.Build(),
			}.Build(),
			isUnsafe: true,
		},
		{
			name: "Unsafe MCP Bundle Service",
			config: configv1.UpstreamServiceConfig_builder{
				McpService: configv1.McpUpstreamService_builder{
					BundleConnection: configv1.McpBundleConnection_builder{}.Build(),
				}.Build(),
			}.Build(),
			isUnsafe: true,
		},
		{
			name: "Safe MCP HTTP Service",
			config: configv1.UpstreamServiceConfig_builder{
				McpService: configv1.McpUpstreamService_builder{
					HttpConnection: configv1.McpStreamableHttpConnection_builder{}.Build(),
				}.Build(),
			}.Build(),
			isUnsafe: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isUnsafe, isUnsafeConfig(tt.config))
		})
	}
}

func TestHandleServiceStatus_Mocked(t *testing.T) {
	store := memory.NewStore()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockToolManager := tool.NewMockManagerInterface(ctrl)
	app := &Application{
		ToolManager: mockToolManager,
	}

	// Setup: Add a service to the store
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
	}.Build()
	require.NoError(t, store.SaveService(context.Background(), svc))

	t.Run("Status Inactive", func(t *testing.T) {
		mockToolManager.EXPECT().ListServices().Return(nil)

		req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "test-service", store)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Inactive", resp["status"])
	})

	t.Run("Status Active", func(t *testing.T) {
		mockToolManager.EXPECT().ListServices().Return([]*tool.ServiceInfo{
			{Name: "test-service"},
		})

		req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "test-service", store)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, "Active", resp["status"])
	})

	t.Run("Service Not Found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/services/unknown-service/status", nil)
		w := httptest.NewRecorder()

		app.handleServiceStatus(w, req, "unknown-service", store)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// Tests from api_coverage_test.go

func TestHandleServices(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleServices(store)

	// Test POST Success
	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		Id:   proto.String(uuid.New().String()),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080"),
		}.Build(),
	}.Build()
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
	svc.SetName("")
	body, _ = opts.Marshal(svc)
	req = httptest.NewRequest(http.MethodPost, "/services", bytes.NewReader(body))
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", w.Code)
	}
}

func TestHandleServiceDetail(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleServiceDetail(store)

	httpSvc := &configv1.HttpUpstreamService{}
	httpSvc.SetAddress("http://127.0.0.1:8080")

	svc := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("test-service"),
		Id:          proto.String(uuid.New().String()),
		HttpService: httpSvc,
	}.Build()
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
	svc.SetHttpService(configv1.HttpUpstreamService_builder{
		Address: proto.String("http://updated:8080"),
	}.Build())
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

func TestHandleServiceStatus_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleServiceDetail(store)

	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://127.0.0.1:8080"),
		}.Build(),
	}.Build()
	_ = store.SaveService(context.Background(), svc)

	req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleSettings_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleSettings(store)

	// Test POST
	settings := configv1.GlobalSettings_builder{
		AllowedIps: []string{"127.0.0.1"},
	}.Build()
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

func TestHandleTools_Detailed(t *testing.T) {
	app, _ := setupApiTestApp()
	handler := app.handleTools()

	req := httptest.NewRequest(http.MethodGet, "/tools", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandlePrompts_Detailed(t *testing.T) {
	app, _ := setupApiTestApp()
	handler := app.handlePrompts()

	req := httptest.NewRequest(http.MethodGet, "/prompts", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleResources_Detailed(t *testing.T) {
	app, _ := setupApiTestApp()
	handler := app.handleResources()

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleSecrets_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleSecrets(store)

	// Test POST
	secret := configv1.Secret_builder{
		Name:  proto.String("my-secret"),
		Id:    proto.String("my-secret-id"),
		Value: proto.String("super-secret"),
	}.Build()
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

func TestHandleSecretDetail_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleSecretDetail(store)

	secret := configv1.Secret_builder{
		Id:    proto.String("sec-123"),
		Name:  proto.String("my-secret"),
		Value: proto.String("super-secret"),
	}.Build()
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

func TestHandleProfiles_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleProfiles(store)

	profile := configv1.ProfileDefinition_builder{
		Name: proto.String("dev"),
	}.Build()
	opts := protojson.MarshalOptions{UseProtoNames: true}
	body, _ := opts.Marshal(profile)
	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/profiles", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

func TestHandleProfileDetail_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleProfileDetail(store)

	profile := configv1.ProfileDefinition_builder{
		Name: proto.String("dev"),
	}.Build()
	_ = store.SaveProfile(context.Background(), profile)

	// Test GET
	req := httptest.NewRequest(http.MethodGet, "/profiles/dev", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	// Test PUT
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

func TestHandleCollections_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleCollections(store)

	collection := configv1.Collection_builder{
		Name: proto.String("col1"),
	}.Build()
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

func TestHandleCollectionDetail_Detailed(t *testing.T) {
	app, store := setupApiTestApp()
	handler := app.handleCollectionDetail(store)

	svc1 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("svc1"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://foo"),
		}.Build(),
	}.Build()

	collection := configv1.Collection_builder{
		Name:     proto.String("col1"),
		Services: []*configv1.UpstreamServiceConfig{svc1},
	}.Build()
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

func TestHandleExecute_Detailed(t *testing.T) {
	app, _ := setupApiTestApp()
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
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected 500, got %d", w.Code)
	}
}

// Tests from api_extra_test.go

// Mock Resource
type mockResource struct {
	uri     string
	content string
}

func (m *mockResource) Resource() *mcp.Resource {
	return &mcp.Resource{URI: m.uri}
}
func (m *mockResource) Service() string { return "mock" }
func (m *mockResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      m.uri,
				Text:     m.content,
				MIMEType: "text/plain",
			},
		},
	}, nil
}
func (m *mockResource) Subscribe(ctx context.Context) error { return nil }

// Mock Prompt
type mockPrompt struct {
	name string
}

func (m *mockPrompt) Prompt() *mcp.Prompt {
	return &mcp.Prompt{Name: m.name}
}
func (m *mockPrompt) Service() string { return "mock" }
func (m *mockPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{
		Messages: []*mcp.PromptMessage{
			{
				Role: mcp.Role("user"),
				Content: &mcp.TextContent{
					Text: "Executed prompt " + m.name,
				},
			},
		},
	}, nil
}

func TestHandleResourceRead(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	app.ResourceManager = resource.NewManager()

	// Add a mock resource
	res := &mockResource{uri: "mock://test", content: "hello world"}
	app.ResourceManager.AddResource(res)

	handler := app.handleResourceRead()

	t.Run("ReadResource", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/read?uri=mock://test", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result mcp.ReadResourceResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Len(t, result.Contents, 1)

		content := result.Contents[0]
		assert.Equal(t, "mock://test", content.URI)
		assert.Equal(t, "hello world", content.Text)
	})

	t.Run("ReadResource_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/read?uri=mock://nonexistent", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ReadResource_MissingURI", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/resources/read", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/resources/read?uri=mock://test", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestHandlePromptExecute(t *testing.T) {
	fs := afero.NewMemMapFs()
	app := NewApplication()
	app.fs = fs
	app.PromptManager = prompt.NewManager()

	// Add a mock prompt
	p := &mockPrompt{name: "test-prompt"}
	app.PromptManager.AddPrompt(p)

	handler := app.handlePromptExecute()

	t.Run("ExecutePrompt", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt/execute", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var result mcp.GetPromptResult
		err := json.Unmarshal(w.Body.Bytes(), &result)
		require.NoError(t, err)
		require.Len(t, result.Messages, 1)

		content, ok := result.Messages[0].Content.(*mcp.TextContent)
		if ok {
			assert.Equal(t, "Executed prompt test-prompt", content.Text)
		}
	})

	t.Run("ExecutePrompt_NotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/prompts/nonexistent/execute", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ExecutePrompt_InvalidAction", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/prompts/test-prompt/other", bytes.NewReader([]byte("{}")))
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("MethodNotAllowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/prompts/test-prompt/execute", nil)
		w := httptest.NewRecorder()

		handler(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

// Tests from api_bug_test.go

func TestHandleProfiles_LargeBody(t *testing.T) {
	app, store := setupApiTestApp()

	// Create a large body (> 1MB)
	largeBody := make([]byte, 2*1024*1024) // 2MB

	req := httptest.NewRequest(http.MethodPost, "/profiles", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler := app.handleProfiles(store)
	handler.ServeHTTP(w, req)

	if w.Code == http.StatusRequestEntityTooLarge {
		t.Logf("Got 413 as expected")
	} else if w.Code != http.StatusBadRequest {
		// If bug exists, it might be 400 due to unmarshal error on random bytes
		t.Logf("Got %d, bug might still exist if not 413", w.Code)
	}
}

func TestHandleProfileDetail_LargeBody(t *testing.T) {
	app, store := setupApiTestApp()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPut, "/profiles/test", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler := app.handleProfileDetail(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 or 400, got %d", w.Code)
	}
}

func TestHandleSettings_LargeBody(t *testing.T) {
	app, store := setupApiTestApp()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPost, "/settings", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler := app.handleSettings(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 or 400, got %d", w.Code)
	}
}

func TestHandleCollections_LargeBody(t *testing.T) {
	app, store := setupApiTestApp()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPost, "/collections", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler := app.handleCollections(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 or 400, got %d", w.Code)
	}
}

func TestHandleCollectionDetail_LargeBody(t *testing.T) {
	app, store := setupApiTestApp()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPut, "/collections/test", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler := app.handleCollectionDetail(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 or 400, got %d", w.Code)
	}
}

func TestHandleSecrets_LargeBody(t *testing.T) {
	app, store := setupApiTestApp()

	largeBody := make([]byte, 2*1024*1024) // 2MB
	req := httptest.NewRequest(http.MethodPost, "/secrets", bytes.NewReader(largeBody))
	w := httptest.NewRecorder()

	handler := app.handleSecrets(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge && w.Code != http.StatusBadRequest {
		t.Errorf("Expected 413 or 400, got %d", w.Code)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, context.DeadlineExceeded
}

func TestHandleProfiles_ReadError(t *testing.T) {
	app, store := setupApiTestApp()

	req := httptest.NewRequest(http.MethodPost, "/profiles", &errorReader{})
	w := httptest.NewRecorder()

	handler := app.handleProfiles(store)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", w.Code)
	}
}

// Tests from api_error_test.go

type MockServiceRegistry struct {
	mock.Mock
}

func (m *MockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

func (m *MockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

func (m *MockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	args := m.Called(serviceID)
	if info := args.Get(0); info != nil {
		return info.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	args := m.Called(serviceID)
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*configv1.UpstreamServiceConfig), args.Bool(1)
	}
	return nil, args.Bool(1)
}

func (m *MockServiceRegistry) GetServiceError(serviceID string) (string, bool) {
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

func TestHandleServices_IncludesError(t *testing.T) {
	db, err := sqlite.NewDB(":memory:")
	require.NoError(t, err)
	defer db.Close()
	store := sqlite.NewStore(db)

	mockRegistry := new(MockServiceRegistry)
	service1 := &configv1.UpstreamServiceConfig{}
	service1.SetName("service-1")
	service1.SetId("service-1")

	service2 := &configv1.UpstreamServiceConfig{}
	service2.SetName("service-2")
	service2.SetId("service-2")

	service3 := &configv1.UpstreamServiceConfig{}
	service3.SetName("service-3")
	service3.SetSanitizedName("service-3-sanitized")

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{service1, service2, service3}, nil)
	mockRegistry.On("GetServiceError", "service-1").Return("", false)
	mockRegistry.On("GetServiceError", "service-2").Return("Connection refused", true)
	mockRegistry.On("GetServiceError", "service-3-sanitized").Return("Another error", true)

	app := NewApplication()
	app.ServiceRegistry = mockRegistry

	handler := app.handleServices(store)
	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var services []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&services)
	require.NoError(t, err)

	assert.Len(t, services, 3)

	var s1, s2, s3 map[string]interface{}
	for _, s := range services {
		if s["name"] == "service-1" {
			s1 = s
		} else if s["name"] == "service-2" {
			s2 = s
		} else if s["name"] == "service-3" {
			s3 = s
		}
	}

	assert.NotNil(t, s1)
	assert.NotNil(t, s2)
	assert.NotNil(t, s3)

	assert.Nil(t, s1["last_error"])
	assert.Equal(t, "Connection refused", s2["last_error"])
	assert.Equal(t, "Another error", s3["last_error"])
}

// Tests from api_security_test.go

func TestAPIHandler_SecurityValidation(t *testing.T) {
	store := memory.NewStore()
	app := &Application{
		ToolManager: tool.NewManager(nil),
	}

	handler := app.createAPIHandler(store)

	t.Run("Invalid URL Scheme", func(t *testing.T) {
		httpSvc := &configv1.HttpUpstreamService{}
		httpSvc.SetAddress("gopher://malicious.com")

		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("malicious-service")
		svc.SetHttpService(httpSvc)
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid service configuration")
	})

	t.Run("Absolute Bundle Path", func(t *testing.T) {
		bundleConn := &configv1.McpBundleConnection{}
		bundleConn.SetBundlePath("/etc/passwd")

		mcpSvc := &configv1.McpUpstreamService{}
		mcpSvc.SetBundleConnection(bundleConn)

		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("absolute-path-service")
		svc.SetMcpService(mcpSvc)
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid service configuration")
	})

	t.Run("Block Filesystem Service (Regular User)", func(t *testing.T) {
		osFs := &configv1.OsFs{}
		fsSvc := &configv1.FilesystemUpstreamService{}
		fsSvc.SetRootPaths(map[string]string{"/": "/"})
		fsSvc.SetOs(osFs)

		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("unsafe-fs")
		svc.SetFilesystemService(fsSvc)
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("Allow Filesystem Service (Admin)", func(t *testing.T) {
		osFs := &configv1.OsFs{}
		fsSvc := &configv1.FilesystemUpstreamService{}
		fsSvc.SetRootPaths(map[string]string{"/": "/"})
		fsSvc.SetOs(osFs)

		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("unsafe-fs-admin")
		svc.SetFilesystemService(fsSvc)
		body, _ := protojson.Marshal(svc)

		req := httptest.NewRequest("POST", "/services", bytes.NewReader(body))
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

// Tests from api_security_error_test.go

type errorResource struct{}

func (e *errorResource) Resource() *mcp.Resource { return &mcp.Resource{URI: "error://test"} }
func (e *errorResource) Service() string         { return "test" }
func (e *errorResource) Read(ctx context.Context) (*mcp.ReadResourceResult, error) {
	return nil, errors.New("read failed")
}
func (e *errorResource) Subscribe(ctx context.Context) error { return nil }

type errorPrompt struct{}

func (e *errorPrompt) Prompt() *mcp.Prompt { return &mcp.Prompt{Name: "error-prompt"} }
func (e *errorPrompt) Service() string     { return "test" }
func (e *errorPrompt) Get(ctx context.Context, args json.RawMessage) (*mcp.GetPromptResult, error) {
	return nil, errors.New("get failed")
}

func TestHandleResourceReadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockResManager := resource.NewMockManagerInterface(ctrl)
	app, _ := setupApiTestApp()
	app.ResourceManager = mockResManager

	mockResManager.EXPECT().GetResource("error://test").Return(&errorResource{}, true)

	req := httptest.NewRequest(http.MethodGet, "/resources/read?uri=error://test", nil)
	w := httptest.NewRecorder()

	app.handleResourceRead().ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandlePromptExecuteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPromptManager := prompt.NewMockManagerInterface(ctrl)
	app, _ := setupApiTestApp()
	app.PromptManager = mockPromptManager

	mockPromptManager.EXPECT().GetPrompt("error-prompt").Return(&errorPrompt{}, true)

	req := httptest.NewRequest(http.MethodPost, "/prompts/error-prompt/execute", nil)
	w := httptest.NewRecorder()

	app.handlePromptExecute().ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func stringPtr(s string) *string {
	return &s
}

func TestHandleAuditExport(t *testing.T) {
	app, _ := setupApiTestApp()
	app.standardMiddlewares = &middleware.StandardMiddlewares{}

	dbPath := "./audit_test_export.db"
	defer os.Remove(dbPath)

	sqliteStore, err := audit.NewSQLiteAuditStore(dbPath)
	require.NoError(t, err)
	entry1 := audit.Entry{
		Timestamp:  time.Now().Add(-1 * time.Hour),
		ToolName:   "tool-1",
		UserID:     "user-1",
		DurationMs: 123,
		Arguments:  []byte(`{"key":"val"}`),
	}
	require.NoError(t, sqliteStore.Write(context.Background(), entry1))
	sqliteStore.Close()

	storageType := configv1.AuditConfig_STORAGE_TYPE_SQLITE
	auditCfg := &configv1.AuditConfig{}
	auditCfg.SetEnabled(true)
	auditCfg.SetStorageType(storageType)
	auditCfg.SetOutputPath(dbPath)
	audit, err := middleware.NewAuditMiddleware(auditCfg)
	require.NoError(t, err)
	app.standardMiddlewares.Audit = audit
	defer audit.Close()

	req, _ := http.NewRequest("GET", "/audit/export?tool_name=tool-1", nil)
	rr := httptest.NewRecorder()
	mux := app.createAPIHandler(app.Storage)
	mux.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "text/csv", rr.Header().Get("Content-Type"))

	csvReader := csv.NewReader(rr.Body)
	records, err := csvReader.ReadAll()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(records), 2)
}

func TestHandleInitiateOAuth(t *testing.T) {
	store := memory.NewStore()
	am := auth.NewManager()
	am.SetStorage(store)
	app := &Application{AuthManager: am}

	svcID := "github"
	clientId := &configv1.SecretValue{}
	clientId.SetPlainText("client-id")
	clientSecret := &configv1.SecretValue{}
	clientSecret.SetPlainText("client-secret")

	oauth2 := &configv1.OAuth2Auth{}
	oauth2.SetClientId(clientId)
	oauth2.SetClientSecret(clientSecret)
	oauth2.SetAuthorizationUrl("https://github.com/login/oauth/authorize")
	oauth2.SetTokenUrl("https://github.com/login/oauth/access_token")

	authConfig := &configv1.Authentication{}
	authConfig.SetOauth2(oauth2)

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName(svcID)
	svc.SetUpstreamAuth(authConfig)
	store.SaveService(context.Background(), svc)

	t.Run("Success_Service", func(t *testing.T) {
		body := map[string]string{"service_id": svcID, "redirect_url": "http://localhost/cb"}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(http.MethodPost, "/auth/oauth/initiate", bytes.NewReader(bodyBytes))
		req = req.WithContext(auth.ContextWithUser(req.Context(), "user1"))
		w := httptest.NewRecorder()
		app.handleInitiateOAuth(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestHandleLogsWS(t *testing.T) {
	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	time.Sleep(100 * time.Millisecond)
	testMsg := []byte("test log message")
	logging.GlobalBroadcaster.Broadcast(testMsg)

	_, msg, err := ws.ReadMessage()
	require.NoError(t, err)
	assert.Equal(t, testMsg, msg)
}


func TestHandleUsers(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	t.Run("CreateUser", func(t *testing.T) {
		user := &configv1.User{}
		user.SetId("user1")
		userBytes, _ := protojson.Marshal(user)
		body, _ := json.Marshal(map[string]json.RawMessage{"user": json.RawMessage(userBytes)})
		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
		// Inject admin role
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)
	})
}

func TestCreateUser_PasswordHashing(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	payload := `{"user": {"id": "test-user-hash", "authentication": {"basic_auth": {"username": "test", "password_hash": "plain-password"}}}}`
	req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(payload))
	// Inject admin role
	ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler(w, req)

	user, _ := store.GetUser(context.Background(), "test-user-hash")
	assert.True(t, strings.HasPrefix(user.GetAuthentication().GetBasicAuth().GetPasswordHash(), "$2a$"))
}

func TestHandleUsers_Security_Redaction(t *testing.T) {
	app := NewApplication()
	app.fs = afero.NewMemMapFs()
	app.AuthManager = auth.NewManager()
	store := memory.NewStore()
	app.Storage = store
	handler := app.handleUsers(store)

	user := &configv1.User{}
	user.SetId("secure-user")

	apiKeyAuth := &configv1.APIKeyAuth{}
	apiKeyAuth.SetVerificationValue("super-secret-key")

	authConfig := &configv1.Authentication{}
	authConfig.SetApiKey(apiKeyAuth)
	user.SetAuthentication(authConfig)
	store.CreateUser(context.Background(), user)

	t.Run("ListUsers_ShouldNotLeakSecrets", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		// Inject admin role
		ctx := auth.ContextWithRoles(req.Context(), []string{"admin"})
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		handler(w, req)
		assert.NotContains(t, w.Body.String(), "super-secret-key")
	})
}

func TestCheckURLReachability(t *testing.T) {
	// Enable loopback for this test since we are testing reachability against a local test server.
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	assert.NoError(t, checkURLReachability(context.Background(), server.URL))
}

func TestHandleServiceValidate(t *testing.T) {
	app := &Application{}
	httpSvc := &configv1.HttpUpstreamService{}
	httpSvc.SetAddress("http://example.com")

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("test-service")
	svc.SetHttpService(httpSvc)
	body, _ := protojson.Marshal(svc)
	req := httptest.NewRequest(http.MethodPost, "/services/validate", bytes.NewReader(body))
	w := httptest.NewRecorder()
	app.handleServiceValidate().ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUploadFile_Security(t *testing.T) {
	app := NewApplication()
	t.Run("Reflected XSS", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test&file\"name.txt")
		part.Write([]byte("content"))
		writer.Close()
		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		app.uploadFile(w, req)
		assert.NotContains(t, w.Body.String(), "test&file\"name.txt")
		assert.Contains(t, w.Body.String(), "test_file_name.txt")
	})
}

type TestMockTool struct {
	toolDef *mcp_router_v1.Tool
}

func (m *TestMockTool) Tool() *mcp_router_v1.Tool { return m.toolDef }
func (m *TestMockTool) MCPTool() *mcp.Tool        { return nil }
func (m *TestMockTool) Execute(ctx context.Context, req *tool.ExecutionRequest) (any, error) {
	return nil, nil
}
func (m *TestMockTool) GetCacheConfig() *configv1.CacheConfig { return nil }

type MockServiceStore struct {
	services []*configv1.UpstreamServiceConfig
}

func (s *MockServiceStore) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	return nil, nil
}
func (s *MockServiceStore) HasConfigSources() bool { return false }
func (s *MockServiceStore) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return nil
}
func (s *MockServiceStore) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	return nil, nil
}
func (s *MockServiceStore) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	return s.services, nil
}
func (s *MockServiceStore) DeleteService(ctx context.Context, name string) error  { return nil }
func (s *MockServiceStore) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	return nil, nil
}
func (s *MockServiceStore) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	return nil
}
func (s *MockServiceStore) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	return nil, nil
}
func (s *MockServiceStore) DeleteSecret(ctx context.Context, id string) error { return nil }
func (s *MockServiceStore) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	return nil, nil
}
func (s *MockServiceStore) SaveProfile(ctx context.Context, p *configv1.ProfileDefinition) error {
	return nil
}
func (s *MockServiceStore) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	return nil, nil
}
func (s *MockServiceStore) DeleteProfile(ctx context.Context, name string) error { return nil }
func (s *MockServiceStore) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) {
	return nil, nil
}
func (s *MockServiceStore) SaveServiceCollection(ctx context.Context, c *configv1.Collection) error {
	return nil
}
func (s *MockServiceStore) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	return nil, nil
}
func (s *MockServiceStore) DeleteServiceCollection(ctx context.Context, name string) error {
	return nil
}
func (s *MockServiceStore) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	return nil, nil
}
func (s *MockServiceStore) SaveGlobalSettings(ctx context.Context, gs *configv1.GlobalSettings) error {
	return nil
}
func (s *MockServiceStore) Close() error                                           { return nil }
func (s *MockServiceStore) CreateUser(ctx context.Context, user *configv1.User) error { return nil }
func (s *MockServiceStore) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	return nil, nil
}
func (s *MockServiceStore) ListUsers(ctx context.Context) ([]*configv1.User, error) { return nil, nil }
func (s *MockServiceStore) UpdateUser(ctx context.Context, user *configv1.User) error { return nil }
func (s *MockServiceStore) DeleteUser(ctx context.Context, id string) error           { return nil }
func (s *MockServiceStore) SaveToken(ctx context.Context, token *configv1.UserToken) error {
	return nil
}
func (s *MockServiceStore) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	return nil, nil
}
func (s *MockServiceStore) DeleteToken(ctx context.Context, userID, serviceID string) error {
	return nil
}
func (s *MockServiceStore) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	return nil, nil
}
func (s *MockServiceStore) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	return nil, nil
}
func (s *MockServiceStore) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	return nil
}
func (s *MockServiceStore) DeleteCredential(ctx context.Context, id string) error { return nil }

type TestMockServiceRegistry struct {
	services []*configv1.UpstreamServiceConfig
}

func (m *TestMockServiceRegistry) RegisterService(ctx context.Context, serviceConfig *configv1.UpstreamServiceConfig) (string, []*configv1.ToolDefinition, []*configv1.ResourceDefinition, error) {
	return "", nil, nil, nil
}
func (m *TestMockServiceRegistry) UnregisterService(ctx context.Context, serviceName string) error {
	return nil
}
func (m *TestMockServiceRegistry) GetAllServices() ([]*configv1.UpstreamServiceConfig, error) {
	return m.services, nil
}
func (m *TestMockServiceRegistry) GetServiceInfo(serviceID string) (*tool.ServiceInfo, bool) {
	return nil, false
}
func (m *TestMockServiceRegistry) GetServiceConfig(serviceID string) (*configv1.UpstreamServiceConfig, bool) {
	return nil, false
}
func (m *TestMockServiceRegistry) GetServiceError(serviceID string) (string, bool) { return "", false }

func TestHandleServices_ToolCount(t *testing.T) {
	busProvider, _ := bus.NewProvider(nil)
	tm := tool.NewManager(busProvider)

	tm.AddTool(&TestMockTool{toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool1"), ServiceId: proto.String("service-1")}.Build()})
	tm.AddTool(&TestMockTool{toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool2"), ServiceId: proto.String("service-1")}.Build()})
	tm.AddTool(&TestMockTool{toolDef: mcp_router_v1.Tool_builder{Name: proto.String("tool3"), ServiceId: proto.String("service-2")}.Build()})

	app := NewApplication()
	app.ToolManager = tm

	app.ServiceRegistry = &TestMockServiceRegistry{
		services: func() []*configv1.UpstreamServiceConfig {
			s1 := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("service-1"),
				Id:   proto.String("service-1"),
			}.Build()
			s2 := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("service-2"),
				Id:   proto.String("service-2"),
			}.Build()
			s3 := configv1.UpstreamServiceConfig_builder{
				Name: proto.String("service-3"),
				Id:   proto.String("service-3"),
			}.Build()
			return []*configv1.UpstreamServiceConfig{s1, s2, s3}
		}(),
	}

	req := httptest.NewRequest(http.MethodGet, "/services", nil)
	rr := httptest.NewRecorder()

	handler := app.handleServices(&MockServiceStore{})
	handler.ServeHTTP(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)

	var response []map[string]any
	json.Unmarshal(rr.Body.Bytes(), &response)
	require.Len(t, response, 3)

	for _, s := range response {
		if s["name"] == "service-1" {
			assert.Equal(t, float64(2), s["tool_count"])
		}
	}
}

func TestSkillServiceServer(t *testing.T) {
	tempDir := t.TempDir()
	manager, _ := skill.NewManager(tempDir)
	server := NewSkillServiceServer(manager)
	ctx := context.Background()

	t.Run("CreateSkill", func(t *testing.T) {
		req := pb.CreateSkillRequest_builder{
			Skill: configv1.Skill_builder{
				Name:         proto.String("test-skill"),
				Description:  proto.String("A test skill"),
				Instructions: proto.String("Do something"),
			}.Build(),
		}.Build()
		resp, err := server.CreateSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-skill", resp.GetSkill().GetName())
	})

	t.Run("GetSkill", func(t *testing.T) {
		req := pb.GetSkillRequest_builder{Name: "test-skill"}.Build()
		resp, err := server.GetSkill(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-skill", resp.GetSkill().GetName())
	})

	t.Run("DeleteSkill", func(t *testing.T) {
		req := pb.DeleteSkillRequest_builder{Name: "test-skill"}.Build()
		_, err := server.DeleteSkill(ctx, req)
		require.NoError(t, err)
	})
}

func TestConfigDiffGeneration(t *testing.T) {
	app := NewApplication()
	fs := afero.NewMemMapFs()
	ctx := context.Background()
	configPath := "config.yaml"

	initialConfig := `
upstream_services:
  - name: "echo"
    http_service:
      address: "http://echo.service"
`
	afero.WriteFile(fs, configPath, []byte(initialConfig), 0644)
	app.ReloadConfig(ctx, fs, []string{configPath})

	invalidConfig := `
upstream_services:
  - name: "echo"
    http_service:
      address: "http://echo.service"
  - invalid_indentation
`
	afero.WriteFile(fs, configPath, []byte(invalidConfig), 0644)
	err := app.ReloadConfig(ctx, fs, []string{configPath})
	assert.Error(t, err)
	assert.Contains(t, app.configDiff, "invalid_indentation")
}

func TestSecretLeak(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_secrets.db")
	db, _ := sqlite.NewDB(dbPath)
	defer db.Close()
	store := sqlite.NewStore(db)

	app := NewApplication()
	app.fs = afero.NewMemMapFs()

	handler := app.createAPIHandler(store)
	ts := httptest.NewServer(handler)
	defer ts.Close()

	secretID := "sensitive-secret-123"
	body := map[string]interface{}{
		"id":   secretID,
		"name": "My Secret",
		"key":  "my_secret_key",
		"value": "SUPER_SECRET_VALUE",
	}
	bodyBytes, _ := json.Marshal(body)
	http.Post(ts.URL+"/secrets", "application/json", bytes.NewReader(bodyBytes))

	resp, _ := http.Get(ts.URL + "/secrets/" + secretID)
	var result map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "failed to decode response body")
	assert.Equal(t, "[REDACTED]", result["value"])
}


func TestReproduction_ProtocolCompliance(t *testing.T) {
	fs := afero.NewMemMapFs()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l, _ := net.Listen("tcp", "127.0.0.1:0")
	httpPort := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	app := NewApplication()
	afero.WriteFile(fs, "/config.yaml", []byte("upstream_services: []"), 0o644)

	errChan := make(chan error, 1)
	go func() {
		errChan <- app.Run(RunOptions{Ctx: ctx, Fs: fs, Stdio: false, JSONRPCPort: fmt.Sprintf("127.0.0.1:%d", httpPort), GRPCPort: "127.0.0.1:0", ConfigPaths: []string{"/config.yaml"}, APIKey: "", ShutdownTimeout: 5*time.Second})
	}()

	require.NoError(t, app.WaitForStartup(ctx))
	baseURL := fmt.Sprintf("http://127.0.0.1:%d/mcp", httpPort)
	// Use local HealthCheck polling instead of integration package to avoid cycle
	require.Eventually(t, func() bool {
		return HealthCheck(io.Discard, fmt.Sprintf("127.0.0.1:%d", httpPort), 100*time.Millisecond) == nil
	}, 5*time.Second, 100*time.Millisecond)

	reqBody := `{"jsonrpc": "2.0", "method": "non_existent_method", "id": 1}`
	resp, err := http.Post(baseURL, "application/json", bytes.NewBufferString(reqBody))
	require.NoError(t, err, "http.Post failed")
	defer func() { _ = resp.Body.Close() }()
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err, "failed to decode response body")
	assert.NotNil(t, result["error"])

	cancel()
	<-errChan
}

func TestHandleAuthTest(t *testing.T) {
	store := memory.NewStore()
	am := auth.NewManager()
	am.SetStorage(store)
	app := &Application{AuthManager: am, Storage: store}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") == "mykey" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusUnauthorized)
		}
	}))
	defer ts.Close()

	credID := "cred-http"
	loc := configv1.APIKeyAuth_HEADER
	cred := configv1.Credential_builder{
		Id: proto.String(credID),
		Authentication: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				In:                &loc,
				ParamName:         proto.String("X-API-Key"),
				VerificationValue: proto.String("mykey"),
			}.Build(),
		}.Build(),
	}.Build()
	store.SaveCredential(context.Background(), cred)

	req := AuthTestRequest{
		CredentialID: credID,
		ServiceType:  "HTTP",
		ServiceConfig: map[string]any{"http_service": map[string]any{"address": ts.URL}},
	}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
	w := httptest.NewRecorder()
	app.handleAuthTest()(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}
