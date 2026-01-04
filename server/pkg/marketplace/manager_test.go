// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_AddSubscription(t *testing.T) {
	m := NewManager()
	sub := Subscription{
		Name:        "Test Sub",
		Description: "A test subscription",
		SourceURL:   "http://example.com/feed",
		IsActive:    true,
	}

	created, err := m.AddSubscription(sub)
	assert.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, "Test Sub", created.Name)
	assert.WithinDuration(t, time.Now(), created.CreatedAt, time.Second)

	// Verify it exists in list
	all := m.ListSubscriptions()
	// Should accept at least 2 (default + new)
	assert.GreaterOrEqual(t, len(all), 2)
}

func TestManager_UpdateSubscription(t *testing.T) {
	m := NewManager()
	sub := Subscription{Name: "To Update"}
	created, _ := m.AddSubscription(sub)

	created.Name = "Updated Name"
	updated, err := m.UpdateSubscription(created.ID, created)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)

	// Verify get
	got, ok := m.GetSubscription(created.ID)
	assert.True(t, ok)
	assert.Equal(t, "Updated Name", got.Name)
}

func TestManager_DeleteSubscription(t *testing.T) {
	m := NewManager()
	sub := Subscription{Name: "To Delete"}
	created, _ := m.AddSubscription(sub)

	err := m.DeleteSubscription(created.ID)
	assert.NoError(t, err)

	_, ok := m.GetSubscription(created.ID)
	assert.False(t, ok)
}

func TestManager_SyncSubscription(t *testing.T) {
	m := NewManager()
	// Check default popular sync
	subs := m.ListSubscriptions()
	found := false
	for _, s := range subs {
		if s.ID == "popular" {
			found = true
			break
		}
	}
	assert.True(t, found, "Default popular subscription should exist")

	// Sync it
	err := m.SyncSubscription(context.TODO(), "popular")
	assert.NoError(t, err)

	// Verify check
	synced, ok := m.GetSubscription("popular")
	assert.True(t, ok)
	assert.NotEmpty(t, synced.Services)
	assert.Equal(t, "Filesystem", synced.Services[0].Name)
}
