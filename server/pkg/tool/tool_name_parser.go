// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/consts"
)

// ParseToolName deconstructs a fully qualified tool name into its namespace (service ID) and the bare tool name.
//
// Summary: Deconstructs a fully qualified tool name into its namespace (service ID) and the bare tool name.
//
// Parameters:
//   - toolName (string): Parameter.
//
// Returns:
//   - string: Return value.
//   - string: Return value.
//   - error: Return value.
//
// Errors:
//   - error: If an error occurs.
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

// GetFullyQualifiedToolName constructs a fully qualified tool name from a service ID and a method name.
//
// Summary: Constructs a fully qualified tool name from a service ID and a method name.
//
// Parameters:
//   - serviceID (string): Parameter.
//   - methodName (string): Parameter.
//
// Returns:
//   - string: Return value.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.

func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
