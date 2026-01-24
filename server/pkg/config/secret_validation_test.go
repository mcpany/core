package config

import (
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
)

func TestValidateSecret_EnvVar(t *testing.T) {
	os.Unsetenv("MISSING_VAR")
	secret := &configv1.SecretValue{
		Value: &configv1.SecretValue_EnvironmentVariable{
			EnvironmentVariable: "MISSING_VAR",
		},
	}
	err := validateSecretValue(secret)
	if err == nil {
		t.Error("Expected error for missing env var")
	}

	os.Setenv("EXISTING_VAR", "value")
	secret.Value = &configv1.SecretValue_EnvironmentVariable{
		EnvironmentVariable: "EXISTING_VAR",
	}
	err = validateSecretValue(secret)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
