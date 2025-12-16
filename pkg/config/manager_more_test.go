package config

import (
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamServiceManager_ApplyAuthentication_Error(t *testing.T) {
	mgr := &UpstreamServiceManager{}
	req := httptest.NewRequest("GET", "http://example.com", nil)

	// Auth with non-existent env var
	auth := &configv1.UpstreamAuthentication{
		AuthMethod: &configv1.UpstreamAuthentication_ApiKey{
			ApiKey: &configv1.UpstreamAPIKeyAuth{
				ApiKey: &configv1.SecretValue{
					Value: &configv1.SecretValue_EnvironmentVariable{
						EnvironmentVariable: "NON_EXISTENT_VAR_FOR_TESTING",
					},
				},
				HeaderName: proto.String("X-API-Key"),
			},
		},
	}

	err := mgr.applyAuthentication(req, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable \"NON_EXISTENT_VAR_FOR_TESTING\" is not set")
}
