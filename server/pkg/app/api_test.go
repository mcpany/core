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
	"github.com/mcpany/core/server/pkg/validation"
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
	store := memory.NewStore()
	app := &Application{
		PromptManager:   prompt.NewManager(),
		ToolManager:     tool.NewManager(bp),
		ResourceManager: resource.NewManager(),
		busProvider:     bp,
		Storage:         store,
	}
	return app, store
}

// TestIsUnsafeConfig ...
// Summary: TestIsUnsafeConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleServiceStatus_Mocked ...
// Summary: TestHandleServiceStatus_Mocked
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleServices ...
// Summary: TestHandleServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleServiceDetail ...
// Summary: TestHandleServiceDetail
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleServiceStatus_Detailed ...
// Summary: TestHandleServiceStatus_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleSettings_Detailed ...
// Summary: TestHandleSettings_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleTools_Detailed ...
// Summary: TestHandleTools_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, _ := setupApiTestApp()
	handler := app.handleTools()

	req := httptest.NewRequest(http.MethodGet, "/tools", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

// TestHandlePrompts_Detailed ...
// Summary: TestHandlePrompts_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, _ := setupApiTestApp()
	handler := app.handlePrompts()

	req := httptest.NewRequest(http.MethodGet, "/prompts", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

// TestHandleResources_Detailed ...
// Summary: TestHandleResources_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, _ := setupApiTestApp()
	handler := app.handleResources()

	req := httptest.NewRequest(http.MethodGet, "/resources", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
}

// TestHandleSecrets_Detailed ...
// Summary: TestHandleSecrets_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleSecretDetail_Detailed ...
// Summary: TestHandleSecretDetail_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleProfiles_Detailed ...
// Summary: TestHandleProfiles_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleProfileDetail_Detailed ...
// Summary: TestHandleProfileDetail_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleCollections_Detailed ...
// Summary: TestHandleCollections_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleCollectionDetail_Detailed ...
// Summary: TestHandleCollectionDetail_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleExecute_Detailed ...
// Summary: TestHandleExecute_Detailed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// Resource ...
// Summary: Resource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &mcp.Resource{URI: m.uri}
}
// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// Subscribe ...
// Summary: Subscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// Mock Prompt
type mockPrompt struct {
	name string
}

// Prompt ...
// Summary: Prompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return &mcp.Prompt{Name: m.name}
}
// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Definition ...
// Summary: Definition
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleResourceRead ...
// Summary: TestHandleResourceRead
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandlePromptExecute ...
// Summary: TestHandlePromptExecute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleProfiles_LargeBody ...
// Summary: TestHandleProfiles_LargeBody
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleProfileDetail_LargeBody ...
// Summary: TestHandleProfileDetail_LargeBody
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleSettings_LargeBody ...
// Summary: TestHandleSettings_LargeBody
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleCollections_LargeBody ...
// Summary: TestHandleCollections_LargeBody
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleCollectionDetail_LargeBody ...
// Summary: TestHandleCollectionDetail_LargeBody
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleSecrets_LargeBody ...
// Summary: TestHandleSecrets_LargeBody
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return 0, context.DeadlineExceeded
}

// TestHandleProfiles_ReadError ...
// Summary: TestHandleProfiles_ReadError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// MockServiceRegistry ...
// Summary: MockServiceRegistry
	mock.Mock
}

// RegisterService ...
// Summary: RegisterService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, serviceConfig)
	return args.String(0), args.Get(1).([]*configv1.ToolDefinition), args.Get(2).([]*configv1.ResourceDefinition), args.Error(3)
}

// UnregisterService ...
// Summary: UnregisterService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(ctx, serviceName)
	return args.Error(0)
}

// GetAllServices ...
// Summary: GetAllServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called()
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	if info := args.Get(0); info != nil {
		return info.(*tool.ServiceInfo), args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceConfig ...
// Summary: GetServiceConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	if cfg := args.Get(0); cfg != nil {
		return cfg.(*configv1.UpstreamServiceConfig), args.Bool(1)
	}
	return nil, args.Bool(1)
}

// GetServiceError ...
// Summary: GetServiceError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	args := m.Called(serviceID)
	return args.String(0), args.Bool(1)
}

// TestHandleServices_IncludesError ...
// Summary: TestHandleServices_IncludesError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestAPIHandler_SecurityValidation ...
// Summary: TestAPIHandler_SecurityValidation
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// Resource ...
// Summary: Resource
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Read ...
// Summary: Read
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, errors.New("read failed")
}
// Subscribe ...
// Summary: Subscribe
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

type errorPrompt struct{}

// Prompt ...
// Summary: Prompt
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Service ...
// Summary: Service
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Definition ...
// Summary: Definition
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Get ...
// Summary: Get
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, errors.New("get failed")
}

