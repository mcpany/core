package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/alexliesenfeld/health"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestHTTPClientWrapper(t *testing.T) {
	client := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}
	checker := &mockChecker{status: health.StatusUp}

	wrapper := NewHTTPClientWrapper(client, config, checker)

	assert.True(t, wrapper.IsHealthy(context.Background()))
	assert.NoError(t, wrapper.Close())
}

func TestHTTPClientWrapper_NoChecker(t *testing.T) {
	client := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}

	wrapper := &HTTPClientWrapper{
	    Client: client,
	    config: config,
	    checker: nil,
	}
	assert.True(t, wrapper.IsHealthy(context.Background()))
}

func TestHTTPClientWrapper_CheckerDown(t *testing.T) {
	client := &http.Client{}
	config := &configv1.UpstreamServiceConfig{}
	checker := &mockChecker{status: health.StatusDown}

	wrapper := NewHTTPClientWrapper(client, config, checker)
	assert.False(t, wrapper.IsHealthy(context.Background()))
}
