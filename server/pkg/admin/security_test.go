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
	secretVal := configv1.SecretValue_builder{
		PlainText: proto.String("SUPER_SECRET_KEY"),
	}.Build()

	apiKeyAuth := configv1.APIKeyAuth_builder{
		Value: secretVal,
	}.Build()

	auth := configv1.Authentication_builder{
		ApiKey: apiKeyAuth,
	}.Build()

	svc := configv1.UpstreamServiceConfig_builder{
		Name:         proto.String("secret-service"),
		Id:           proto.String("secret-service"),
		UpstreamAuth: auth,
	}.Build()

	sr := &MockServiceRegistry{
		services: []*configv1.UpstreamServiceConfig{svc},
	}

	s := NewServer(nil, tm, sr, nil, nil, nil)
	ctx := context.Background()

	// ListServices
	listResp, err := s.ListServices(ctx, &pb.ListServicesRequest{})
	require.NoError(t, err)
	require.Len(t, listResp.GetServices(), 1)

	// Check if secret is leaked
	gotAuth := listResp.GetServices()[0].GetUpstreamAuth()
	require.NotNil(t, gotAuth)
	apiKey := gotAuth.GetApiKey()
	require.NotNil(t, apiKey)

	// This confirms the vulnerability is fixed
	assert.Empty(t, apiKey.GetValue().GetPlainText(), "Secret should be redacted")
}
