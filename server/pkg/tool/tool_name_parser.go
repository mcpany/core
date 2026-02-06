// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/consts"
)

// ParseToolName deconstructs a fully qualified tool name into its namespace
// (service ID) and the bare tool name. The expected format is
// "<namespace>/-/--<tool_name>". If the separator is not found, the entire
// input is treated as the tool name with an empty namespace.
//
// toolName is the fully qualified tool name to parse.
// It returns the namespace, the bare tool name, and an error if the tool name
// is invalid (e.g., empty).
//
// Parameters:
//   - toolName: string.
//
// Returns:
//   - string:
//   - string:
//   - error:
func ParseToolName(toolName string) (namespace string, tool string, err error) {
	namespace, tool, found := strings.Cut(toolName, consts.ToolNameServiceSeparator)
	if !found {
		tool = namespace
		namespace = ""
	}

	tool = strings.TrimPrefix(tool, "--")

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
//
// Parameters:
//   - methodName: string.
//
// Returns:
//   - string:
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
