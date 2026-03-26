// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package alerts

import (
	"testing"
)

// TestManager_CreateAndGet ...
// Summary: TestManager_CreateAndGet
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m := NewManager()
	alert := &Alert{
		Title:    "Test Alert",
		Severity: SeverityInfo,
		Status:   StatusActive,
	}
	created := m.CreateAlert(alert)
	if created.ID == "" {
		t.Error("expected ID to be generated")
	}

	got := m.GetAlert(created.ID)
	if got == nil {
		t.Error("expected to get alert")
	}
	if got.Title != "Test Alert" {
		t.Errorf("expected title 'Test Alert', got '%s'", got.Title)
	}
}

// TestManager_List ...
// Summary: TestManager_List
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m := NewManager()
	// Should have seeded data (5 items)
	list := m.ListAlerts()
	if len(list) < 5 {
		t.Errorf("expected at least 5 seeded alerts, got %d", len(list))
	}
}

// TestManager_Update ...
// Summary: TestManager_Update
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m := NewManager()
	alert := &Alert{Title: "Test", Status: StatusActive}
	created := m.CreateAlert(alert)

	updated := m.UpdateAlert(created.ID, &Alert{Status: StatusResolved})
	if updated.Status != StatusResolved {
		t.Errorf("expected status Resolved, got %s", updated.Status)
	}

	got := m.GetAlert(created.ID)
	if got.Status != StatusResolved {
		t.Errorf("expected persisted status Resolved, got %s", got.Status)
	}
}

// TestManager_Webhook ...
// Summary: TestManager_Webhook
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m := NewManager()
	url := "http://example.com/webhook"
	m.SetWebhookURL(url)

	if got := m.GetWebhookURL(); got != url {
		t.Errorf("expected webhook URL %s, got %s", url, got)
	}
}

// TestManager_GetAlertStats ...
// Summary: TestManager_GetAlertStats
// Parameters:
//   - None.
// Returns:
//   - None.
// Errors:
//   - None.
// Side Effects:
//   - None.
	m := NewManager()
	stats := m.GetAlertStats()
	if stats == nil {
		t.Error("expected non-nil stats")
	}

	// With the seeded data, we should have 1 active critical, 1 active warning, and at least some total today
	if stats.ActiveCritical != 1 {
		t.Errorf("expected 1 active critical alert, got %d", stats.ActiveCritical)
	}
	if stats.ActiveWarning != 1 {
		t.Errorf("expected 1 active warning alert, got %d", stats.ActiveWarning)
	}
	if stats.TotalToday < 1 {
		t.Errorf("expected >0 total today, got %d", stats.TotalToday)
	}
	if stats.MTTR == "" {
		t.Error("expected non-empty MTTR")
	}
}
