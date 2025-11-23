// Copyright (C) 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"time"

	"github.com/mcpany/core/tests/integration"
)

// Re-exporting these from the integration package so that framework users
// don't need to import both.
var (
	FindFreePort          = integration.FindFreePort
	NewManagedProcess     = integration.NewManagedProcess
	WaitForTCPPort        = integration.WaitForTCPPort
	GetProjectRoot        = integration.GetProjectRoot
	ServiceStartupTimeout = 15 * time.Second
)

type AITool interface {
	Install()
	AddMCP(name, endpoint string)
	RemoveMCP(name string)
	Run(apiKey, model, prompt string) (string, error)
}
