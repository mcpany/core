// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package healthstatus

// HealthStatus represents the health of a service.
type HealthStatus int

const (
	// UNKNOWN indicates that the health of the service has not been determined yet.
	UNKNOWN HealthStatus = iota
	// HEALTHY indicates that the service is operating correctly.
	HEALTHY
	// UNHEALTHY indicates that the service is not operating correctly.
	UNHEALTHY
)

func (s HealthStatus) String() string {
	switch s {
	case HEALTHY:
		return "HEALTHY"
	case UNHEALTHY:
		return "UNHEALTHY"
	default:
		return "UNKNOWN"
	}
}
