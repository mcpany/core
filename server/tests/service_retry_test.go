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
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// TestServiceRetry ...
// Summary: TestServiceRetry
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
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
// MockStorage implements storage.Storage for testing
// Summary: MockStorage
	mock.Mock
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
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.McpAnyServerConfig), args.Error(1)
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
	return true
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
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*configv1.UpstreamServiceConfig), args.Error(1)
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
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*configv1.GlobalSettings), args.Error(1)
}

// Other interface methods - stubbed to panic if called (unexpected) or return nil error
// Other interface methods - stubbed to panic if called (unexpected) or return nil error
// Summary: Watch
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return nil, nil // Not used in this test
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
	panic("unexpected call to GetService")
}
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
	return nil
}
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
	return nil
}
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
	args := m.Called()
	return args.Error(0)
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