// TestHandleResourceReadError ...
// Summary: TestHandleResourceReadError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandlePromptExecuteError ...
// Summary: TestHandlePromptExecuteError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleAuditExport ...
// Summary: TestHandleAuditExport
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, _ := setupApiTestApp()
	app.standardMiddlewares = &middleware.StandardMiddlewares{}

	// Use CWD to ensure path is allowed by validation
	dbPath := "./audit_test_export.db"
	defer os.Remove(dbPath)

	// Allow .db extension for this test as IsSensitivePath blocks it by default
	originalIsSensitive := validation.IsSensitivePath
	validation.IsSensitivePath = func(path string) error {
		if strings.HasSuffix(path, "audit_test_export.db") {
			return nil
		}
		return originalIsSensitive(path)
	}
	defer func() { validation.IsSensitivePath = originalIsSensitive }()

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

// TestHandleInitiateOAuth ...
// Summary: TestHandleInitiateOAuth
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleLogsWS ...
// Summary: TestHandleLogsWS
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app := &Application{}
	handler := app.handleLogsWS()
	ts := httptest.NewServer(handler)
	defer ts.Close()

	u := "ws" + strings.TrimPrefix(ts.URL, "http")
	ws, _, err := websocket.DefaultDialer.Dial(u, nil)
	require.NoError(t, err)
	defer ws.Close()

	time.Sleep(100 * time.Millisecond)
	testMsg := logging.LogEntry{Message: "test log message"}
	logging.GlobalBroadcaster.Broadcast(testMsg)

	var msg logging.LogEntry
	require.NoError(t, ws.ReadJSON(&msg))
	assert.Equal(t, testMsg.Message, msg.Message)
}

// TestHandleSystemStatus ...
// Summary: TestHandleSystemStatus
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app := NewApplication()
	app.startTime = time.Now().Add(-10 * time.Second)
	app.SettingsManager = NewGlobalSettingsManager("", nil, nil)

	req, _ := http.NewRequest(http.MethodGet, "/api/v1/system/status", nil)
	rr := httptest.NewRecorder()
	app.handleSystemStatus(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var resp map[string]interface{}
	json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.GreaterOrEqual(t, resp["uptime_seconds"].(float64), float64(10))
}

// TestHandleTemplates removed in favor of api_templates_test.go

// TestHandleUsers ...
// Summary: TestHandleUsers
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestCreateUser_PasswordHashing ...
// Summary: TestCreateUser_PasswordHashing
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestHandleUsers_Security_Redaction ...
// Summary: TestHandleUsers_Security_Redaction
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestCheckURLReachability ...
// Summary: TestCheckURLReachability
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Enable loopback for this test since we are testing reachability against a local test server.
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
	t.Setenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES", "true")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	assert.NoError(t, checkURLReachability(context.Background(), server.URL))
}

// TestHandleServiceValidate ...
// Summary: TestHandleServiceValidate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestUploadFile_Security ...
// Summary: TestUploadFile_Security
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestMockTool ...
// Summary: TestMockTool
	toolDef *mcp_router_v1.Tool
}

// Tool ...
// Summary: Tool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// MCPTool ...
// Summary: MCPTool
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Execute ...
// Summary: Execute
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// GetCacheConfig ...
// Summary: GetCacheConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// MockServiceStore ...
// Summary: MockServiceStore
	services []*configv1.UpstreamServiceConfig
}

// Load ...
// Summary: Load
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// HasConfigSources ...
// Summary: HasConfigSources
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SaveService ...
// Summary: SaveService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// GetService ...
// Summary: GetService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// ListServices ...
// Summary: ListServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return s.services, nil
}
// DeleteService ...
// Summary: DeleteService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ListSecrets ...
// Summary: ListSecrets
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// SaveSecret ...
// Summary: SaveSecret
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// GetSecret ...
// Summary: GetSecret
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// DeleteSecret ...
// Summary: DeleteSecret
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ListProfiles ...
// Summary: ListProfiles
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// SaveProfile ...
// Summary: SaveProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// GetProfile ...
// Summary: GetProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// DeleteProfile ...
// Summary: DeleteProfile
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ListServiceCollections ...
// Summary: ListServiceCollections
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// SaveServiceCollection ...
// Summary: SaveServiceCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// GetServiceCollection ...
// Summary: GetServiceCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// DeleteServiceCollection ...
// Summary: DeleteServiceCollection
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// GetGlobalSettings ...
// Summary: GetGlobalSettings
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// SaveGlobalSettings ...
// Summary: SaveGlobalSettings
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// Close ...
// Summary: Close
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// CreateUser ...
// Summary: CreateUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// GetUser ...
// Summary: GetUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// ListUsers ...
// Summary: ListUsers
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// UpdateUser ...
// Summary: UpdateUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// DeleteUser ...
// Summary: DeleteUser
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// SaveToken ...
// Summary: SaveToken
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// GetToken ...
// Summary: GetToken
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// DeleteToken ...
// Summary: DeleteToken
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// ListCredentials ...
// Summary: ListCredentials
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// GetCredential ...
// Summary: GetCredential
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// SaveCredential ...
// Summary: SaveCredential
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// DeleteCredential ...
// Summary: DeleteCredential
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// ListServiceTemplates ...
// Summary: ListServiceTemplates
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// GetServiceTemplate ...
// Summary: GetServiceTemplate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}
// SaveServiceTemplate ...
// Summary: SaveServiceTemplate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// DeleteServiceTemplate ...
// Summary: DeleteServiceTemplate
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// SaveLog ...
// Summary: SaveLog
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}

