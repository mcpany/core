// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"context"
	"testing"

	pb "github.com/mcpany/core/proto/admin/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/storage/memory"
	"github.com/mcpany/core/server/pkg/tool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServer_BackupRestore(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	store := memory.NewStore()
	tm := tool.NewMockManagerInterface(ctrl)
	sr := &MockServiceRegistry{} // Defined in server_test.go
	s := NewServer(nil, tm, sr, store, nil, nil)
	ctx := context.Background()

	// 1. Populate Store
	user := &configv1.User{}
	user.SetId("user1")
	user.SetRoles([]string{"admin"})
	require.NoError(t, store.CreateUser(ctx, user))

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetId("svc1")
	svc.SetName("svc1")
	cmdSvc := &configv1.CommandLineUpstreamService{}
	cmdSvc.SetCommand("echo")
	svc.SetCommandLineService(cmdSvc)
	require.NoError(t, store.SaveService(ctx, svc))

	secret := &configv1.Secret{}
	secret.SetId("sec1")
	secret.SetName("api-key")
	secret.SetValue("123")
	require.NoError(t, store.SaveSecret(ctx, secret))

	// 2. Create Backup
	backupResp, err := s.CreateBackup(ctx, &pb.CreateBackupRequest{})
	require.NoError(t, err)
	assert.NotNil(t, backupResp.GetConfig())
	assert.Len(t, backupResp.GetConfig().GetUsers(), 1)
	assert.Len(t, backupResp.GetConfig().GetUpstreamServices(), 1)
	assert.Len(t, backupResp.GetConfig().GetSecrets(), 1)
	assert.Equal(t, "user1", backupResp.GetConfig().GetUsers()[0].GetId())

	// 3. Clear Store (Simulate fresh start or restore to new instance)
	store2 := memory.NewStore()
	s2 := NewServer(nil, tm, sr, store2, nil, nil)

	// 4. Restore Backup
	restoreReq := &pb.RestoreBackupRequest{}
	restoreReq.SetConfig(backupResp.GetConfig())
	restoreResp, err := s2.RestoreBackup(ctx, restoreReq)
	require.NoError(t, err)
	assert.Equal(t, int32(1), restoreResp.GetUsersRestored())
	assert.Equal(t, int32(1), restoreResp.GetServicesRestored())
	assert.Equal(t, int32(1), restoreResp.GetSecretsRestored())

	// 5. Verify Store2 State
	u, err := store2.GetUser(ctx, "user1")
	require.NoError(t, err)
	assert.Equal(t, "user1", u.GetId())

	srv, err := store2.GetService(ctx, "svc1")
	require.NoError(t, err)
	assert.Equal(t, "svc1", srv.GetName())

	sec, err := store2.GetSecret(ctx, "sec1")
	require.NoError(t, err)
	assert.Equal(t, "api-key", sec.GetName())
}
