// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/consts"
)

// Summary: Deconstructs a fully qualified tool name into its namespace.
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

// Summary: Constructs a fully qualified tool name from a.
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
