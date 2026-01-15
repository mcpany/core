// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package profile

import (
	"strings"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestResolveProfile(t *testing.T) {
	profiles := []*configv1.ProfileDefinition{
		{
			Name: proto.String("root"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-a": {
					Enabled: proto.Bool(true),
				},
			},
			Secrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "root-value"}},
			},
		},
		{
			Name:             proto.String("intermediate"),
			ParentProfileIds: []string{"root"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-b": {
					Enabled: proto.Bool(true),
				},
				// Override service-a partially
				"service-a": {
					Enabled: proto.Bool(true),
				},
			},
			Secrets: map[string]*configv1.SecretValue{
				"INTERMEDIATE_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "intermediate-value"}},
			},
		},
		{
			Name:             proto.String("leaf"),
			ParentProfileIds: []string{"intermediate"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-c": {
					Enabled: proto.Bool(true),
				},
				// Disable service-b
				"service-b": {
					Enabled: proto.Bool(false),
				},
			},
			Secrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "leaf-override-root"}},
			},
		},
	}

	m := NewManager(profiles)

	tests := []struct {
		name            string
		profile         string
		expectedConfigs map[string]*configv1.ProfileServiceConfig
		expectedSecrets map[string]*configv1.SecretValue
		wantErr         bool
	}{
		{
			name:    "resolve root",
			profile: "root",
			expectedConfigs: map[string]*configv1.ProfileServiceConfig{
				"service-a": {
					Enabled: proto.Bool(true),
				},
			},
			expectedSecrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "root-value"}},
			},
		},
		{
			name:    "resolve intermediate",
			profile: "intermediate",
			expectedConfigs: map[string]*configv1.ProfileServiceConfig{
				"service-a": {
					Enabled: proto.Bool(true),
				},
				"service-b": {
					Enabled: proto.Bool(true),
				},
			},
			expectedSecrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET":         {Value: &configv1.SecretValue_PlainText{PlainText: "root-value"}},
				"INTERMEDIATE_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "intermediate-value"}},
			},
		},
		{
			name:    "resolve leaf",
			profile: "leaf",
			expectedConfigs: map[string]*configv1.ProfileServiceConfig{
				"service-a": {
					Enabled: proto.Bool(true),
				},
				"service-b": {
					Enabled: proto.Bool(false), // Overridden to false
				},
				"service-c": {
					Enabled: proto.Bool(true),
				},
			},
			expectedSecrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET":         {Value: &configv1.SecretValue_PlainText{PlainText: "leaf-override-root"}}, // Overridden
				"INTERMEDIATE_SECRET": {Value: &configv1.SecretValue_PlainText{PlainText: "intermediate-value"}},
			},
		},
		{
			name:    "unknown profile",
			profile: "unknown",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotConfigs, gotSecrets, err := m.ResolveProfile(tt.profile)
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.expectedConfigs, gotConfigs, protocmp.Transform()); diff != "" {
				t.Errorf("ResolveProfile() configs mismatch (-want +got):\n%s", diff)
			}
			if diff := cmp.Diff(tt.expectedSecrets, gotSecrets, protocmp.Transform()); diff != "" {
				t.Errorf("ResolveProfile() secrets mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestResolveProfile_DiamondInheritance(t *testing.T) {
	profiles := []*configv1.ProfileDefinition{
		{
			Name: proto.String("base"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-common": {
					Enabled: proto.Bool(true),
				},
			},
		},
		{
			Name:             proto.String("mixin1"),
			ParentProfileIds: []string{"base"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-1": {
					Enabled: proto.Bool(true),
				},
			},
		},
		{
			Name:             proto.String("mixin2"),
			ParentProfileIds: []string{"base"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-2": {
					Enabled: proto.Bool(true),
				},
			},
		},
		{
			Name:             proto.String("final"),
			ParentProfileIds: []string{"mixin1", "mixin2"},
		},
	}

	m := NewManager(profiles)

	// This should succeed, gathering configs from base, mixin1, and mixin2.
	configs, _, err := m.ResolveProfile("final")
	if err != nil {
		t.Fatalf("ResolveProfile('final') failed with error: %v", err)
	}

	expectedConfigs := map[string]*configv1.ProfileServiceConfig{
		"service-common": {Enabled: proto.Bool(true)},
		"service-1":      {Enabled: proto.Bool(true)},
		"service-2":      {Enabled: proto.Bool(true)},
	}

	if diff := cmp.Diff(expectedConfigs, configs, protocmp.Transform()); diff != "" {
		t.Errorf("ResolveProfile() configs mismatch (-want +got):\n%s", diff)
	}
}

func TestGetProfileDefinition(t *testing.T) {
	profiles := []*configv1.ProfileDefinition{
		{
			Name: proto.String("test-profile"),
		},
	}
	m := NewManager(profiles)

	// Test case 1: Profile exists
	p, ok := m.GetProfileDefinition("test-profile")
	if !ok {
		t.Error("GetProfileDefinition('test-profile') returned false, want true")
	}
	if p.GetName() != "test-profile" {
		t.Errorf("GetProfileDefinition('test-profile') returned name %s, want test-profile", p.GetName())
	}

	// Test case 2: Profile does not exist
	_, ok = m.GetProfileDefinition("non-existent")
	if ok {
		t.Error("GetProfileDefinition('non-existent') returned true, want false")
	}
}

func TestResolveProfile_Errors(t *testing.T) {
	tests := []struct {
		name       string
		profiles   []*configv1.ProfileDefinition
		target     string
		wantErrStr string
	}{
		{
			name: "real cycle detected",
			profiles: []*configv1.ProfileDefinition{
				{
					Name:             proto.String("A"),
					ParentProfileIds: []string{"B"},
				},
				{
					Name:             proto.String("B"),
					ParentProfileIds: []string{"A"},
				},
			},
			target:     "A",
			wantErrStr: "cycle detected in profile inheritance: A",
		},
		{
			name: "parent not found",
			profiles: []*configv1.ProfileDefinition{
				{
					Name:             proto.String("child"),
					ParentProfileIds: []string{"missing-parent"},
				},
			},
			target:     "child",
			wantErrStr: "parent profile not found: missing-parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewManager(tt.profiles)
			_, _, err := m.ResolveProfile(tt.target)
			if err == nil {
				t.Error("ResolveProfile() expected error, got nil")
				return
			}
			if !strings.Contains(err.Error(), tt.wantErrStr) {
				t.Errorf("ResolveProfile() error = %v, want error containing %q", err, tt.wantErrStr)
			}
		})
	}
}
