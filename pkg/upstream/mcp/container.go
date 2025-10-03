/*
 * Copyright 2025 Author(s) of MCPX
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

package mcp

import (
	"strings"
)

// DefaultNodeImage is the default Docker image for Node.js based commands.
const DefaultNodeImage = "node:lts-alpine"

// DefaultPythonImage is the default Docker image for Python based commands.
const DefaultPythonImage = "python:3-slim"

// DefaultAlpineImage is the default fallback Docker image.
const DefaultAlpineImage = "alpine:latest"

// GetContainerImageForCommand determines the appropriate container image to use for a given stdio command.
// It uses a set of predefined rules to select an image based on the command name.
func GetContainerImageForCommand(command string) string {
	switch {
	case command == "npx" || command == "npm" || command == "node":
		return DefaultNodeImage
	case strings.HasPrefix(command, "python"):
		return DefaultPythonImage
	default:
		return DefaultAlpineImage
	}
}
