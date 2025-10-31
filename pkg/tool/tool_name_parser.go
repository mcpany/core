/*
 * Copyright 2025 Author(s) of MCP Any
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

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/pkg/consts"
)

// ParseToolName deconstructs a fully qualified tool name into its namespace
// (service ID) and the bare tool name. The expected format is
// "<namespace>/-/--<tool_name>". If the separator is not found, the entire
// input is treated as the tool name with an empty namespace.
//
// toolName is the fully qualified tool name to parse.
// It returns the namespace, the bare tool name, and an error if the tool name
// is invalid (e.g., empty).
func ParseToolName(toolName string) (namespace string, tool string, err error) {
	parts := strings.SplitN(toolName, consts.ToolNameServiceSeparator, 2)
	if len(parts) == 2 {
		namespace = parts[0]
		tool = parts[1]
	} else {
		tool = parts[0]
	}

	if tool == "" || tool == "/" {
		return "", "", fmt.Errorf("invalid tool name: %s", toolName)
	}
	return namespace, tool, nil
}

// GetFullyQualifiedToolName constructs a fully qualified tool name from a
// service ID and a method name, using the standard separator.
//
// serviceID is the unique identifier of the service.
// methodName is the name of the tool/method within the service.
// It returns the combined, fully qualified tool name.
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
