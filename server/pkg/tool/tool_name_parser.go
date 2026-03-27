// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"strings"

	"github.com/mcpany/core/server/pkg/consts"
)

// ParseToolName provides parsetoolname functionality.
//
// Summary: ParseToolName.
//
// Parameters.
//   - toolName: The parameter.
//   - tool: The parameter.
//   - err: The parameter.
//
// Returns.
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

// GetFullyQualifiedToolName provides getfullyqualifiedtoolname functionality.
//
// Summary: GetFullyQualifiedToolName.
//
// Parameters.
//   - serviceID: The parameter.
//   - methodName: The parameter.
//
// Returns.
//   - result: The result.
func GetFullyQualifiedToolName(serviceID, methodName string) string {
	return fmt.Sprintf("%s%s%s", serviceID, consts.ToolNameServiceSeparator, methodName)
}
