package profile

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
)

func TestResolveProfile(t *testing.T) {
	profiles := []*configv1.ProfileDefinition{
		configv1.ProfileDefinition_builder{
			Name: proto.String("root"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-a": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
			Secrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET": configv1.SecretValue_builder{
					PlainText: proto.String("root-value"),
				}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name:             proto.String("intermediate"),
			ParentProfileIds: []string{"root"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-b": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
				// Override service-a partially
				"service-a": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
			Secrets: map[string]*configv1.SecretValue{
				"INTERMEDIATE_SECRET": configv1.SecretValue_builder{
					PlainText: proto.String("intermediate-value"),
				}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name:             proto.String("leaf"),
			ParentProfileIds: []string{"intermediate"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-c": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
				// Disable service-b
				"service-b": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(false),
				}.Build(),
			},
			Secrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET": configv1.SecretValue_builder{
					PlainText: proto.String("leaf-override-root"),
				}.Build(),
			},
		}.Build(),
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
				"service-a": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
			expectedSecrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET": configv1.SecretValue_builder{
					PlainText: proto.String("root-value"),
				}.Build(),
			},
		},
		{
			name:    "resolve intermediate",
			profile: "intermediate",
			expectedConfigs: map[string]*configv1.ProfileServiceConfig{
				"service-a": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
				"service-b": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
			expectedSecrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET":         configv1.SecretValue_builder{PlainText: proto.String("root-value")}.Build(),
				"INTERMEDIATE_SECRET": configv1.SecretValue_builder{PlainText: proto.String("intermediate-value")}.Build(),
			},
		},
		{
			name:    "resolve leaf",
			profile: "leaf",
			expectedConfigs: map[string]*configv1.ProfileServiceConfig{
				"service-a": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
				"service-b": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(false), // Overridden to false
				}.Build(),
				"service-c": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
			expectedSecrets: map[string]*configv1.SecretValue{
				"ROOT_SECRET":         configv1.SecretValue_builder{PlainText: proto.String("leaf-override-root")}.Build(), // Overridden
				"INTERMEDIATE_SECRET": configv1.SecretValue_builder{PlainText: proto.String("intermediate-value")}.Build(),
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
		configv1.ProfileDefinition_builder{
			Name: proto.String("base"),
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-common": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name:             proto.String("mixin1"),
			ParentProfileIds: []string{"base"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-1": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name:             proto.String("mixin2"),
			ParentProfileIds: []string{"base"},
			ServiceConfig: map[string]*configv1.ProfileServiceConfig{
				"service-2": configv1.ProfileServiceConfig_builder{
					Enabled: proto.Bool(true),
				}.Build(),
			},
		}.Build(),
		configv1.ProfileDefinition_builder{
			Name:             proto.String("final"),
			ParentProfileIds: []string{"mixin1", "mixin2"},
		}.Build(),
	}

	m := NewManager(profiles)

	// This should succeed, gathering configs from base, mixin1, and mixin2.
	configs, _, err := m.ResolveProfile("final")
	if err != nil {
		t.Fatalf("ResolveProfile('final') failed with error: %v", err)
	}

	expectedConfigs := map[string]*configv1.ProfileServiceConfig{
		"service-common": configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
		"service-1":      configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
		"service-2":      configv1.ProfileServiceConfig_builder{Enabled: proto.Bool(true)}.Build(),
	}

	if diff := cmp.Diff(expectedConfigs, configs, protocmp.Transform()); diff != "" {
		t.Errorf("ResolveProfile() configs mismatch (-want +got):\n%s", diff)
	}
}

func TestGetProfileDefinition(t *testing.T) {
	profiles := []*configv1.ProfileDefinition{
		configv1.ProfileDefinition_builder{
			Name: proto.String("test-profile"),
		}.Build(),
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
				configv1.ProfileDefinition_builder{
					Name:             proto.String("A"),
					ParentProfileIds: []string{"B"},
				}.Build(),
				configv1.ProfileDefinition_builder{
					Name:             proto.String("B"),
					ParentProfileIds: []string{"A"},
				}.Build(),
			},
			target:     "A",
			wantErrStr: "cycle detected in profile inheritance: A",
		},
		{
			name: "parent not found",
			profiles: []*configv1.ProfileDefinition{
				configv1.ProfileDefinition_builder{
					Name:             proto.String("child"),
					ParentProfileIds: []string{"missing-parent"},
				}.Build(),
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
