// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/mcpany/core/server/pkg/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	healthlib "github.com/alexliesenfeld/health"
	"github.com/samber/lo"
)

// TestAutoHealthFallback validates that services without explicit health checks
// are still checked for connectivity and reported as down if unreachable.
func TestAutoHealthFallback(t *testing.T) {
	ctx := context.Background()

	// 1. Define a service pointing to a closed port (likely unreachable)
	// We use a random high port that is unlikely to be in use.
	unreachableAddr := "http://localhost:59999"

	config := &configv1.UpstreamServiceConfig{
		Name: lo.ToPtr("auto-health-test-service"),
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Address: &unreachableAddr,
				// No HealthCheck configured
			},
		},
	}

	// 2. Create the checker
	checker := health.NewChecker(config)
	require.NotNil(t, checker, "Checker should be created")

	// 3. Perform the check
	// The check should fail because the address is unreachable
	result := checker.Check(ctx)

	// 4. Assert status is DOWN
	assert.Equal(t, healthlib.StatusDown, result.Status, "Service should be reported as DOWN")
}
