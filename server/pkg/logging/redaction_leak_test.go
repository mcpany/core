// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package logging_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/mcpany/core/server/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func TestTextFormatDeepRedactionLeak(t *testing.T) {
	// This test confirms that the Text handler NO LONGER leaks secrets nested in structs.
	logging.ForTestsOnlyResetLogger()
	var buf bytes.Buffer
	logging.Init(slog.LevelInfo, &buf, "text")
	logger := logging.GetLogger()

	type SecretStruct struct {
		ApiKey string `json:"api_key"`
	}
	data := SecretStruct{ApiKey: "secret123"}

	// This mimics how one might log a request or configuration object
	logger.Info("config loaded", "data", data)

	output := buf.String()

	// We expect the secret to be redacted.
	assert.NotContains(t, output, "secret123", "Text handler should not leak nested secrets")
    assert.Contains(t, output, "[REDACTED]", "Text handler should contain redacted placeholder")
}
