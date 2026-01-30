// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateManager_Empty(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	assert.NotNil(t, tm)
	assert.Len(t, tm.ListTemplates(), len(BuiltinTemplates))
}

func TestSaveTemplate_New(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("Test Service")
	tmpl.SetId("test-id")
	tmpl.SetVersion("1.0.0")

	err := tm.SaveTemplate(tmpl)
	require.NoError(t, err)

	list := tm.ListTemplates()
	assert.Len(t, list, len(BuiltinTemplates)+1)

	// Find the new template
	found := false
	for _, tmpl := range list {
		if tmpl.GetId() == "test-id" {
			found = true
			assert.Equal(t, "Test Service", tmpl.GetName())
			break
		}
	}
	assert.True(t, found, "Newly saved template not found")

	// Verify file existence
	content, err := os.ReadFile(filepath.Join(tempDir, "templates.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Test Service")
}

func TestSaveTemplate_Update(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("Test Service")
	tmpl.SetId("test-id")
	tmpl.SetVersion("1.0.0")

	err := tm.SaveTemplate(tmpl)
	require.NoError(t, err)

	// Update
	tmpl.SetVersion("1.0.1")
	err = tm.SaveTemplate(tmpl)
	require.NoError(t, err)

	list := tm.ListTemplates()
	assert.Len(t, list, len(BuiltinTemplates)+1)

	for _, tmpl := range list {
		if tmpl.GetId() == "test-id" {
			assert.Equal(t, "1.0.1", tmpl.GetVersion())
		}
	}
}

func TestSaveTemplate_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	tm1 := NewTemplateManager(tempDir)

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("Persistent Service")
	tmpl.SetId("persist-id")

	err := tm1.SaveTemplate(tmpl)
	require.NoError(t, err)

	// Create new manager pointing to same dir
	tm2 := NewTemplateManager(tempDir)
	list := tm2.ListTemplates()
	assert.Len(t, list, len(BuiltinTemplates)+1)

	found := false
	for _, tmpl := range list {
		if tmpl.GetId() == "persist-id" {
			found = true
			assert.Equal(t, "Persistent Service", tmpl.GetName())
		}
	}
	assert.True(t, found)
}

func TestDeleteTemplate(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	tmpl1 := &configv1.UpstreamServiceConfig{}
	tmpl1.SetName("S1")
	tmpl1.SetId("id1")

	tmpl2 := &configv1.UpstreamServiceConfig{}
	tmpl2.SetName("S2")
	tmpl2.SetId("id2")

	require.NoError(t, tm.SaveTemplate(tmpl1))
	require.NoError(t, tm.SaveTemplate(tmpl2))

	assert.Len(t, tm.ListTemplates(), len(BuiltinTemplates)+2)

	err := tm.DeleteTemplate("id1")
	require.NoError(t, err)

	list := tm.ListTemplates()
	assert.Len(t, list, len(BuiltinTemplates)+1)

	foundS2 := false
	foundS1 := false
	for _, tmpl := range list {
		if tmpl.GetId() == "id2" {
			foundS2 = true
		}
		if tmpl.GetId() == "id1" {
			foundS1 = true
		}
	}
	assert.True(t, foundS2)
	assert.False(t, foundS1)

	// Verify persistence
	tm2 := NewTemplateManager(tempDir)
	assert.Len(t, tm2.ListTemplates(), len(BuiltinTemplates)+1)
}

func TestDeleteTemplate_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("S1")
	tmpl.SetId("id1")

	require.NoError(t, tm.SaveTemplate(tmpl))

	err := tm.DeleteTemplate("non-existent")
	require.NoError(t, err)
	assert.Len(t, tm.ListTemplates(), len(BuiltinTemplates)+1)
}

func TestConcurrency_Safe(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)

	var wg sync.WaitGroup
	workers := 10
	iterations := 50

	// Writer goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				tmpl := &configv1.UpstreamServiceConfig{}
				tmpl.SetName("Concurrent Service")
				tmpl.SetId("concurrent-id")

				if err := tm.SaveTemplate(tmpl); err != nil {
					t.Errorf("failed to save: %v", err)
				}
			}
		}(i)
	}

	// Deleter goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if err := tm.DeleteTemplate("concurrent-id"); err != nil {
					t.Errorf("failed to delete: %v", err)
				}
			}
		}(i)
	}

	wg.Wait()
}
