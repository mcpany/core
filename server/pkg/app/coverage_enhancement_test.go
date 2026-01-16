// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// TestHandleServiceStatus_EdgeCases covers handleServiceStatus edge cases
func TestHandleServiceStatus_EdgeCases(t *testing.T) {
	app := setupTestApp()
	store := app.Storage

	// 1. Create a service
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_HttpConnection{
					HttpConnection: &configv1.McpStreamableHttpConnection{
						HttpAddress: proto.String("http://localhost:8080"),
					},
				},
			},
		},
	}
	err := store.SaveService(context.Background(), svc)
	require.NoError(t, err)

	// 2. Query status (GET)
	req := httptest.NewRequest(http.MethodGet, "/services/test-service/status", nil)
	rr := httptest.NewRecorder()

	// Direct call or via router
	handler := app.handleServiceDetail(store)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var statusMap map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &statusMap)
	require.NoError(t, err)
	assert.Equal(t, "test-service", statusMap["name"])
	assert.Contains(t, statusMap, "status")

	// 3. Invalid Method
	req = httptest.NewRequest(http.MethodPost, "/services/test-service/status", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// 4. Non-existent service
	req = httptest.NewRequest(http.MethodGet, "/services/missing/status", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHandleCollectionApply_EdgeCases(t *testing.T) {
	app := setupTestApp()
	store := app.Storage

	// 1. Create Collection
	col := &configv1.Collection{
		Name: proto.String("test-col"),
		Services: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("svc-in-col"),
				ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
					McpService: &configv1.McpUpstreamService{
						ConnectionType: &configv1.McpUpstreamService_HttpConnection{
							HttpConnection: &configv1.McpStreamableHttpConnection{
								HttpAddress: proto.String("http://example.com"),
							},
						},
					},
				},
			},
		},
	}
	err := store.SaveServiceCollection(context.Background(), col)
	require.NoError(t, err)

	// 2. Apply (POST)
	req := httptest.NewRequest(http.MethodPost, "/collections/test-col/apply", nil)
	rr := httptest.NewRecorder()

	handler := app.handleCollectionDetail(store)
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	// Verify service created
	svc, err := store.GetService(context.Background(), "svc-in-col")
	require.NoError(t, err)
	require.NotNil(t, svc)

	// 3. Invalid Method
	req = httptest.NewRequest(http.MethodGet, "/collections/test-col/apply", nil)
	rr = httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	// handleCollectionApply returns 405 if not POST
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
}

func TestHandleExecute_Errors_EdgeCases(t *testing.T) {
	app := setupTestApp()

	// 1. Wrong Method
	req := httptest.NewRequest(http.MethodGet, "/execute", nil)
	rr := httptest.NewRecorder()
	app.handleExecute().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)

	// 2. Invalid JSON
	req = httptest.NewRequest(http.MethodPost, "/execute", bytes.NewBufferString("{invalid"))
	rr = httptest.NewRecorder()
	app.handleExecute().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)

	// 3. Valid JSON but execution fails (no tool)
	req = httptest.NewRequest(http.MethodPost, "/execute", bytes.NewBufferString(`{"tool_name":"missing"}`))
	rr = httptest.NewRecorder()
	app.handleExecute().ServeHTTP(rr, req)
	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestHandleUsers_Coverage_EdgeCases(t *testing.T) {
	app := setupTestApp()
	store := app.Storage

	// 1. Create User
	user := &configv1.User{
		Id: proto.String("test-user"),
		Roles: []string{"admin"},
	}
	body, _ := json.Marshal(user)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	app.handleUsers(store).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusCreated, rr.Code)

	// 2. Get Users
	req = httptest.NewRequest(http.MethodGet, "/users", nil)
	rr = httptest.NewRecorder()
	app.handleUsers(store).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "test-user")

	// 3. Get User Detail
	req = httptest.NewRequest(http.MethodGet, "/users/test-user", nil)
	rr = httptest.NewRecorder()
	app.handleUserDetail(store).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// 4. Update User
	user.Roles = []string{"editor"}
	body, _ = json.Marshal(user)
	req = httptest.NewRequest(http.MethodPut, "/users/test-user", bytes.NewReader(body))
	rr = httptest.NewRecorder()
	app.handleUserDetail(store).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusOK, rr.Code)

	// 5. Delete User
	req = httptest.NewRequest(http.MethodDelete, "/users/test-user", nil)
	rr = httptest.NewRecorder()
	app.handleUserDetail(store).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusNoContent, rr.Code)

	// 6. Invalid Body for Create
	req = httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString("{"))
	rr = httptest.NewRecorder()
	app.handleUsers(store).ServeHTTP(rr, req)
	assert.Equal(t, http.StatusBadRequest, rr.Code)
}
