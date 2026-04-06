package tool

import (
	"testing"

	"github.com/mcpany/core/server/pkg/serviceregistry"
)

func TestManager_GetToolCountForService(t *testing.T) {
	tm := NewManager()

	serviceID1 := "service-1"
	serviceID2 := "service-2"
	serviceIDUnhealthy := "service-unhealthy"

	// Setup service info in the sync.Map
	tm.serviceInfo.Store(serviceID1, serviceregistry.ServiceInfo{
		ID:           serviceID1,
		HealthStatus: serviceregistry.HealthStatusHealthy,
	})
	tm.serviceInfo.Store(serviceID2, serviceregistry.ServiceInfo{
		ID:           serviceID2,
		HealthStatus: serviceregistry.HealthStatusHealthy,
	})
	tm.serviceInfo.Store(serviceIDUnhealthy, serviceregistry.ServiceInfo{
		ID:           serviceIDUnhealthy,
		HealthStatus: serviceregistry.HealthStatusUnhealthy,
	})

	// Setup tools mapping directly since we're testing the getter
	tm.mu.Lock()
	tm.serviceToolIDs = map[string]map[string]bool{
		serviceID1: {
			"tool-a": true,
			"tool-b": true,
			"tool-c": true,
		},
		serviceID2: {},
		serviceIDUnhealthy: {
			"tool-d": true,
		},
	}
	tm.mu.Unlock()

	tests := []struct {
		name      string
		serviceID string
		expected  int
	}{
		{
			name:      "Healthy service with multiple tools",
			serviceID: serviceID1,
			expected:  3,
		},
		{
			name:      "Healthy service with no tools",
			serviceID: serviceID2,
			expected:  0,
		},
		{
			name:      "Unhealthy service should return 0 regardless of tools",
			serviceID: serviceIDUnhealthy,
			expected:  0,
		},
		{
			name:      "Non-existent service should return 0",
			serviceID: "non-existent-service",
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tm.GetToolCountForService(tt.serviceID)
			if count != tt.expected {
				t.Errorf("GetToolCountForService(%q) = %d; want %d", tt.serviceID, count, tt.expected)
			}
		})
	}
}

func TestManager_GetAllowedServiceIDs(t *testing.T) {
	tm := NewManager()

	profile1 := "profile-alpha"
	profile2 := "profile-beta"

	// Set up the cache directly to test the getter method
	tm.mu.Lock()
	tm.allowedServicesCache = map[string]map[string]bool{
		profile1: {
			"service-1": true,
			"service-2": true,
		},
		profile2: {},
	}
	tm.mu.Unlock()

	tests := []struct {
		name      string
		profileID string
		wantOk    bool
		wantCount int
	}{
		{
			name:      "Existing profile with allowed services",
			profileID: profile1,
			wantOk:    true,
			wantCount: 2,
		},
		{
			name:      "Existing profile with empty allowed services",
			profileID: profile2,
			wantOk:    true,
			wantCount: 0,
		},
		{
			name:      "Non-existent profile",
			profileID: "non-existent-profile",
			wantOk:    false,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, ok := tm.GetAllowedServiceIDs(tt.profileID)

			if ok != tt.wantOk {
				t.Errorf("GetAllowedServiceIDs(%q) ok = %v; want %v", tt.profileID, ok, tt.wantOk)
			}

			if tt.wantOk {
				if len(allowed) != tt.wantCount {
					t.Errorf("GetAllowedServiceIDs(%q) count = %d; want %d", tt.profileID, len(allowed), tt.wantCount)
				}

				// Verify mapping matches for profile1
				if tt.profileID == profile1 {
					if !allowed["service-1"] || !allowed["service-2"] {
						t.Errorf("GetAllowedServiceIDs(%q) missing expected keys", tt.profileID)
					}
				}
			}
		})
	}
}
