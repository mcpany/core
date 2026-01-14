// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package diagnostics

import (
	"context"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/appconsts"
	"github.com/mcpany/core/server/pkg/serviceregistry"
	"github.com/stretchr/testify/assert"
)

func TestService_GenerateReport(t *testing.T) {
	// Setup
	registry := serviceregistry.New(nil, nil, nil, nil, nil)
	svc := NewService(registry)
	appconsts.Version = "v1.2.3" // Mock version

	// Execute
	report, err := svc.GenerateReport(context.Background())

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "v1.2.3", report.System.Version)
	assert.NotEmpty(t, report.System.GoVersion)
	assert.NotEmpty(t, report.System.Uptime)
	assert.Greater(t, report.System.UptimeSeconds, 0.0)
	assert.NotNil(t, report.Services) // Should be empty slice or nil, but checked as slice usually

	// Sleep briefly to ensure uptime increases (though not strictly necessary for unit test)
	time.Sleep(10 * time.Millisecond)
	report2, _ := svc.GenerateReport(context.Background())
	assert.Greater(t, report2.System.UptimeSeconds, report.System.UptimeSeconds)
}
