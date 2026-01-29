package http

import (
	"context"
	"testing"

	"github.com/mcpany/core/server/pkg/pool"
	"github.com/stretchr/testify/assert"
)

func TestHTTPUpstream_CheckHealth_BeforeRegister(t *testing.T) {
	pm := pool.NewManager()
	upstream := NewUpstream(pm)

	type HealthChecker interface {
		CheckHealth(ctx context.Context) error
	}

	hc, ok := upstream.(HealthChecker)
	assert.True(t, ok)

	err := hc.CheckHealth(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no address configured")
}
