// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/consts"
)

// Summary: ParseToolName deconstructs a fully qualified tool name into its namespace (service ID) and the bare tool name. Parses a fully qualified tool name.
//
// Parameters:
//   - toolName (string): The toolName parameter.
//
// Returns:
//   - string: The resulting string.
//   - string: The resulting string.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.
//
// Side Effects:
//   - None.
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

// Summary: GetFullyQualifiedToolName constructs a fully qualified tool name from a service ID and a method name. Constructs a fully qualified tool name.
//
// Parameters:
//   - serviceID (string): The serviceID parameter.
//   - methodName (string): The methodName parameter.
//
// Returns:
//   - string: The resulting string.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
