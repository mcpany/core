// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"regexp"

	"github.com/mcpany/core/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ShouldExport determines whether a named item (tool, prompt, or resource) should be exported
// based on the provided ExportPolicy.
func ShouldExport(name string, policy *configv1.ExportPolicy) bool {
	if policy == nil {
		// Default to Allow/Export if no policy is present?
		// Usually default is everything is exported unless restricted.
		return true
	}

	// Iterate strict rules first
	for _, rule := range policy.GetRules() {
		if rule.GetNameRegex() == "" {
			continue
		}
		matched, err := regexp.MatchString(rule.GetNameRegex(), name)
		if err != nil {
			logging.GetLogger().Error("Invalid regex in export policy", "regex", rule.GetNameRegex(), "error", err)
			continue
		}
		if matched {
			return rule.GetAction() == configv1.ExportPolicy_EXPORT
		}
	}

	// Check default action
	if policy.GetDefaultAction() == configv1.ExportPolicy_UNEXPORT {
		return false
	}
	// EXPORT or UNSPECIFIED -> Export
	return true
}
