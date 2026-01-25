// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"google.golang.org/protobuf/proto"
)

func TestServer_SecretLeakage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := tool.NewMockManagerInterface(ctrl)

	// Create a service with a secret
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("secret-service"),
		Id:   proto.String("secret-service"),
		UpstreamAuth: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					Value: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{
							PlainText: "SUPER_SECRET_KEY",
						},
					},
				},
			},
		},
	}

	sr := &MockServiceRegistry{
		services: []*configv1.UpstreamServiceConfig{svc},
	}

	s := NewServer(nil, tm, sr, nil, nil, nil)
	ctx := context.Background()

	// ListServices
	listResp, err := s.ListServices(ctx, &pb.ListServicesRequest{})
	require.NoError(t, err)
	require.Len(t, listResp.Services, 1)

	// Check if secret is leaked
	auth := listResp.Services[0].GetUpstreamAuth()
	require.NotNil(t, auth)
	apiKey := auth.GetApiKey()
	require.NotNil(t, apiKey)

	// This confirms the vulnerability is fixed
	assert.Empty(t, apiKey.GetValue().GetPlainText(), "Secret should be redacted")
}
