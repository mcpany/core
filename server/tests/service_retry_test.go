// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tests

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestServiceRetry(t *testing.T) {
	// Get two separate ephemeral ports to avoid application/mock collisions
	// Use 127.0.0.2 to avoid collisions with other local services and test SSRF bypass
	l1, err := util.ListenWithRetry(context.Background(), "tcp", "127.0.0.2:0")
	require.NoError(t, err)
	targetPort := l1.Addr().(*net.TCPAddr).Port
	l1.Close()

	l2, err := util.ListenWithRetry(context.Background(), "tcp", "127.0.0.2:0")
	require.NoError(t, err)
	appPort := l2.Addr().(*net.TCPAddr).Port
	l2.Close()

	targetURL := fmt.Sprintf("http://127.0.0.2:%d/mcp", targetPort)

	// 2. Start the Application with MockStorage
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create config object
	httpConn := configv1.McpStreamableHttpConnection_builder{
		HttpAddress: proto.String(targetURL),
	}.Build()

	mcpSvc := configv1.McpUpstreamService_builder{
		HttpConnection: httpConn,
	}.Build()

	resilience := configv1.ResilienceConfig_builder{
		Timeout: durationpb.New(500 * time.Millisecond),
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Name:       proto.String("delayed-mcp"),
		McpService: mcpSvc,
		Resilience: resilience,
	}.Build()

	config := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
	}.Build()

	mockStore := new(MockStorage)
	mockStore.On("Load", mock.Anything).Return(config, nil)
	mockStore.On("ListServices", mock.Anything).Return([]*configv1.UpstreamServiceConfig{}, nil)
	mockStore.On("GetGlobalSettings", mock.Anything).Return(&configv1.GlobalSettings{}, nil)
	mockStore.On("Close").Return(nil)

	a := app.NewApplication()
	a.Storage = mockStore

	go func() {
		// Empty config paths as we supply config via Storage
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              afero.NewMemMapFs(),
			Stdio:           false,
			JSONRPCPort:     fmt.Sprintf("127.0.0.2:%d", appPort),
			GRPCPort:        "",
			ConfigPaths:     nil,
			APIKey:          "",
			ShutdownTimeout: 1 * time.Second,
		}
		err := a.Run(opts)
		if err != nil && ctx.Err() == nil {
			t.Logf("Application run error: %v", err)
		}
	}()

	// Wait for app to start
	err = a.WaitForStartup(ctx)
	if err != nil {
		t.Fatalf("Failed to wait for startup: %v", err)
	}

	// Verify service failed to register
	require.Eventually(t, func() bool {
		if a.ServiceRegistry == nil {
			return false
		}
		_, hasError := a.ServiceRegistry.GetServiceError("delayed-mcp")
		return hasError
	}, 30*time.Second, 100*time.Millisecond, "ServiceRegistry not initialized or service did not fail as expected")

	errStr, hasError := a.ServiceRegistry.GetServiceError("delayed-mcp")
	t.Logf("Initial Service Error: %s (hasError: %v)", errStr, hasError)

	if !hasError {
		_, infoOk := a.ServiceRegistry.GetServiceInfo("delayed-mcp")
		if infoOk {
			t.Log("Service registered successfully unexpectedly!")
		}
	} else {
		t.Log("Service correctly failed to register initially.")
	}

	// 3. Start the mock MCP service
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"jsonrpc": "2.0", "id": 1, "result": {"protocolVersion": "2024-11-05", "capabilities": {}, "serverInfo": {"name": "mock", "version": "1.0"}}}`))
	}))

	var l3 net.Listener
	require.Eventually(t, func() bool {
		l3, err = util.ListenWithRetry(context.Background(), "tcp", fmt.Sprintf("127.0.0.2:%d", targetPort))
		if err != nil {
			t.Logf("Failed to re-bind to port 127.0.0.2:%d: %v. Retrying...", targetPort, err)
			return false
		}
		return true
	}, 10*time.Second, 100*time.Millisecond, "Failed to re-bind to port 127.0.0.2:%d after retries", targetPort)

	ts.Listener = l3
	ts.Start()
	defer ts.Close()

	t.Logf("Started mock service at %s", targetURL)

	// 4. Wait and see if it recovers
	t.Log("Waiting for retry...")

	// Check if service is now healthy
	require.Eventually(t, func() bool {
		_, hasError := a.ServiceRegistry.GetServiceError("delayed-mcp")
		return !hasError
	}, 30*time.Second, 500*time.Millisecond, "Service failed to recover within timeout")

	t.Log("Service recovered successfully!")
}

// MockStorage implements storage.Storage for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Load(ctx context.Context) (*configv1.McpAnyServerConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.McpAnyServerConfig), args.Error(1)
}

func (m *MockStorage) HasConfigSources() bool {
	return true
}

func (m *MockStorage) ListServices(ctx context.Context) ([]*configv1.UpstreamServiceConfig, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
}

func (m *MockStorage) GetGlobalSettings(ctx context.Context) (*configv1.GlobalSettings, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.GlobalSettings), args.Error(1)
}

// Other interface methods - stubbed to panic if called (unexpected) or return nil error
func (m *MockStorage) Watch(ctx context.Context) (<-chan *configv1.McpAnyServerConfig, error) {
	return nil, nil // Not used in this test
}
func (m *MockStorage) GetService(ctx context.Context, name string) (*configv1.UpstreamServiceConfig, error) {
	panic("unexpected call to GetService")
}
func (m *MockStorage) SaveService(ctx context.Context, service *configv1.UpstreamServiceConfig) error {
	return nil
}
func (m *MockStorage) DeleteService(ctx context.Context, name string) error {
	return nil
}
func (m *MockStorage) ListSecrets(ctx context.Context) ([]*configv1.Secret, error) {
	return nil, nil
}
func (m *MockStorage) GetSecret(ctx context.Context, id string) (*configv1.Secret, error) {
	return nil, nil
}
func (m *MockStorage) SaveSecret(ctx context.Context, secret *configv1.Secret) error {
	return nil
}
func (m *MockStorage) DeleteSecret(ctx context.Context, id string) error {
	return nil
}
func (m *MockStorage) CreateUser(ctx context.Context, user *configv1.User) error { return nil }
func (m *MockStorage) GetUser(ctx context.Context, id string) (*configv1.User, error) {
	return nil, nil
}
func (m *MockStorage) ListUsers(ctx context.Context) ([]*configv1.User, error)   { return nil, nil }
func (m *MockStorage) UpdateUser(ctx context.Context, user *configv1.User) error { return nil }
func (m *MockStorage) DeleteUser(ctx context.Context, id string) error           { return nil }
func (m *MockStorage) ListProfiles(ctx context.Context) ([]*configv1.ProfileDefinition, error) {
	return nil, nil
}
func (m *MockStorage) GetProfile(ctx context.Context, name string) (*configv1.ProfileDefinition, error) {
	return nil, nil
}
func (m *MockStorage) SaveProfile(ctx context.Context, profile *configv1.ProfileDefinition) error {
	return nil
}
func (m *MockStorage) DeleteProfile(ctx context.Context, name string) error { return nil }
func (m *MockStorage) ListServiceCollections(ctx context.Context) ([]*configv1.Collection, error) {
	return nil, nil
}
func (m *MockStorage) GetServiceCollection(ctx context.Context, name string) (*configv1.Collection, error) {
	return nil, nil
}
func (m *MockStorage) SaveServiceCollection(ctx context.Context, collection *configv1.Collection) error {
	return nil
}
func (m *MockStorage) DeleteServiceCollection(ctx context.Context, name string) error { return nil }
func (m *MockStorage) SaveToken(ctx context.Context, token *configv1.UserToken) error { return nil }
func (m *MockStorage) GetToken(ctx context.Context, userID, serviceID string) (*configv1.UserToken, error) {
	return nil, nil
}
func (m *MockStorage) DeleteToken(ctx context.Context, userID, serviceID string) error { return nil }
func (m *MockStorage) ListCredentials(ctx context.Context) ([]*configv1.Credential, error) {
	return nil, nil
}
func (m *MockStorage) GetCredential(ctx context.Context, id string) (*configv1.Credential, error) {
	return nil, nil
}
func (m *MockStorage) SaveCredential(ctx context.Context, cred *configv1.Credential) error {
	return nil
}
func (m *MockStorage) DeleteCredential(ctx context.Context, id string) error { return nil }
func (m *MockStorage) GetDashboardLayout(ctx context.Context, userID string) (string, error) {
	return "", nil
}
func (m *MockStorage) SaveDashboardLayout(ctx context.Context, userID string, layoutJSON string) error {
	return nil
}
func (m *MockStorage) SaveGlobalSettings(ctx context.Context, settings *configv1.GlobalSettings) error {
	return nil
}
func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}
