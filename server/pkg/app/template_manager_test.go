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

func TestNewTemplateManager_Seeded(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	assert.NotNil(t, tm)
	assert.NotEmpty(t, tm.ListTemplates())
	assert.Len(t, tm.ListTemplates(), len(BuiltinTemplates))
}

func TestSaveTemplate_New(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	// Clear seeds for easier testing
	tm.templates = []*configv1.UpstreamServiceConfig{}

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("Test Service")
	tmpl.SetId("test-id")
	tmpl.SetVersion("1.0.0")

	err := tm.SaveTemplate(tmpl)
	require.NoError(t, err)

	list := tm.ListTemplates()
	assert.Len(t, list, 1)
	assert.Equal(t, "Test Service", list[0].GetName())
	assert.Equal(t, "test-id", list[0].GetId())

	// Verify file existence
	content, err := os.ReadFile(filepath.Join(tempDir, "templates.json"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Test Service")
}

func TestSaveTemplate_Update(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	tm.templates = []*configv1.UpstreamServiceConfig{}

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
	assert.Len(t, list, 1)
	assert.Equal(t, "1.0.1", list[0].GetVersion())
}

func TestSaveTemplate_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	tm1 := NewTemplateManager(tempDir)
	tm1.templates = []*configv1.UpstreamServiceConfig{}

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("Persistent Service")
	tmpl.SetId("persist-id")

	err := tm1.SaveTemplate(tmpl)
	require.NoError(t, err)

	// Create new manager pointing to same dir
	tm2 := NewTemplateManager(tempDir)
	// It will load the one we saved.
	// Note: It will NOT seed because the file exists and is not empty (it has 1 item).
	list := tm2.ListTemplates()
	assert.Len(t, list, 1)
	assert.Equal(t, "Persistent Service", list[0].GetName())
}

func TestDeleteTemplate(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	tm.templates = []*configv1.UpstreamServiceConfig{}

	tmpl1 := &configv1.UpstreamServiceConfig{}
	tmpl1.SetName("S1")
	tmpl1.SetId("id1")

	tmpl2 := &configv1.UpstreamServiceConfig{}
	tmpl2.SetName("S2")
	tmpl2.SetId("id2")

	require.NoError(t, tm.SaveTemplate(tmpl1))
	require.NoError(t, tm.SaveTemplate(tmpl2))

	assert.Len(t, tm.ListTemplates(), 2)

	err := tm.DeleteTemplate("id1")
	require.NoError(t, err)

	list := tm.ListTemplates()
	assert.Len(t, list, 1)
	assert.Equal(t, "S2", list[0].GetName())

	// Verify persistence
	tm2 := NewTemplateManager(tempDir)
	// Will NOT seed because file exists and has 1 item
	assert.Len(t, tm2.ListTemplates(), 1)
}

func TestDeleteTemplate_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	tm := NewTemplateManager(tempDir)
	tm.templates = []*configv1.UpstreamServiceConfig{}

	tmpl := &configv1.UpstreamServiceConfig{}
	tmpl.SetName("S1")
	tmpl.SetId("id1")

	require.NoError(t, tm.SaveTemplate(tmpl))

	err := tm.DeleteTemplate("non-existent")
	require.NoError(t, err)
	assert.Len(t, tm.ListTemplates(), 1)
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
