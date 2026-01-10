// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

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

func TestUserManagement_LongPassword(t *testing.T) {
	// Setup temporary DB path
	dbPath := t.TempDir() + "/mcpany_test_longpass.db"

	// Minimal config
	configContent := fmt.Sprintf(`
global_settings:
    db_driver: "sqlite"
    db_path: "%s"
    api_key: "test-api-key"
`, dbPath)

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

	// Retry loop for ports
	for attempt := 0; attempt < 3; attempt++ {
		jsonrpcPort = integration.FindFreePort(t)
		grpcRegPort = integration.FindFreePort(t)
		for grpcRegPort == jsonrpcPort {
			grpcRegPort = integration.FindFreePort(t)
		}

		ctx, cancel = context.WithCancel(context.Background())
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

	// Create a long password (> 72 bytes)
	longPassword := strings.Repeat("a", 75)

	// Create user with long password using Basic Auth (Username/Password)
	// We use the protobuf types correctly now.
	user1 := &configv1.User{
		Id: proto.String("user-longpass"),
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username: proto.String("longuser"),
					Password: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{
							PlainText: longPassword,
						},
					},
				},
			},
		},
		ProfileIds: []string{"profile-1"},
		Roles:      []string{"viewer"},
	}

	// This calls `Password(string)` internally when persisting to DB.
	createResp, err := adminClient.CreateUser(ctx, &pb_admin.CreateUserRequest{User: user1})
	require.NoError(t, err)
	require.Equal(t, "user-longpass", createResp.User.GetId())

	// Verify we can get the user
	getResp, err := adminClient.GetUser(ctx, &pb_admin.GetUserRequest{UserId: proto.String("user-longpass")})
	require.NoError(t, err)
	require.Equal(t, "user-longpass", getResp.User.GetId())

	// We confirmed that CreateUser succeeds with a long password.
	// This exercises the `Password()` function which previously panicked/errored for passwords > 72 bytes.
}