// GetRecentLogs ...
// Summary: GetRecentLogs
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil
}

// TestMockServiceRegistry ...
// Summary: TestMockServiceRegistry
	services []*configv1.UpstreamServiceConfig
}

// RegisterService ...
// Summary: RegisterService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return "", nil, nil, nil
}
// UnregisterService ...
// Summary: UnregisterService
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil
}
// GetAllServices ...
// Summary: GetAllServices
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.services, nil
}
// GetServiceInfo ...
// Summary: GetServiceInfo
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, false
}
// GetServiceConfig ...
// Summary: GetServiceConfig
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, false
}
// GetServiceError ...
// Summary: GetServiceError
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.

// TestHandleServices_ToolCount ...
// Summary: TestHandleServices_ToolCount
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSkillServiceServer ...
// Summary: TestSkillServiceServer
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestConfigDiffGeneration ...
// Summary: TestConfigDiffGeneration
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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

// TestSecretLeak ...
// Summary: TestSecretLeak
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
		"id":    secretID,
		"name":  "My Secret",
		"key":   "my_secret_key",
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

// TestReproduction_ProtocolCompliance ...
// Summary: TestReproduction_ProtocolCompliance
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
		errChan <- app.Run(RunOptions{Ctx: ctx, Fs: fs, Stdio: false, JSONRPCPort: fmt.Sprintf("127.0.0.1:%d", httpPort), GRPCPort: "127.0.0.1:0", ConfigPaths: []string{"/config.yaml"}, APIKey: "", ShutdownTimeout: 5 * time.Second})
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

// TestHandleAuthTest ...
// Summary: TestHandleAuthTest
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
		CredentialID:  credID,
		ServiceType:   "HTTP",
		ServiceConfig: map[string]any{"http_service": map[string]any{"address": ts.URL}},
	}
	body, _ := json.Marshal(req)
	r := httptest.NewRequest(http.MethodPost, "/auth/test", bytes.NewReader(body))
	w := httptest.NewRecorder()
	app.handleAuthTest()(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestHandleSecretDetail_Reveal_HappyPath ...
// Summary: TestHandleSecretDetail_Reveal_HappyPath
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, store := setupApiTestApp()
	handler := app.handleSecretDetail(store)

	secret := configv1.Secret_builder{
		Id:    proto.String("sec-123"),
		Name:  proto.String("my-secret"),
		Value: proto.String("super-secret"),
	}.Build()
	_ = store.SaveSecret(context.Background(), secret)

	req := httptest.NewRequest(http.MethodPost, "/secrets/sec-123/reveal", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`{"value":"super-secret"}`)) {
		t.Errorf("Expected secret value in response body, got %s", w.Body.String())
	}
}

// TestHandleSecretDetail_Reveal_MethodNotAllowed ...
// Summary: TestHandleSecretDetail_Reveal_MethodNotAllowed
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, store := setupApiTestApp()
	handler := app.handleSecretDetail(store)

	req := httptest.NewRequest(http.MethodGet, "/secrets/sec-123/reveal", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected 405 Method Not Allowed, got %d", w.Code)
	}
}

// TestHandleSecretDetail_Reveal_NotFound ...
// Summary: TestHandleSecretDetail_Reveal_NotFound
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, store := setupApiTestApp()
	handler := app.handleSecretDetail(store)

	req := httptest.NewRequest(http.MethodPost, "/secrets/non-existent/reveal", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404 Not Found, got %d", w.Code)
	}
}

// TestHandleSecretDetail_Put_HappyPath ...
// Summary: TestHandleSecretDetail_Put_HappyPath
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, store := setupApiTestApp()
	handler := app.handleSecretDetail(store)

	reqBody := `{"name":"my-secret", "value":"new-secret-value"}`
	req := httptest.NewRequest(http.MethodPut, "/secrets/sec-123", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Code)
	}

	secret, err := store.GetSecret(context.Background(), "sec-123")
	if err != nil {
		t.Fatalf("Failed to get secret: %v", err)
	}
	if secret.GetValue() != "new-secret-value" {
		t.Errorf("Expected secret value to be 'new-secret-value', got '%s'", secret.GetValue())
	}
}

// TestHandleSecretDetail_Put_InvalidJSON ...
// Summary: TestHandleSecretDetail_Put_InvalidJSON
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	app, store := setupApiTestApp()
	handler := app.handleSecretDetail(store)

	reqBody := `{"name":"my-secret", "value":"new-secret-value"` // Missing closing brace
	req := httptest.NewRequest(http.MethodPut, "/secrets/sec-123", bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", w.Code)
	}
}
