package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestHandleDashboardHealth(t *testing.T) {
	// Setup
	app := NewApplication()
	store := memory.NewStore()
	app.ServiceRegistry = nil // Use store fallback for this test

	// Create a service
	svc := &configv1.UpstreamServiceConfig{
		Name:    proto.String("test-service"),
		Id:      proto.String("test-id"),
		Disable: proto.Bool(false),
	}
	err := store.SaveService(context.Background(), svc)
	require.NoError(t, err)

	// Create disabled service
	disabledSvc := &configv1.UpstreamServiceConfig{
		Name:    proto.String("disabled-service"),
		Id:      proto.String("disabled-id"),
		Disable: proto.Bool(true),
	}
	err = store.SaveService(context.Background(), disabledSvc)
	require.NoError(t, err)

	// Add active service to ToolManager to simulate "healthy" state
	// Note: We need to ensure ToolManager is initialized (NewApplication does it)
	app.ToolManager.AddServiceInfo("test-id", &tool.ServiceInfo{
		Name:         "test-service",
		Config:       svc,
		HealthStatus: "healthy",
	})

	// Create handler
	handler := app.handleDashboardHealth(store)

	// Execute
	req := httptest.NewRequest("GET", "/dashboard/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// Verify
	assert.Equal(t, http.StatusOK, w.Code)

	var healths []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Status  string `json:"status"`
		Message string `json:"message"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &healths)
	require.NoError(t, err)

	assert.Len(t, healths, 2)

	// Verify test-service
	var s1 *struct{ ID, Name, Status, Message string }
	for i := range healths {
		if healths[i].Name == "test-service" {
			val := struct{ ID, Name, Status, Message string }{
				healths[i].ID, healths[i].Name, healths[i].Status, healths[i].Message,
			}
			s1 = &val
			break
		}
	}
	require.NotNil(t, s1)
	assert.Equal(t, "healthy", s1.Status)

	// Verify disabled-service
	var s2 *struct{ ID, Name, Status, Message string }
	for i := range healths {
		if healths[i].Name == "disabled-service" {
			val := struct{ ID, Name, Status, Message string }{
				healths[i].ID, healths[i].Name, healths[i].Status, healths[i].Message,
			}
			s2 = &val
			break
		}
	}
	require.NotNil(t, s2)
	assert.Equal(t, "inactive", s2.Status)
}
