// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"testing"
)

func TestManager_Rules(t *testing.T) {
	m := NewManager()

	// 1. Create Rule
	rule := &AlertRule{
		Name:      "Test Rule",
		Metric:    "cpu",
		Operator:  ">",
		Threshold: 80,
		Enabled:   true,
		Severity:  SeverityWarning,
	}
	created := m.CreateRule(rule)
	if created.ID == "" {
		t.Error("Expected rule ID to be generated")
	}
	if created.Name != "Test Rule" {
		t.Errorf("Expected name 'Test Rule', got %s", created.Name)
	}

	// 2. Get Rule
	fetched := m.GetRule(created.ID)
	if fetched == nil {
		t.Error("Failed to fetch created rule")
	} else if fetched.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, fetched.ID)
	}

	// 3. List Rules
	// NewManager no longer seeds rules. So total should be 1.
	list := m.ListRules()
	if len(list) != 1 {
		t.Errorf("Expected 1 rule, got %d", len(list))
	}

	// 4. Update Rule
	update := &AlertRule{
		Name:      "Updated Rule",
		Metric:    "memory",
		Operator:  "<",
		Threshold: 20,
		Enabled:   false,
		Severity:  SeverityCritical,
	}
	updated := m.UpdateRule(created.ID, update)
	if updated == nil {
		t.Error("Failed to update rule")
	}
	if updated.Name != "Updated Rule" {
		t.Errorf("Expected name 'Updated Rule', got %s", updated.Name)
	}
	if updated.Metric != "memory" {
		t.Errorf("Expected metric 'memory', got %s", updated.Metric)
	}

	// Verify persistence of update
	fetchedAfterUpdate := m.GetRule(created.ID)
	if fetchedAfterUpdate.Name != "Updated Rule" {
		t.Error("Update was not persisted")
	}

	// 5. Delete Rule
	err := m.DeleteRule(created.ID)
	if err != nil {
		t.Errorf("Failed to delete rule: %v", err)
	}

	deleted := m.GetRule(created.ID)
	if deleted != nil {
		t.Error("Rule should be nil after deletion")
	}

	listAfterDelete := m.ListRules()
	if len(listAfterDelete) != 0 {
		t.Errorf("Expected 0 rules after deletion, got %d", len(listAfterDelete))
	}
}
