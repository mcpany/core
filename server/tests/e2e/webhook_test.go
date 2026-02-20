// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	pb_admin "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
)

func TestWebhookFlow(t *testing.T) {
	// Enable file config for this test
	os.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")
	defer os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")

	// Setup temporary DB path
	dbPath := t.TempDir() + "/mcpany_webhook_test.db"

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

	// Start Mock Webhook Receiver
	webhookReceived := make(chan string, 1)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webhookReceived <- r.Header.Get("X-MCP-Event")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Connect to gRPC Admin Service
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcRegPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	adminClient := pb_admin.NewAdminServiceClient(conn)

	// 1. Create Webhook
	wh := configv1.SystemWebhook_builder{
		Url:    ts.URL,
		Events: []string{"all"},
		Active: true,
	}.Build()

	createResp, err := adminClient.CreateSystemWebhook(ctx, pb_admin.CreateSystemWebhookRequest_builder{Webhook: wh}.Build())
	require.NoError(t, err)
	createdWh := createResp.GetWebhook()
	assert.NotEmpty(t, createdWh.GetId())
	assert.Equal(t, ts.URL, createdWh.GetUrl())

	// 2. Verify List
	listResp, err := adminClient.ListSystemWebhooks(ctx, &pb_admin.ListSystemWebhooksRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.GetWebhooks(), 1)
	assert.Equal(t, createdWh.GetId(), listResp.GetWebhooks()[0].GetId())

	// 3. Test Delivery (Simulated)
	testResp, err := adminClient.TestSystemWebhook(ctx, pb_admin.TestSystemWebhookRequest_builder{Id: proto.String(createdWh.GetId())}.Build())
	require.NoError(t, err)
	assert.True(t, testResp.GetSuccess())

	// 4. Delete
	_, err = adminClient.DeleteSystemWebhook(ctx, pb_admin.DeleteSystemWebhookRequest_builder{Id: proto.String(createdWh.GetId())}.Build())
	require.NoError(t, err)

	// Verify Deleted
	listResp, err = adminClient.ListSystemWebhooks(ctx, &pb_admin.ListSystemWebhooksRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.GetWebhooks(), 0)
}
