// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"context"
	"errors"
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

// MockProvider ...
// Summary: MockProvider
	name     string
	services []*configv1.UpstreamServiceConfig
	err      error
}

// Name ...
// Summary: Name
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.name
}

// Discover ...
// Summary: Discover
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	return m.services, m.err
}

// TestManager_Run ...
// Summary: TestManager_Run
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	manager := NewManager()

	// Provider 1: Success
	p1 := &MockProvider{
		name: "p1",
		services: []*configv1.UpstreamServiceConfig{
			configv1.UpstreamServiceConfig_builder{Name: pointer("s1")}.Build(),
		},
	}
	manager.RegisterProvider(p1)

	// Provider 2: Failure
	p2 := &MockProvider{
		name: "p2",
		err:  errors.New("failed"),
	}
	manager.RegisterProvider(p2)

	// Run discovery
	services := manager.Run(context.Background())

	// Check results
	assert.Len(t, services, 1)
	assert.Equal(t, "s1", services[0].GetName())

	// Check statuses
	status1, ok := manager.GetProviderStatus("p1")
	assert.True(t, ok)
	assert.Equal(t, "OK", status1.Status)
	assert.Equal(t, 1, status1.DiscoveredCount)
	assert.Empty(t, status1.LastError)
	assert.WithinDuration(t, time.Now(), status1.LastRunAt, time.Second)

	status2, ok := manager.GetProviderStatus("p2")
	assert.True(t, ok)
	assert.Equal(t, "ERROR", status2.Status)
	assert.Equal(t, "failed", status2.LastError)
	assert.WithinDuration(t, time.Now(), status2.LastRunAt, time.Second)
}

func pointer(s string) *string {
	return &s
}

// TestManager_GetStatuses ...
// Summary: TestManager_GetStatuses
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	manager := NewManager()

	p1 := &MockProvider{name: "p1"}
	manager.RegisterProvider(p1)

	statuses := manager.GetStatuses()
	assert.Len(t, statuses, 1)
	assert.Equal(t, "p1", statuses[0].Name)
	assert.Equal(t, "PENDING", statuses[0].Status)
}
