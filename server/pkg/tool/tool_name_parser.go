// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/consts"
)

// ParseToolName deconstructs a fully qualified tool name into its namespace.
//
// Summary: deconstructs a fully qualified tool name into its namespace.
//
// Parameters:
//   - toolName: string. The toolName.
//
// Returns:
//   - namespace: string. The string.
//   - tool: string. The string.
//   - err: error. An error if the operation fails.
//
// Throws/Errors:
//   Returns an error if the operation fails.
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

// GetFullyQualifiedToolName constructs a fully qualified tool name from a.
//
// Summary: constructs a fully qualified tool name from a.
//
// Parameters:
//   - serviceID: string. The serviceID.
//   - methodName: string. The methodName.
//
// Returns:
//   - string: The string.
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
