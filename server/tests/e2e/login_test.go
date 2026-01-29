// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	pb_admin "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestLoginFlow(t *testing.T) {
	// Enable file config for this test
	os.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")
	defer os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")

	// Setup temporary DB path
	dbPath := t.TempDir() + "/mcpany_login_test.db"

	configContent := fmt.Sprintf(`
global_settings:
    db_driver: "sqlite"
    db_path: "%s"
`, dbPath)

	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	var ctx context.Context
	var cancel context.CancelFunc

	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	appRunner := app.NewApplication()

	done := make(chan struct{})
	go func() {
		defer close(done)
		fs := afero.NewOsFs()
		opts := app.RunOptions{
			Ctx:             ctx,
			Fs:              fs,
			Stdio:           false,
			JSONRPCPort:     "127.0.0.1:0",
			GRPCPort:        "127.0.0.1:0",
			ConfigPaths:     []string{tmpFile.Name()},
			ShutdownTimeout: 5 * time.Second,
		}
		if err := appRunner.Run(opts); err != nil && err != context.Canceled {
			t.Logf("Application run error: %v", err)
		}
	}()
	defer func() {
		cancel()
		<-done
	}()

	err = appRunner.WaitForStartup(ctx)
	require.NoError(t, err, "Failed to wait for startup")

	jsonrpcPort := int(appRunner.BoundHTTPPort.Load())
	grpcRegPort := int(appRunner.BoundGRPCPort.Load())

	// Wait for health check
	httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
	integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

	// Connect to gRPC Admin Service to create user
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcRegPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	adminClient := pb_admin.NewAdminServiceClient(conn)

	// 1. Create User with Basic Auth (Get-or-Create to avoid flakes)
	username := "e2e-login-user"
	password := "password123"
	user := &configv1.User{
		Id: proto.String(username),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username:     proto.String(username),
					PasswordHash: proto.String(password), // Will be hashed by server
				},
			},
		},
		Roles: []string{"admin"},
	}

	// Try to get first
	getResp, err := adminClient.GetUser(ctx, &pb_admin.GetUserRequest{UserId: proto.String(username)})
	if err == nil && getResp.User.GetId() == username {
		// User exists, check roles
		require.Contains(t, getResp.User.GetRoles(), "admin", "User exists but missing admin role")
		updateResp, err := adminClient.UpdateUser(ctx, &pb_admin.UpdateUserRequest{User: user})
		require.NoError(t, err, "Failed to update existing user")
		require.Equal(t, username, updateResp.User.GetId())
	} else {
		// Create
		createResp, err := adminClient.CreateUser(ctx, &pb_admin.CreateUserRequest{User: user})
		require.NoError(t, err, "Failed to create user")
		require.Equal(t, username, createResp.User.GetId())
		require.Contains(t, createResp.User.GetRoles(), "admin")
		// Ensure password was hashed (not equal to plaintext)
		assert.NotEqual(t, password, createResp.User.GetAuthentication().GetBasicAuth().GetPasswordHash())
	}

	// Verify User in DB via gRPC to be sure it persisted
	getResp, err = adminClient.GetUser(ctx, &pb_admin.GetUserRequest{UserId: proto.String(username)})
	require.NoError(t, err)
	require.Contains(t, getResp.User.GetRoles(), "admin", "Persisted user missing admin role")

	// 2. Attempt Login via REST API
	loginReq := map[string]string{
		"username": username,
		"password": password,
	}
	body, _ := json.Marshal(loginReq)
	loginURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/auth/login", jsonrpcPort)

	resp, err := http.Post(loginURL, "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(resp.Body)
		t.Fatalf("Login failed: status=%d body=%s", resp.StatusCode, buf.String())
	}
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResp struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(resp.Body).Decode(&loginResp)
	require.NoError(t, err)
	require.NotEmpty(t, loginResp.Token)

	// Verify token is base64 of user:password
	decoded, err := base64.StdEncoding.DecodeString(loginResp.Token)
	require.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%s:%s", username, password), string(decoded))

	// 3. Use Token to Access Protected Endpoint (e.g. List Users)
	// We need a trusted client, but wait, AuthManager handles auth.
	// If we use the token, it should work.

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("http://127.0.0.1:%d/api/v1/users", jsonrpcPort), nil)
	require.NoError(t, err)

	// Add Basic Auth header manually using the token
	req.Header.Set("Authorization", "Basic "+loginResp.Token)

	apiResp, err := client.Do(req)
	require.NoError(t, err)
	defer apiResp.Body.Close()

	if apiResp.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(apiResp.Body)
		t.Logf("API access failed: status=%d body=%s", apiResp.StatusCode, buf.String())
	}
	assert.Equal(t, http.StatusOK, apiResp.StatusCode)
}
