// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
// NewTemplateManager creates a new instance of TemplateManager.
//
// Summary: Initializes a new TemplateManager.
//
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
// Parameters:
//   - dataDir: string. The directory where template data is persisted.
//
// Returns:
//   - *TemplateManager: The initialized manager.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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

// ListTemplates returns a list of all stored templates.
//
// Summary: Retrieves all managed templates.
//
// Returns:
//   - []*configv1.UpstreamServiceConfig: A list of templates.
// Parameters:
//   - standard arguments based on function signature.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
func (tm *TemplateManager) ListTemplates() []*configv1.UpstreamServiceConfig {
// SaveTemplate saves or updates a template.
//
// Summary: Persists a template.
//
// Parameters:
//   - template: *configv1.UpstreamServiceConfig. The template to save.
//
// Returns:
//   - error: An error if persistence fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
// DeleteTemplate deletes a template by its ID or Name.
//
// Summary: Removes a template.
//
// Parameters:
//   - idOrName: string. The ID or Name of the template to delete.
//
// Returns:
//   - error: An error if deletion or persistence fails.
// Errors:
//   - triggers relevant error states on failure.
// Side Effects:
//   - updates relevant subsystem state or network conditions.
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
