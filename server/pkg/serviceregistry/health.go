// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package serviceregistry

// HealthPoint represents a single data point in the health history of a service.
type HealthPoint struct {
	Timestamp int64  `json:"timestamp"` // Unix timestamp in milliseconds
	Status    string `json:"status"`    // "healthy", "unhealthy", "degraded", "inactive"
	Latency   int64  `json:"latency"`   // Latency in milliseconds
	Message   string `json:"message,omitempty"`
}
