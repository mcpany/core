package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// The instruction implies adding "google.golang.org/protobuf/proto" again,
// but it's already present. The provided snippet for the change is also
// syntactically incorrect. Assuming the intent was to ensure the import
// is present, and it already is, no change is strictly needed.
// However, to faithfully follow the instruction's literal `Code Edit`
// which shows an additional `) "google.golang.org/protobuf/proto")`,
// this would result in a syntax error.
// Given the instruction "Add proto import" and the fact that
// `google.golang.org/protobuf/proto` is already imported, I will
// assume the instruction was to ensure its presence and no actual
// modification is required.
// If the intent was to add a *different* proto import, the instruction
// or snippet would need to be more specific.
// Since the instruction is "Add proto import" and the import
// `google.golang.org/protobuf/proto` is already in the file,
// I will not make any changes to the import block.
// The provided `Code Edit` snippet seems to be malformed or
// indicates a redundant addition that would cause a syntax error.

func TestRegexTrimValidationBug(t *testing.T) {
	// Setup: plain_text value with spaces
	config := func() *configv1.McpAnyServerConfig {
		secret := configv1.SecretValue_builder{
			PlainText:       proto.String(" value "),
			ValidationRegex: proto.String("^value$"),
		}.Build()

		stdio := configv1.McpStdioConnection_builder{
			Command: proto.String("ls"),
			Env: map[string]*configv1.SecretValue{
				"TEST_TRIM": secret,
			},
		}.Build()

		mcp := configv1.McpUpstreamService_builder{
			StdioConnection: stdio,
		}.Build()

		svc := configv1.UpstreamServiceConfig_builder{
			Name:       proto.String("test-trim-bug"),
			McpService: mcp,
		}.Build()

		return configv1.McpAnyServerConfig_builder{
			UpstreamServices: []*configv1.UpstreamServiceConfig{svc},
		}.Build()
	}()

	// This assumes that execLookPath is mocked or "ls" exists.
	// We need to mock execLookPath to avoid dependency on system "ls" or PATH.
	// But `validator.go` uses `execLookPath` var which we can override in test package.

	oldLookPath := execLookPath
	defer func() { execLookPath = oldLookPath }()
	execLookPath = func(file string) (string, error) {
		return "/bin/ls", nil
	}

	errs := Validate(context.Background(), config, Server)

	// Expectation: It should pass because validation logic now trims the value.
	assert.Empty(t, errs, "Validation errors not expected")
}
