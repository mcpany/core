// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package sql implements SQL storage for the application.
package sql

import (
	"time"
)

// UpstreamService represents the database model for an upstream service configuration.
type UpstreamService struct {
	ID        string `gorm:"primaryKey"`
	Name      string `gorm:"uniqueIndex;not null"`
	Config    string `gorm:"not null"` // JSON content
	CreatedAt time.Time
	UpdatedAt time.Time
}
