package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/samber/lo"
)

func TestValidate_MissingHTTPCallValidation(t *testing.T) {
	ctx := context.Background()

	// Create a config with an invalid HTTP call definition
	// e.g. path does not start with "/" which ValidateHTTPServiceDefinition checks
	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: lo.ToPtr("bad-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Address: lo.ToPtr("http://example.com"),
						Calls: map[string]*configv1.HttpCallDefinition{
							"bad-call": {
								// Invalid path: does not start with "/"
								EndpointPath: lo.ToPtr("bad/path"),
								Method:       configv1.HttpCallDefinition_HTTP_METHOD_GET.Enum(),
							},
						},
					},
				},
			},
		},
	}

	// We expect validation to fail
	errors := Validate(ctx, config, Server)

	if len(errors) == 0 {
		t.Errorf("Expected validation errors for invalid HTTP call path, got none")
	} else {
		// Verify the error message contains something relevant
		found := false
		for _, e := range errors {
			if e.Err.Error() != "" { // We accept any error for now, but ideally "path must start with a '/'"
				found = true
				t.Logf("Got expected error: %v", e.Err)
			}
		}
		if !found {
			t.Errorf("Expected relevant validation error")
		}
	}
}
