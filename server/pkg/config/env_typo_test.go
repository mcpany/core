package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestEnvVarTypoSuggestion(t *testing.T) {
	tests := []struct {
		name           string
		targetEnv      string
		setEnv         string // The actual env var to set in OS
		setVal         string
		shouldSuggest  bool
		suggestionPart string
	}{
		{
			name:           "Typo_Suffix",
			targetEnv:      "OPENAI_API_KEY",
			setEnv:         "OPENAI_API_KEY_123",
			setVal:         "val",
			shouldSuggest:  true,
			suggestionPart: `Did you mean "OPENAI_API_KEY_123"?`,
		},
		{
			name:           "Typo_Insertion",
			targetEnv:      "DATABASE_URL",
			setEnv:         "DTABASE_URL",
			setVal:         "val",
			shouldSuggest:  true,
			suggestionPart: `Did you mean "DTABASE_URL"?`,
		},
		{
			name:          "No_Similar_Env",
			targetEnv:     "TOTALLY_UNIQUE_VAR",
			setEnv:        "SOMETHING_ELSE_ENTIRELY",
			setVal:        "val",
			shouldSuggest: false,
		},
		{
			name:          "Short_Target_Ignored",
			targetEnv:     "AB",
			setEnv:        "ABC", // Distance is 1, but target < 3
			setVal:        "val",
			shouldSuggest: false,
		},
		{
			name:           "Long_Target_Threshold_Cap",
			// Length 22. Threshold = 22/3 = 7. Cap at 5.
			// Distance 4 should match.
			targetEnv:      "VERY_LONG_ENVIRONMENT_VARIABLE_NAME",
			setEnv:         "VERY_LONG_ENVIRONMENT_VARIABLE_NAM", // Distance 1
			setVal:         "val",
			shouldSuggest:  true,
			suggestionPart: `Did you mean "VERY_LONG_ENVIRONMENT_VARIABLE_NAM"?`,
		},
		{
			name:           "Long_Target_Threshold_Cap_Exceeded",
			// Length 22. Threshold capped at 5.
			// If distance is 6, it should NOT match.
			targetEnv:      "VERY_LONG_ENVIRONMENT_VARIABLE_NAME",
			setEnv:         "VERY_LONG_ENV_VAR_NAME_CHANGE", // Distance > 5
			setVal:         "val",
			shouldSuggest:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv != "" {
				os.Setenv(tt.setEnv, tt.setVal)
				defer os.Unsetenv(tt.setEnv)
			}
			// Ensure target is unset (in case it leaks from env)
			os.Unsetenv(tt.targetEnv)

			secret := configv1.SecretValue_builder{
				EnvironmentVariable: proto.String(tt.targetEnv),
			}.Build()

			err := validateSecretValue(context.Background(), secret)
			assert.Error(t, err)

			ae, ok := err.(*ActionableError)
			assert.True(t, ok, "Error should be ActionableError")

			if tt.shouldSuggest {
				assert.Contains(t, ae.Suggestion, tt.suggestionPart)
			} else {
				// Should not contain "Did you mean"
				assert.NotContains(t, ae.Suggestion, "Did you mean")
			}
		})
	}
}

func TestFindSimilarEnvVar_Coverage(t *testing.T) {
	// Direct test for findSimilarEnvVar to ensure we hit branches like logic for empty parts or exact matches
	// although exact match shouldn't happen in the main flow if os.LookupEnv returns false.

	// We can't easily mock os.Environ() without changing the code structure or using a variable.
	// But we can verify behavior with existing env.

	// Case: Empty return
	res := findSimilarEnvVar("NON_EXISTENT_VAR_WITH_NO_MATCHES_123456789")
	assert.Equal(t, "", res)
}
