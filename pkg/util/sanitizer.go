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

package util

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	disallowedCharsPattern = regexp.MustCompile(`[^a-zA-Z0-9_-]`)
	placeholder            = "_"
)

// SanitizeToolName cleans a tool name by replacing any disallowed characters with underscores.
// This is to ensure that tool names can be used in environments that have restrictions on characters,
// such as Prometheus metrics. The function is safe for concurrent use.
func SanitizeToolName(id string) (string, error) {
	if id == "" {
		return "", fmt.Errorf("id cannot be empty")
	}
	sanitized := disallowedCharsPattern.ReplaceAllString(id, placeholder)
	sanitized = strings.ReplaceAll(sanitized, "-", placeholder)
	return sanitized, nil
}
