// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package e2e

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
	"database/sql"
	"encoding/json"

	pb_admin "github.com/mcpany/core/proto/admin/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "modernc.org/sqlite"
)

func TestAuditLogE2E(t *testing.T) {
	os.Setenv("MCPANY_ENABLE_FILE_CONFIG", "true")
	defer os.Unsetenv("MCPANY_ENABLE_FILE_CONFIG")

	dbPath := t.TempDir() + "/mcpany_audit_e2e_test.db"
	configContent := fmt.Sprintf(`
global_settings:
    db_driver: "sqlite"
    db_path: "%s"
    audit:
        enabled: true
`, dbPath)

	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Seed the database
	db, err := sql.Open("sqlite", dbPath)
	require.NoError(t, err)

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME NOT NULL,
			tool_name TEXT NOT NULL,
			user_id TEXT,
			profile_id TEXT,
			arguments TEXT,
			result TEXT,
			error TEXT,
			duration TEXT,
			duration_ms INTEGER,
			trace_id TEXT,
			span_id TEXT,
			parent_id TEXT
		);
	`)
	require.NoError(t, err)

	_, err = db.Exec(`
		INSERT INTO audit_logs (timestamp, tool_name, user_id, profile_id, arguments, result, error, duration, duration_ms, trace_id, span_id, parent_id)
		VALUES
		('2023-10-27T10:00:00Z', 'calculate_tax', 'user123', 'default', '{"amount": 1000, "rate": 0.05}', '{"tax": 50, "total": 1050}', '', '120ms', 120, 'trace-12345', 'span-1', ''),
		('2023-10-27T10:05:00Z', 'create_user', 'admin', 'default', '{"username": "newuser"}', '', 'User already exists', '45ms', 45, 'trace-67890', 'span-2', '')
	`)
	require.NoError(t, err)
	db.Close()

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
			t.Logf("Application run error: %%v", err)
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

	httpUrl := fmt.Sprintf("http://127.0.0.1:%d/healthz", jsonrpcPort)
	integration.WaitForHTTPHealth(t, httpUrl, 10*time.Second)

	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", grpcRegPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(t, err)
	defer conn.Close()

	adminClient := pb_admin.NewAdminServiceClient(conn)

	// Since we seeded the DB, it should not be empty
	resp, err := adminClient.ListAuditLogs(ctx, &pb_admin.ListAuditLogsRequest{})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Len(t, resp.Entries, 2)
}
