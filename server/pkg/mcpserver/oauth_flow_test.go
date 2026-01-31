// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package mcpserver

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	v1 "github.com/mcpany/core/proto/api/v1"
	bus_pb "github.com/mcpany/core/proto/bus"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/auth"
	"github.com/mcpany/core/server/pkg/bus"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestOAuthFlow_Complete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Setup Mock OAuth Server
	mockOAuth := auth.NewMockOAuth2Server(t)
	defer mockOAuth.Close()

	// 2. Setup Components
	memStore := memory.NewStore()
	authManager := auth.NewManager()
	authManager.SetStorage(memStore)

	// Setup user in context
	userID := "test-user-123"
	ctx = auth.ContextWithUser(ctx, userID)

	messageBus := bus_pb.MessageBus_builder{}.Build()
	messageBus.SetInMemory(bus_pb.InMemoryBus_builder{}.Build())
	busProvider, err := bus.NewProvider(messageBus)
	require.NoError(t, err)

	regServer, err := NewRegistrationServer(busProvider, authManager)
	require.NoError(t, err)

	// 3. Create a Credential with OAuth Config
	credID := uuid.New().String()

	// Create Authentication using builders.
	oauthConfig := configv1.OAuth2Auth_builder{
		ClientId: configv1.SecretValue_builder{
			PlainText: proto.String("client-id"),
		}.Build(),
		ClientSecret: configv1.SecretValue_builder{
			PlainText: proto.String("client-secret"),
		}.Build(),
		AuthorizationUrl: proto.String(mockOAuth.URL + "/auth"),
		TokenUrl:         proto.String(mockOAuth.URL + "/token"),
		Scopes:           proto.String("read write"),
	}.Build()

	authConfig := configv1.Authentication_builder{
		Oauth2: oauthConfig,
	}.Build()

	cred := configv1.Credential_builder{
		Id:             proto.String(credID),
		Name:           proto.String("test-oauth-cred"),
		Authentication: authConfig,
	}.Build()

	err = memStore.SaveCredential(ctx, cred)
	require.NoError(t, err)

	// 4. Initiate OAuth Flow via RegistrationServer
	initReq := v1.InitiateOAuth2FlowRequest_builder{
		CredentialId: credID,
		RedirectUrl:  "http://127.0.0.1:3000/callback",
	}.Build()

	initResp, err := regServer.InitiateOAuth2Flow(ctx, initReq)
	require.NoError(t, err)
	require.NotEmpty(t, initResp.GetAuthorizationUrl())
	require.NotEmpty(t, initResp.GetState())
	assert.Contains(t, initResp.GetAuthorizationUrl(), mockOAuth.URL+"/auth")

	// Parse URL to check state
	u, err := url.Parse(initResp.GetAuthorizationUrl())
	require.NoError(t, err)
	assert.Equal(t, initResp.GetState(), u.Query().Get("state"))

	// 5. Simulate Callback (skip actual HTTP redirection)
	// We manually call AuthManager.HandleOAuthCallback with a mock code
	mockCode := "mock_auth_code"
	err = authManager.HandleOAuthCallback(ctx, userID, "", credID, mockCode, "http://127.0.0.1:3000/callback")
	require.NoError(t, err)

	// 6. Verify Token was saved to Credential
	updatedCred, err := memStore.GetCredential(ctx, credID)
	require.NoError(t, err)
	require.NotNil(t, updatedCred.GetToken())
	assert.Equal(t, "Bearer", updatedCred.GetToken().GetTokenType())
}
