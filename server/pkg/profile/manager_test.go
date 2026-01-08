// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package profile

import (
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
