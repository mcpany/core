// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// TemplateManager manages the persistence and lifecycle of service templates.
//
// Summary: manages the persistence and lifecycle of service templates.
type TemplateManager struct {
	mu        sync.RWMutex
	templates []*configv1.UpstreamServiceConfig
	filePath  string
}

// NewTemplateManager creates a new instance of TemplateManager.
//
// Summary: creates a new instance of TemplateManager.
//
// Parameters:
//   - dataDir: string. The dataDir.
//
// Returns:
//   - *TemplateManager: The *TemplateManager.
func NewTemplateManager(dataDir string) *TemplateManager {
	tm := &TemplateManager{
		filePath: filepath.Join(dataDir, "templates.json"),
	}
	if err := tm.load(); err != nil {
		logging.GetLogger().Info("No existing templates found or failed to load, starting empty", "error", err)
	}
	tm.seedAndSave()
	return tm
}

// seedAndSave helper to avoid lock contention.
func (tm *TemplateManager) seedAndSave() {
	tm.mu.Lock()
	if len(tm.templates) > 0 {
		tm.mu.Unlock()
		return
	}

	logging.GetLogger().Info("Seeding builtin templates", "count", len(BuiltinTemplates))
	for _, t := range BuiltinTemplates {
		tm.templates = append(tm.templates, proto.Clone(t).(*configv1.UpstreamServiceConfig))
	}
	tm.mu.Unlock()

	if err := tm.save(); err != nil {
		logging.GetLogger().Error("failed to save builtin templates", "error", err)
	}
}

func (tm *TemplateManager) load() error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	data, err := os.ReadFile(tm.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var list []json.RawMessage
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}

	tm.templates = make([]*configv1.UpstreamServiceConfig, 0, len(list))
	for _, raw := range list {
		var svc configv1.UpstreamServiceConfig
		if err := protojson.Unmarshal(raw, &svc); err != nil {
			logging.GetLogger().Error("failed to unmarshal template", "error", err)
			continue
		}
		tm.templates = append(tm.templates, &svc)
	}
	return nil
}

func (tm *TemplateManager) save() error {
	// Access should be held by caller or we accept race if called internally?
	// It's safer to not lock here if caller locks, but here we want to ensure atomic write.
	// But `load` locks.
	// Let's rely on internal helpers or just Lock in public methods.
	// Making save private and assuming caller has lock effectively?
	// No, `save` is IO heavy.
	// Implementation:
	// Lock for Read serialized data.
	// Unlock.
	// Write to file.

	tm.mu.RLock()
	opts := protojson.MarshalOptions{UseProtoNames: true}
	list := make([]json.RawMessage, 0, len(tm.templates))
	for _, t := range tm.templates {
		b, err := opts.Marshal(t)
		if err != nil {
			continue
		}
		list = append(list, b)
	}
	tm.mu.RUnlock()

	data, err := json.MarshalIndent(list, "", "  ")
	if err != nil {
		return err
	}

	// Ensure dir exists
	if err := os.MkdirAll(filepath.Dir(tm.filePath), 0750); err != nil {
		return err
	}
	return os.WriteFile(tm.filePath, data, 0600)
}

// ListTemplates returns a list of all stored templates.
//
// Summary: returns a list of all stored templates.
//
// Parameters:
//   None.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: The []*configv1.UpstreamServiceConfig.
func (tm *TemplateManager) ListTemplates() []*configv1.UpstreamServiceConfig {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	// Return a copy?
	// Shallow copy of slice is fine if objects are treated immutable or we clone.
	// For now just return slice.
	res := make([]*configv1.UpstreamServiceConfig, len(tm.templates))
	copy(res, tm.templates)
	return res
}

// SaveTemplate saves or updates a template.
//
// Summary: saves or updates a template.
//
// Parameters:
//   - template: *configv1.UpstreamServiceConfig. The template.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (tm *TemplateManager) SaveTemplate(template *configv1.UpstreamServiceConfig) error {
	tm.mu.Lock()
	found := false
	for i, t := range tm.templates {
		if t.GetName() == template.GetName() { // Identify by Name for now? Or ID? ID is safer.
			// If ID missing, use Name?
			if template.GetId() != "" && t.GetId() == template.GetId() {
				tm.templates[i] = template
				found = true
				break
			}
			if template.GetId() == "" && t.GetName() == template.GetName() {
				tm.templates[i] = template
				found = true
				break
			}
		}
	}
	if !found {
		tm.templates = append(tm.templates, template)
	}
	tm.mu.Unlock()
	return tm.save()
}

// DeleteTemplate deletes a template by its ID or Name.
//
// Summary: deletes a template by its ID or Name.
//
// Parameters:
//   - idOrName: string. The idOrName.
//
// Returns:
//   - error: An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
func (tm *TemplateManager) DeleteTemplate(idOrName string) error {
	tm.mu.Lock()
	newTemplates := make([]*configv1.UpstreamServiceConfig, 0, len(tm.templates))
	for _, t := range tm.templates {
		if t.GetId() == idOrName || t.GetName() == idOrName {
			continue
		}
		newTemplates = append(newTemplates, t)
	}
	tm.templates = newTemplates
	tm.mu.Unlock()
	return tm.save()
}
