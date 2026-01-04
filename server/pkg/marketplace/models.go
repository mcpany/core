// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package marketplace

import (
	"time"
)

// Subscription represents a subscribed collection of services.
type Subscription struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	SourceURL   string         `json:"source_url"` // URL to fetch the collection from
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	LastSynced  time.Time      `json:"last_synced"`
	Services    []ServiceEntry `json:"services"`
}

// ServiceEntry represents a single service within a collection.
type ServiceEntry struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Type        string            `json:"type"` // e.g. "http", "stdio"
	Config      map[string]string `json:"config"`
}

// Collection represents a list of services exported by a marketplace/registry.
type Collection struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Services    []ServiceEntry `json:"services"`
}
