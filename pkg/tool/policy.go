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

// EvaluateCallPolicy checks if a call should be allowed based on the policies.
// If arguments is nil, it performs a static check (ignoring rules with argument_regex).
// It returns true if the call is allowed, false otherwise.
func EvaluateCallPolicy(policies []*configv1.CallPolicy, toolName, callID string, arguments []byte) (bool, error) {
	for _, policy := range policies {
		if policy == nil {
			continue
		}
		policyBlocked := false
		matchedRule := false
		for _, rule := range policy.GetRules() {
			matched := true
			if rule.GetNameRegex() != "" {
				if matchedTool, _ := regexp.MatchString(rule.GetNameRegex(), toolName); !matchedTool {
					matched = false
				}
			}
			if matched && rule.GetCallIdRegex() != "" {
				if matchedCall, _ := regexp.MatchString(rule.GetCallIdRegex(), callID); !matchedCall {
					matched = false
				}
			}
			if matched && rule.GetArgumentRegex() != "" {
				if arguments == nil {
					// Cannot match argument regex at registration time
					matched = false
				} else {
					if matchedArgs, _ := regexp.Match(rule.GetArgumentRegex(), arguments); !matchedArgs {
						matched = false
					}
				}
			}

			if matched {
				matchedRule = true
				if rule.GetAction() == configv1.CallPolicy_DENY {
					policyBlocked = true
				}
				break // First match wins
			}
		}
		if !matchedRule {
			if policy.GetDefaultAction() == configv1.CallPolicy_DENY {
				policyBlocked = true
			}
		}

		if policyBlocked {
			return false, nil // Blocked
		}
	}
	return true, nil // Allowed
}
