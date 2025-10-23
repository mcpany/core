/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package framework

import (
	"time"

	"github.com/mcpxy/core/tests/integration"
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
