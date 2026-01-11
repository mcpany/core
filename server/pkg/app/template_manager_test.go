// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"os"
	"path/filepath"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestTemplateManager_Persistence(t *testing.T) {
	// Setup temp dir
	tmpDir, err := os.MkdirTemp("", "template_test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Test NewTemplateManager (initially empty)
	tm := NewTemplateManager(tmpDir)
	assert.Empty(t, tm.ListTemplates())

	// Test SaveTemplate
	tpl1 := &configv1.UpstreamServiceConfig{
		Name: proto.String("svc1"),
		Id:   proto.String("id1"),
	}
	err = tm.SaveTemplate(tpl1)
	require.NoError(t, err)

	list := tm.ListTemplates()
	require.Len(t, list, 1)
	assert.Equal(t, "svc1", list[0].GetName())
	assert.Equal(t, "id1", list[0].GetId())

	// Test Persistence (NewManager from same dir)
	tm2 := NewTemplateManager(tmpDir)
	list2 := tm2.ListTemplates()
	require.Len(t, list2, 1)
	assert.Equal(t, "svc1", list2[0].GetName())

	// Test Update (same name, same ID)
	tpl1Update := &configv1.UpstreamServiceConfig{
		Name:          proto.String("svc1"),
		Id:            proto.String("id1"),
		SanitizedName: proto.String("sname1"),
	}
	err = tm.SaveTemplate(tpl1Update)
	require.NoError(t, err)

	list3 := tm.ListTemplates()
	require.Len(t, list3, 1)
	assert.Equal(t, "sname1", list3[0].GetSanitizedName())

	// Test Delete
	err = tm.DeleteTemplate("id1")
	require.NoError(t, err)
	assert.Empty(t, tm.ListTemplates())

	// Verify persistence after delete
	tm3 := NewTemplateManager(tmpDir)
	assert.Empty(t, tm3.ListTemplates())
}

func TestTemplateManager_SaveLogic(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "template_test_logic")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tm := NewTemplateManager(tmpDir)

	// Save tpl1
	tpl1 := &configv1.UpstreamServiceConfig{
		Name: proto.String("svc1"),
		Id:   proto.String("id1"),
	}
	tm.SaveTemplate(tpl1)

	// Save tpl2 (diff name, diff ID) -> Append
	tpl2 := &configv1.UpstreamServiceConfig{
		Name: proto.String("svc2"),
		Id:   proto.String("id2"),
	}
	tm.SaveTemplate(tpl2)
	assert.Len(t, tm.ListTemplates(), 2)

	// Save tpl3 (same name as svc1, diff ID) -> Append?
	// Logic: if t.GetName() == template.GetName() { ... }
	// If Name matches, it checks ID.
	// if template.GetId() != "" && t.GetId() == template.GetId() -> Update
	// But if IDs differ, does it update?
	// The current logic:
	// if t.GetName() == template.GetName() {
	//    if template.GetId() != "" && t.GetId() == template.GetId() { update }
	//    if template.GetId() == "" && t.GetName() == template.GetName() { update }
	// }
	// If Name matches but ID differs (and not empty), it does NOT update inside the loop?
	// It falls through?
	// And `found` remains false?
	// Then it appends?
	// So we can have multiple templates with same Name but different IDs?

	tplSameNameDiffID := &configv1.UpstreamServiceConfig{
		Name: proto.String("svc1"),
		Id:   proto.String("id3"),
	}
	tm.SaveTemplate(tplSameNameDiffID)

	// If it appended, we have 3.
	// If it updated, we have 2.
	// Based on code reading, it should append because `t.GetId() (id1) != template.GetId() (id3)`.
	list := tm.ListTemplates()
	// Actually, wait.
	// The loop continues if not found?
	// Yes.
	// So it appends.
	// Let's verify this behavior.
	assert.Len(t, list, 3)
}

func TestTemplateManager_LoadCorrupt(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "template_test_corrupt")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	path := filepath.Join(tmpDir, "templates.json")
	os.WriteFile(path, []byte("{invalid json"), 0600)

	// Should not panic, should log error and start empty?
	// load() returns error. NewTemplateManager logs it and continues.
	tm := NewTemplateManager(tmpDir)
	assert.Empty(t, tm.ListTemplates())
}
