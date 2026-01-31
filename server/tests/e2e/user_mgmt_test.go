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
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
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
        redirect_url: "http://127.0.0.1:8080/callback"
`, dbPath, mockOIDC.URL)

	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	var ctx context.Context
	var cancel context.CancelFunc

	// Prepare Context with Trusted OIDC Client
	baseCtx := context.Background()
	oidcClient := mockOIDC.Client()
	baseCtx = oidc.ClientContext(baseCtx, oidcClient)

	ctx, cancel = context.WithCancel(baseCtx)
	defer cancel()

	appRunner := app.NewApplication()

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Mock filesystem with our config
		fs := afero.NewOsFs()
		// Use 127.0.0.1:0 to let OS choose free ports and avoid dual-stack flakes
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     "127.0.0.1:0",
			GRPCPort:        "127.0.0.1:0",
			ConfigPaths:     []string{tmpFile.Name()},
			APIKey:          "",
			ShutdownTimeout: 5 * time.Second,
		}
		err := appRunner.Run(opts)
		if err != nil && err != context.Canceled {
			t.Logf("Application run error: %v", err)
		}
	}()
	defer func() {
		cancel()
		<-done
	}()

	// Wait for app to start
	err = appRunner.WaitForStartup(ctx)
	require.NoError(t, err, "Failed to wait for startup")

	jsonrpcPort := int(appRunner.BoundHTTPPort.Load())
	grpcRegPort := int(appRunner.BoundGRPCPort.Load())

	require.NotZero(t, jsonrpcPort, "JSON RPC Port should be bound")
	require.NotZero(t, grpcRegPort, "gRPC Port should be bound")

	// Wait for health check
	httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
	integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

	// Connect to gRPC Admin Service
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcRegPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	adminClient := pb_admin.NewAdminServiceClient(conn)

	// Create initial user
	userID := fmt.Sprintf("e2e-test-user-%d", time.Now().UnixNano())
	user1 := configv1.User_builder{
		Id: proto.String(userID),
		Authentication: configv1.Authentication_builder{
			ApiKey: configv1.APIKeyAuth_builder{
				ParamName:         proto.String("X-API-Key"),
				VerificationValue: proto.String("secret-key"),
			}.Build(),

		}.Build(),
		ProfileIds: []string{"profile-1"},
		Roles:      []string{"viewer"},
	}.Build()

	createResp, err := adminClient.CreateUser(ctx, pb_admin.CreateUserRequest_builder{User: user1}.Build())
	require.NoError(t, err)
	require.Equal(t, userID, createResp.GetUser().GetId())
	require.Empty(t, createResp.GetUser().GetAuthentication().GetApiKey().GetVerificationValue())

	// Get user
	getResp, err := adminClient.GetUser(ctx, pb_admin.GetUserRequest_builder{UserId: proto.String(userID)}.Build())
	require.NoError(t, err)
	require.Equal(t, userID, getResp.GetUser().GetId())

	// List users
	listResp, err := adminClient.ListUsers(ctx, pb_admin.ListUsersRequest_builder{}.Build())
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listResp.GetUsers()), 1)
	var foundUser *configv1.User
	for _, u := range listResp.GetUsers() {
		if u.GetId() == userID {
			foundUser = u
			break
		}
	}
	require.NotNil(t, foundUser, fmt.Sprintf("%s should be in the list", userID))
	require.Equal(t, userID, foundUser.GetId())

	// Update user
	user1.SetRoles([]string{"admin"})
	updateResp, err := adminClient.UpdateUser(ctx, pb_admin.UpdateUserRequest_builder{User: user1}.Build())
	require.NoError(t, err)
	require.Equal(t, []string{"admin"}, updateResp.GetUser().GetRoles())

	// Delete user
	_, err = adminClient.DeleteUser(ctx, pb_admin.DeleteUserRequest_builder{UserId: proto.String(userID)}.Build())
	require.NoError(t, err)

	// Verify deletion
	_, err = adminClient.GetUser(ctx, pb_admin.GetUserRequest_builder{UserId: proto.String(userID)}.Build())
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
