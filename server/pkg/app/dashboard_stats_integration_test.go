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
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/mcpany/core/server/pkg/topology"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

// MockServiceRegistryForDashboard is a mock implementation of ServiceRegistryInterface.
// MockServiceRegistryForDashboard is a mock implementation of ServiceRegistryInterface.
// Summary: MockServiceRegistryForDashboard
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
	return args.String(0), nil, nil, args.Error(3)
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

// TestHandleDashboardHealth_Integration ...
// Summary: TestHandleDashboardHealth_Integration
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	// Setup TopologyManager
	tm := topology.NewManager(nil, nil)
	defer tm.Close()

	// Seed TopologyManager
	serviceID := "test-service-id"

	// Record some activity (200ms latency)
	tm.RecordActivity("session-1", nil, 200*time.Millisecond, false, serviceID, 0)

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	mockRegistry := new(MockServiceRegistryForDashboard)

	svc := configv1.UpstreamServiceConfig_builder{
		Id:   proto.String(serviceID),
		Name: proto.String("test-service"),
	}.Build()

	mockRegistry.On("GetAllServices").Return([]*configv1.UpstreamServiceConfig{svc}, nil)
	mockRegistry.On("GetServiceError", serviceID).Return("", false)

	app := &Application{
		ServiceRegistry: mockRegistry,
		TopologyManager: tm,
	}

	req := httptest.NewRequest(http.MethodGet, "/dashboard/health", nil)
	w := httptest.NewRecorder()

	handler := app.handleDashboardHealth()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp ServiceHealthResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Len(t, resp.Services, 1)
	svcResp := resp.Services[0]
	assert.Equal(t, serviceID, svcResp.ID)
	// Latency should be "200ms"
	assert.Equal(t, "200ms", svcResp.Latency)

	// Uptime will likely be "0.0%" or "Unknown" because history is empty.
	assert.Equal(t, "Unknown", svcResp.Uptime)
}
