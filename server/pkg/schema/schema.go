// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package schema contains the embedded JSON schema for configuration validation.
package schema

import (
	_ "embed"
)

// ConfigSchema is the JSON schema for the McpAnyServerConfig.
//
//go:embed config_schema.json
var ConfigSchema []byte
