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

	// Create Authentication using builders or structs.
	// Using standard struct pointers since builders caused issues.
	oauthConfig := &configv1.OAuth2Auth{
		ClientId: &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "client-id"},
		},
		ClientSecret: &configv1.SecretValue{
			Value: &configv1.SecretValue_PlainText{PlainText: "client-secret"},
		},
		AuthorizationUrl: proto.String(mockOAuth.URL + "/auth"),
		TokenUrl:         proto.String(mockOAuth.URL + "/token"),
		Scopes:           proto.String("read write"),
	}

	authConfig := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: oauthConfig,
		},
	}

	cred := &configv1.Credential{
		Id:             proto.String(credID),
		Name:           proto.String("test-oauth-cred"),
		Authentication: authConfig,
	}

	err = memStore.SaveCredential(ctx, cred)
	require.NoError(t, err)

	// 4. Initiate OAuth Flow via RegistrationServer
	initReq := &v1.InitiateOAuth2FlowRequest{
		CredentialId: credID,
		RedirectUrl:  "http://localhost:3000/callback",
	}

	initResp, err := regServer.InitiateOAuth2Flow(ctx, initReq)
	require.NoError(t, err)
	require.NotEmpty(t, initResp.AuthorizationUrl)
	require.NotEmpty(t, initResp.State)
	assert.Contains(t, initResp.AuthorizationUrl, mockOAuth.URL+"/auth")

	// Parse URL to check state
	u, err := url.Parse(initResp.AuthorizationUrl)
	require.NoError(t, err)
	assert.Equal(t, initResp.State, u.Query().Get("state"))

	// 5. Simulate Callback (skip actual HTTP redirection)
	// We manually call AuthManager.HandleOAuthCallback with a mock code
	mockCode := "mock_auth_code"
	err = authManager.HandleOAuthCallback(ctx, userID, "", credID, mockCode, "http://localhost:3000/callback")
	require.NoError(t, err)

	// 6. Verify Token was saved to Credential
	updatedCred, err := memStore.GetCredential(ctx, credID)
	require.NoError(t, err)
	require.NotNil(t, updatedCred.Token)
	assert.Equal(t, "Bearer", updatedCred.Token.GetTokenType())
}
