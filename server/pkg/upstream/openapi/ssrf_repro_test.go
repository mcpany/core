package openapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/proto"
)

func TestOpenAPIUpstream_Register_SSRF(t *testing.T) {
	// Ensure no env vars allow loopback
	os.Unsetenv("MCPANY_ALLOW_LOOPBACK_RESOURCES")
	os.Unsetenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS")

	ctx := context.Background()
	mockToolManager := new(MockToolManager)
	upstream := NewOpenAPIUpstream()

	// Start a local server (loopback)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, sampleOpenAPISpecJSONForCacheTest)
	}))
	defer ts.Close()

	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-ssrf"),
		OpenapiService: configv1.OpenapiUpstreamService_builder{
			SpecUrl: proto.String(ts.URL),
		}.Build(),
	}.Build()

	expectedKey, _ := util.SanitizeServiceName("test-service-ssrf")
	mockToolManager.On("AddServiceInfo", expectedKey, mock.Anything).Return()
    // Mock other calls that might happen if it succeeds (vulnerable case)
	mockToolManager.On("GetTool", mock.Anything).Return(nil, false).Maybe()
	mockToolManager.On("AddTool", mock.Anything).Return(nil).Maybe()

	// Sentinel Security: We expect this to FAIL with a security error when fixed.
	// Currently (Vulnerable), it will SUCCEED.
	_, _, _, err := upstream.Register(ctx, config, mockToolManager, nil, nil, false)

	// We assert that it FAILS with SSRF block.
	// If the code is vulnerable, 'err' will be nil, and this assertion will fail.
	assert.Error(t, err, "Registration should fail due to SSRF protection")
	// Note: The specific error message "ssrf attempt blocked" is logged but swallowed by Register.
	// The returned error is generic: "OpenAPI spec content is missing or failed to load..."
}
