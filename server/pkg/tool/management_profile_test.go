// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	v1 "github.com/mcpany/core/proto/mcp_router/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestManager_ProfileMethods_Coverage(t *testing.T) {
	manager := NewManager(nil)

	// Setup profiles
	p1 := &configv1.ProfileDefinition{
		Name: proto.String("p1"),
		Selector: &configv1.ProfileSelector{
			Tags: []string{"tag1"},
		},
		ServiceConfig: map[string]*configv1.ProfileServiceConfig{
			"s1": {Enabled: proto.Bool(true)},
			"s2": {Enabled: proto.Bool(false)},
		},
	}
	manager.SetProfiles([]string{"p1"}, []*configv1.ProfileDefinition{p1})

	// Test IsServiceAllowed
	assert.True(t, manager.IsServiceAllowed("s1", "p1"))
	assert.False(t, manager.IsServiceAllowed("s2", "p1"))
	assert.False(t, manager.IsServiceAllowed("s3", "p1"))
	assert.False(t, manager.IsServiceAllowed("s1", "nonexistent"))

	// Test GetAllowedServiceIDs
	allowed, ok := manager.GetAllowedServiceIDs("p1")
	assert.True(t, ok)
	assert.Contains(t, allowed, "s1")
	assert.NotContains(t, allowed, "s2")
	assert.Equal(t, 1, len(allowed))

	_, ok = manager.GetAllowedServiceIDs("nonexistent")
	assert.False(t, ok)

	// Test ToolMatchesProfile
	mockTool := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("s1"),
				Tags:      []string{"tag1"},
			}
		},
	}
	// Matches by service config
	assert.True(t, manager.ToolMatchesProfile(mockTool, "p1"))

	mockTool3 := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				ServiceId: proto.String("s3"),
				Tags:      []string{"tag1"},
			}
		},
	}
	// Service s3 is not in config, but tag1 matches selector
	assert.True(t, manager.ToolMatchesProfile(mockTool3, "p1"))

	// Test explicit profile assignment on tool
	mockToolExplicit := &MockTool{
		ToolFunc: func() *v1.Tool {
			return &v1.Tool{
				Profiles: []string{"p1"},
			}
		},
	}
	assert.True(t, manager.ToolMatchesProfile(mockToolExplicit, "p1"))
}
