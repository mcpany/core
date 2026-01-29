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
)

func TestServer_SecretLeakage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tm := tool.NewMockManagerInterface(ctrl)

	// Create a service with a secret
	secretVal := &configv1.SecretValue{}
	secretVal.SetPlainText("SUPER_SECRET_KEY")

	apiKeyAuth := &configv1.APIKeyAuth{}
	apiKeyAuth.SetValue(secretVal)

	auth := &configv1.Authentication{}
	auth.SetApiKey(apiKeyAuth)

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetName("secret-service")
	svc.SetId("secret-service")
	svc.SetUpstreamAuth(auth)

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
	gotAuth := listResp.Services[0].GetUpstreamAuth()
	require.NotNil(t, gotAuth)
	apiKey := gotAuth.GetApiKey()
	require.NotNil(t, apiKey)

	// This confirms the vulnerability is fixed
	assert.Empty(t, apiKey.GetValue().GetPlainText(), "Secret should be redacted")
}
