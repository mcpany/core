package util

import (
	"testing"
)

// TestReplaceURLPath_Integration simulates a real-world scenario of constructing an external API URL
// where one of the parameters might be a pointer (e.g. from a struct field that is a pointer).
// This serves as an integration test for the utility package's public API.
func TestReplaceURLPath_Integration(t *testing.T) {
	// Scenario: Constructing a URL for a user profile update API call.
	// The user struct has a pointer to an organization ID (optional field).

	type User struct {
		ID     string
		OrgID  *string
		Active *bool
	}

	orgID := "org-999"
	active := true
	user := User{
		ID:     "user-123",
		OrgID:  &orgID,
		Active: &active,
	}

	// We want to construct: /api/v1/orgs/org-999/users/user-123?active=true

	urlTemplate := "/api/v1/orgs/{{orgID}}/users/{{userID}}"
	queryTemplate := "active={{active}}"

	params := map[string]interface{}{
		"orgID":  user.OrgID,
		"userID": user.ID,
		"active": user.Active,
	}

	// 1. Path Replacement
	gotPath := ReplaceURLPath(urlTemplate, params, nil)
	expectedPath := "/api/v1/orgs/org-999/users/user-123"

	if gotPath != expectedPath {
		t.Errorf("Path mismatch.\nGot:  %s\nWant: %s", gotPath, expectedPath)
	}

	// 2. Query Replacement
	gotQuery := ReplaceURLQuery(queryTemplate, params, nil)
	expectedQuery := "active=true"

	if gotQuery != expectedQuery {
		t.Errorf("Query mismatch.\nGot:  %s\nWant: %s", gotQuery, expectedQuery)
	}

	// 3. Verify nil pointer handling (optional field is nil)
	user2 := User{
		ID:    "user-456",
		OrgID: nil, // No org
	}
	params2 := map[string]interface{}{
		"orgID": user2.OrgID,
	}

	// Should produce <nil> or handle it gracefully?
	// Current behavior: ToString(nil pointer) -> ToString(nil) -> "<nil>"
	// URL encoding might encode < and >
	// <nil> -> %3Cnil%3E

	gotPathNil := ReplaceURLPath("/orgs/{{orgID}}", params2, nil)
	expectedPathNil := "/orgs/%3Cnil%3E"

	if gotPathNil != expectedPathNil {
		t.Errorf("Nil pointer mismatch.\nGot:  %s\nWant: %s", gotPathNil, expectedPathNil)
	}
}
