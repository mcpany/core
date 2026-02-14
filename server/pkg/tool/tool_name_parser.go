// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/consts"
)

// ParseToolName deconstructs a fully qualified tool name into its namespace
// (service ID) and the bare tool name.
//
// The expected format is "<namespace>.<tool_name>".
//
// Summary: Parses a fully qualified tool name.
//
// Parameters:
//   - toolName: string. The fully qualified tool name to parse.
//
// Returns:
//   - namespace: string. The extracted namespace (service ID).
//   - tool: string. The extracted bare tool name.
//   - err: error. An error if the tool name is invalid.
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
// Summary: Constructs a fully qualified tool name.
//
// Parameters:
//   - serviceID: string. The unique identifier of the service.
//   - methodName: string. The name of the tool/method within the service.
//
// Returns:
//   - string: The combined, fully qualified tool name.
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
