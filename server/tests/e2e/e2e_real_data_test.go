// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

//go:build e2e

package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/app"
	"github.com/mcpany/core/server/pkg/storage/sqlite"
	"github.com/mcpany/core/server/pkg/testutil"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func TestE2E_RealData_Seeded(t *testing.T) {
	// 1. Setup DB
	dbPath := t.TempDir() + "/seeded.db"
	db, err := sqlite.NewDB(dbPath)
	require.NoError(t, err)
	defer db.Close()

	// 2. Seed Data
	ctx := context.Background()
	err = testutil.Seed(ctx, db.DB)
	require.NoError(t, err)

	// 3. Start Application
	port, err := getFreePort()
	require.NoError(t, err)
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	a := app.NewApplication()

	runErr := make(chan error, 1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		// Set env vars to allow loopback since we seed "http://localhost:8080"
		t.Setenv("MCPANY_ALLOW_LOOPBACK_RESOURCES", "true")
		t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

		err := a.Run(app.RunOptions{
			Ctx:             ctx,
			Fs:              afero.NewMemMapFs(),
			Stdio:           false,
			JSONRPCPort:     addr,
			GRPCPort:        "", // Disable gRPC for simplicity
			ConfigPaths:     []string{},
			APIKey:          "test-api-key", // Matches seeded global settings
			DBPath:          dbPath,
			ShutdownTimeout: 1 * time.Second,
		})
		if err != nil {
			runErr <- err
		}
	}()

	// Wait for startup
	require.NoError(t, a.WaitForStartup(ctx))

	// 4. Verify API
	baseURL := "http://" + addr
	client := &http.Client{Timeout: 5 * time.Second}

	// Retry loop for readiness
	require.Eventually(t, func() bool {
		req, _ := http.NewRequest("GET", baseURL+"/health", nil)
		resp, err := client.Do(req)
		if err != nil {
			return false
		}
		defer resp.Body.Close()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 100*time.Millisecond)

	// 5. Query Services
	req, err := http.NewRequest("GET", baseURL+"/api/v1/services", nil)
	require.NoError(t, err)
	req.Header.Set("X-API-Key", "test-api-key")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var services []interface{}
	err = json.NewDecoder(resp.Body).Decode(&services)
	require.NoError(t, err)

	found := false
	for _, svc := range services {
		s, ok := svc.(map[string]interface{})
		if !ok {
			continue
		}
		// Check both casing for robustness
		if s["name"] == "test-service" || s["Name"] == "test-service" {
			found = true
			break
		}
	}
	require.True(t, found, "Seeded service 'test-service' not found in API response")
}
