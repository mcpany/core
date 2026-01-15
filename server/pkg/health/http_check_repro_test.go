package health

import (
	"context"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

func TestHTTPCheck_ReproSilentFailure(t *testing.T) {
	ctx := context.Background()

	// Use an address that is guaranteed to fail connection.
	// localhost:1 is likely not listening.
	addr := "localhost:1"

	upstreamConfig := configv1.UpstreamServiceConfig_builder{
		Name:        lo.ToPtr("broken-service"),
		HttpService: configv1.HttpUpstreamService_builder{Address: &addr}.Build(),
	}.Build()

	checker := NewChecker(upstreamConfig)
	assert.NotNil(t, checker)

	// FIXED BEHAVIOR: Should return StatusDown because the address is unreachable and we now check it.
	result := checker.Check(ctx)
	assert.Equal(t, health.StatusDown, result.Status, "Should be Down because address is unreachable")
}
