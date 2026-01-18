// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gogo/protobuf/proto"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestUserManagement(t *testing.T) {
	// Enable file config for this test
	os.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")
	defer os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")

	// Mock OIDC Server (Start FIRST so we can use URL in config)
	mockOIDCMux := http.NewServeMux()
	// Use TLS to satisfy OIDC requirements (or we'd need to assume insecure which is harder to inject globally without flag)
	// We will inject the trusted client via context.
	mockOIDC := httptest.NewTLSServer(mockOIDCMux)
	defer mockOIDC.Close()

	mockOIDCMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, fmt.Sprintf(`{
			"issuer": "%s",
			"authorization_endpoint": "%s/authorize",
			"token_endpoint": "%s/token",
			"jwks_uri": "%s/keys",
			"response_types_supported": ["code"],
			"subject_types_supported": ["public"],
			"id_token_signing_alg_values_supported": ["RS256"]
		}`, mockOIDC.URL, mockOIDC.URL, mockOIDC.URL, mockOIDC.URL))
	})

	// Just return 200 for keys to satisfy startup check if it fetches them
	mockOIDCMux.HandleFunc("/keys", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"keys": []}`)
	})

	// Setup temporary DB path
	dbPath := t.TempDir() + "/mcpany_test.db"

	configContent := fmt.Sprintf(`
global_settings:
    db_driver: "sqlite"
    db_path: "%s"
    api_key: "test-api-key"
    oidc:
        issuer: "%s"
        client_id: "test-client"
        client_secret: "test-secret"
        redirect_url: "http://localhost:8080/callback"
`, dbPath, mockOIDC.URL)

	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	var jsonrpcPort int
	var grpcRegPort int
	var ctx context.Context
	var cancel context.CancelFunc
	var appRunner *app.Application

	// Prepare Context with Trusted OIDC Client
	baseCtx := context.Background()
	oidcClient := mockOIDC.Client()
	baseCtx = oidc.ClientContext(baseCtx, oidcClient)

	// Retry loop for ports
	for attempt := 0; attempt < 3; attempt++ {
		jsonrpcPort = integration.FindFreePort(t)
		grpcRegPort = integration.FindFreePort(t)
		for grpcRegPort == jsonrpcPort {
			grpcRegPort = integration.FindFreePort(t)
		}

		ctx, cancel = context.WithCancel(baseCtx)
		appRunner = app.NewApplication()

		errChan := make(chan error, 1)
		go func() {
			jsonrpcAddress := fmt.Sprintf(":%d", jsonrpcPort)
			grpcRegAddress := fmt.Sprintf(":%d", grpcRegPort)
			// Mock filesystem with our config
			fs := afero.NewOsFs()
			err := appRunner.Run(ctx, fs, false, jsonrpcAddress, grpcRegAddress, []string{tmpFile.Name()}, "", 5*time.Second)
			if err != nil && err != context.Canceled {
				errChan <- err
			}
		}()

		select {
		case err := <-errChan:
			cancel()
			if err != nil && (strings.Contains(err.Error(), "address already in use") || strings.Contains(err.Error(), "bind")) {
				t.Logf("Port conflict detected, retrying...")
				continue
			}
			t.Fatalf("Server failed to start: %v", err)
		case <-time.After(500 * time.Millisecond):
			goto ServerStarted
		}
	}
	t.Fatal("Failed to start server after multiple attempts")

ServerStarted:
	defer cancel()

	// Wait for health check
	httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
	integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

	// Connect to gRPC Admin Service
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcRegPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	adminClient := pb_admin.NewAdminServiceClient(conn)

	// Create initial user
	user1 := &configv1.User{
		Id: proto.String("user-1"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_ApiKey{
				ApiKey: &configv1.APIKeyAuth{
					VerificationValue: proto.String("secret-key"),
				},
			},
		},
		ProfileIds: []string{"profile-1"},
		Roles:      []string{"viewer"},
	}

	createResp, err := adminClient.CreateUser(ctx, &pb_admin.CreateUserRequest{User: user1})
	require.NoError(t, err)
	require.Equal(t, "user-1", createResp.User.GetId())
	require.Equal(t, "secret-key", createResp.User.GetAuthentication().GetApiKey().GetVerificationValue())

	// Get user
	getResp, err := adminClient.GetUser(ctx, &pb_admin.GetUserRequest{UserId: proto.String("user-1")})
	require.NoError(t, err)
	require.Equal(t, "user-1", getResp.User.GetId())

	// List users
	listResp, err := adminClient.ListUsers(ctx, &pb_admin.ListUsersRequest{})
	require.NoError(t, err)
	require.Len(t, listResp.Users, 1)
	require.Equal(t, "user-1", listResp.Users[0].GetId())

	// Update user
	user1.Roles = []string{"admin"}
	updateResp, err := adminClient.UpdateUser(ctx, &pb_admin.UpdateUserRequest{User: user1})
	require.NoError(t, err)
	require.Equal(t, []string{"admin"}, updateResp.User.Roles)

	// Delete user
	_, err = adminClient.DeleteUser(ctx, &pb_admin.DeleteUserRequest{UserId: proto.String("user-1")})
	require.NoError(t, err)

	// Verify deletion
	_, err = adminClient.GetUser(ctx, &pb_admin.GetUserRequest{UserId: proto.String("user-1")})
	require.Error(t, err)

	// Test OIDC Configuration (External Authenticator Hook)
	t.Log("Testing OIDC Login Endpoint...")

	// Verify Auth Login endpoint is active (should redirect to our mock)
	// We use the same client that trusts the mock (oidcClient) but we need to disable redirect following manually if needed.
	// Actually oidcClient follows redirects by default.
	// We want to check the redirect location.
	// So we need a client that does NOT follow redirects but TRUSTS the certs.
	// mockOIDC.Client() returns a client with Transport configured.

	// Clone client config
	checkRedirectClient := *oidcClient
	checkRedirectClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	loginUrl := fmt.Sprintf("http://127.0.0.1:%d/auth/login", jsonrpcPort)
	loginResp, err := checkRedirectClient.Get(loginUrl)
	require.NoError(t, err)
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(loginResp.Body)
		t.Logf("Expected 302, got %d. Body: %s", loginResp.StatusCode, string(body))
	}
	require.Equal(t, http.StatusFound, loginResp.StatusCode)
	loc, err := loginResp.Location()
	require.NoError(t, err)
	// Check that we redirect to the mock OIDC authorization endpoint
	require.Contains(t, loc.String(), mockOIDC.URL+"/authorize")
}
