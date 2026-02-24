// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/discovery"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

// MockProvider for testing
type MockProvider struct {
	name string
}

func (m *MockProvider) Name() string { return m.name }
func (m *MockProvider) Discover(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	time.Sleep(10 * time.Millisecond) // Simulate work
	return []*configv1.UpstreamServiceConfig{
		configv1.UpstreamServiceConfig_builder{Name: proto.String("mock-service")}.Build(),
	}, nil
}

func TestHandleDiscoveryStatus(t *testing.T) {
	app := NewApplication()
	app.DiscoveryManager = discovery.NewManager()
	mockProvider := &MockProvider{name: "mock-provider"}
	app.DiscoveryManager.RegisterProvider(mockProvider)

	// Run discovery once to populate status
	app.DiscoveryManager.Run(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/discovery/status", nil)
	w := httptest.NewRecorder()

	app.handleDiscoveryStatus(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var statuses []*discovery.ProviderStatus
	err := json.Unmarshal(w.Body.Bytes(), &statuses)
	require.NoError(t, err)
	require.Len(t, statuses, 1)
	assert.Equal(t, "mock-provider", statuses[0].Name)
	assert.Equal(t, "OK", statuses[0].Status)
	assert.Equal(t, 1, statuses[0].DiscoveredCount)
}

func TestHandleDiscoveryTrigger(t *testing.T) {
	app := NewApplication()
	app.DiscoveryManager = discovery.NewManager()
	mockProvider := &MockProvider{name: "mock-provider"}
	app.DiscoveryManager.RegisterProvider(mockProvider)

	// We need a mock ServiceRegistry to verify registration
	// But ServiceRegistry in Application is an interface.
	// For this test, we can just ensure it doesn't panic if nil, or verify logic if we mock it.
	// The implementation checks if a.ServiceRegistry != nil.
	// Let's leave it nil for now to test safe execution, or use MockServiceRegistry from api_test.go if available.
	// MockServiceRegistry is in api_test.go but typically test files in same package share types.

	req := httptest.NewRequest(http.MethodPost, "/discovery/trigger", nil)
	w := httptest.NewRecorder()

	app.handleDiscoveryTrigger(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	// Wait for async goroutine (flaky without synchronization, but for unit test we assume it starts)
	// In a real test we might want to wait or use a channel, but here we just check immediate response.
}
