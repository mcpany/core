// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app_test

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateManager(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("Initialize Empty", func(t *testing.T) {
		tm := app.NewTemplateManager(tmpDir)
		assert.NotNil(t, tm)
		assert.Empty(t, tm.ListTemplates())
	})

	t.Run("Initialize With Missing Dir", func(t *testing.T) {
		newDir := filepath.Join(tmpDir, "nested/dir")
		tm := app.NewTemplateManager(newDir)
		assert.NotNil(t, tm)
		assert.Empty(t, tm.ListTemplates())
	})
}

func TestTemplateManager_CRUD(t *testing.T) {
	tmpDir := t.TempDir()
	tm := app.NewTemplateManager(tmpDir)

	svc1 := &configv1.UpstreamServiceConfig{}
	svc1.SetId("svc-1")
	svc1.SetName("Service 1")

	cmdSvc1 := &configv1.CommandLineUpstreamService{}
	cmdSvc1.SetCommand("echo hello")
	svc1.SetCommandLineService(cmdSvc1)

	// Create
	err := tm.SaveTemplate(svc1)
	require.NoError(t, err)

	list := tm.ListTemplates()
	require.Len(t, list, 1)
	assert.Equal(t, "svc-1", list[0].GetId())
	assert.Equal(t, "Service 1", list[0].GetName())

	// Update (by ID)
	svc1Update := &configv1.UpstreamServiceConfig{}
	svc1Update.SetId("svc-1")
	svc1Update.SetName("Service 1 Updated")

	cmdSvcUpdate := &configv1.CommandLineUpstreamService{}
	cmdSvcUpdate.SetCommand("echo world")
	svc1Update.SetCommandLineService(cmdSvcUpdate)

	err = tm.SaveTemplate(svc1Update)
	require.NoError(t, err)

	list = tm.ListTemplates()
	require.Len(t, list, 1)
	assert.Equal(t, "Service 1 Updated", list[0].GetName())

	// Update (by Name, implicit ID check fallback logic in SaveTemplate)

	svc2 := &configv1.UpstreamServiceConfig{}
	svc2.SetName("Service 2")
	cmdSvc2 := &configv1.CommandLineUpstreamService{}
	cmdSvc2.SetCommand("ls")
	svc2.SetCommandLineService(cmdSvc2)

	err = tm.SaveTemplate(svc2)
	require.NoError(t, err)
	require.Len(t, tm.ListTemplates(), 2)

	svc2Update := &configv1.UpstreamServiceConfig{}
	svc2Update.SetName("Service 2") // Same name, no ID
	cmdSvc2Update := &configv1.CommandLineUpstreamService{}
	cmdSvc2Update.SetCommand("ls -la")
	svc2Update.SetCommandLineService(cmdSvc2Update)

	err = tm.SaveTemplate(svc2Update)
	require.NoError(t, err)

	list = tm.ListTemplates()
	require.Len(t, list, 2)
	// Find Service 2
	var found *configv1.UpstreamServiceConfig
	for _, s := range list {
		if s.GetName() == "Service 2" {
			found = s
			break
		}
	}
	require.NotNil(t, found)
	require.NotNil(t, found.GetCommandLineService())
	assert.Equal(t, "ls -la", found.GetCommandLineService().GetCommand())

	// Delete by ID
	err = tm.DeleteTemplate("svc-1")
	require.NoError(t, err)
	assert.Len(t, tm.ListTemplates(), 1)

	// Delete by Name
	err = tm.DeleteTemplate("Service 2")
	require.NoError(t, err)
	assert.Empty(t, tm.ListTemplates())
}

func TestTemplateManager_Persistence(t *testing.T) {
	tmpDir := t.TempDir()
	tm1 := app.NewTemplateManager(tmpDir)

	svc := &configv1.UpstreamServiceConfig{}
	svc.SetId("persist-1")
	svc.SetName("Persistent Service")

	require.NoError(t, tm1.SaveTemplate(svc))

	// Re-initialize manager on same directory
	tm2 := app.NewTemplateManager(tmpDir)
	list := tm2.ListTemplates()
	require.Len(t, list, 1)
	assert.Equal(t, "Persistent Service", list[0].GetName())
}

func TestTemplateManager_Concurrency(t *testing.T) {
	tmpDir := t.TempDir()
	tm := app.NewTemplateManager(tmpDir)

	var wg sync.WaitGroup
	count := 100
	wg.Add(count)

	for i := 0; i < count; i++ {
		go func(idx int) {
			defer wg.Done()
			svc := &configv1.UpstreamServiceConfig{}
			svc.SetId("concurrent-svc") // Contention on same ID
			svc.SetName("Concurrent Service")

			// Mix reads and writes
			if idx%2 == 0 {
				_ = tm.SaveTemplate(svc)
			} else {
				_ = tm.ListTemplates()
			}
		}(i)
	}

	wg.Wait()

	// Final state check - should be valid and contain the service
	list := tm.ListTemplates()
	require.Len(t, list, 1)
	assert.Equal(t, "Concurrent Service", list[0].GetName())
}

func TestTemplateManager_CorruptFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "templates.json")

	// Write garbage
	err := os.WriteFile(filePath, []byte("{ not valid json "), 0600)
	require.NoError(t, err)

	// Should not panic, but log error and start empty
	tm := app.NewTemplateManager(tmpDir)
	assert.Empty(t, tm.ListTemplates())
}
