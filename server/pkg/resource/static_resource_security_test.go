// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package resource

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestStaticResource_SSRFProtection(t *testing.T) {
	// Ensure loopback is BLOCKED (default behavior)
	t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "")
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("secret data"))
	}))
	defer server.Close()

	def := configv1.ResourceDefinition_builder{
		Uri:  proto.String(server.URL),
		Name: proto.String("Secret Resource"),
	}.Build()

	r := NewStaticResource(def, "test-service")

	_, err := r.Read(context.Background())
	require.Error(t, err)
	// The error message comes from SafeDialer in util/net.go
	assert.Contains(t, err.Error(), "ssrf attempt blocked")
}
