// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestBlindFileReadMitigation(t *testing.T) {
	// Create a file in CWD
	fileName := "secret_test.txt"

	err := os.WriteFile(fileName, []byte("super_secret_password"), 0600)
	assert.NoError(t, err)
	defer os.Remove(fileName)

	// Even if we use a regex that MATCHES (e.g. "^s.*"), it should fail because
	// we disallowed regex validation on file paths completely.
	secret := configv1.SecretValue_builder{
		FilePath:        proto.String(fileName),
		ValidationRegex: proto.String("^s.*"),
	}.Build()

	auth := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("x-api-key"),
			Value:     secret,
		}.Build(),
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("malicious_service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://localhost:8080"),
		}.Build(),
		UpstreamAuth: auth,
	}.Build()

	cfg := configv1.McpAnyServerConfig_builder{
		UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
	}.Build()

	errs := Validate(context.Background(), cfg, Server)

    foundMitigationError := false
    for _, e := range errs {
        t.Logf("Error: %v", e.Err)
        if e.Err != nil && strings.Contains(e.Err.Error(), "validation regex is not supported for secret file paths") {
            foundMitigationError = true
        }
    }

    if !foundMitigationError {
        t.Errorf("Expected mitigation error 'validation regex is not supported for secret file paths' but did not find it.")
    } else {
        t.Log("SUCCESS: Blind File Read attempt was blocked.")
    }
}
