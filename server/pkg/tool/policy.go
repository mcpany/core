// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"regexp"

	"github.com/mcpany/core/server/pkg/logging"
	configv1 "github.com/mcpany/core/proto/config/v1"
)

// ShouldExport determines whether a named item (tool, prompt, or resource) should be exported
// based on the provided ExportPolicy.
//
// Parameters:
//   - name: string.
//   - policy: *configv1.ExportPolicy.
//
// Returns:
//   - bool:
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
//
// Parameters:
//   - policies: []*configv1.CallPolicy.
//   - callID: string.
//   - arguments: []byte.
//
// Returns:
//   - bool:
//   - error:
func EvaluateCallPolicy(policies []*configv1.CallPolicy, toolName, callID string, arguments []byte) (bool, error) {
	// Fallback to slower implementation if not using compiled policies
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

// compiledCallPolicyRule holds the pre-compiled regexes for a policy rule.
type compiledCallPolicyRule struct {
	nameRegex     *regexp.Regexp
	callIDRegex   *regexp.Regexp
	argumentRegex *regexp.Regexp
	rule          *configv1.CallPolicyRule
}

// CompiledCallPolicy holds a compiled version of a call policy.
type CompiledCallPolicy struct {
	policy        *configv1.CallPolicy
	compiledRules []compiledCallPolicyRule
}

// CompileCallPolicies compiles a list of call policies.
//
// policies is the policies.
//
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//   - policies: []*configv1.CallPolicy.
//
// Returns:
//   - []*CompiledCallPolicy:
//   - error:
func CompileCallPolicies(policies []*configv1.CallPolicy) ([]*CompiledCallPolicy, error) {
	compiled := make([]*CompiledCallPolicy, 0, len(policies))
	for _, p := range policies {
		if p == nil {
			continue
		}
		cp, err := NewCompiledCallPolicy(p)
		if err != nil {
			return nil, err
		}
		compiled = append(compiled, cp)
	}
	return compiled, nil
}

// NewCompiledCallPolicy compiles a single call policy.
//
// policy is the policy.
//
// Returns the result.
// Returns an error if the operation fails.
//
// Parameters:
//   - policy: *configv1.CallPolicy.
//
// Returns:
//   - *CompiledCallPolicy:
//   - error:
func NewCompiledCallPolicy(policy *configv1.CallPolicy) (*CompiledCallPolicy, error) {
	compiledRules := make([]compiledCallPolicyRule, len(policy.GetRules()))
	for i, rule := range policy.GetRules() {
		var nameRe, callIDRe, argRe *regexp.Regexp
		var err error

		if rule.GetNameRegex() != "" {
			nameRe, err = regexp.Compile(rule.GetNameRegex())
			if err != nil {
				return nil, fmt.Errorf("invalid tool name regex %q: %w", rule.GetNameRegex(), err)
			}
		}

		if rule.GetCallIdRegex() != "" {
			callIDRe, err = regexp.Compile(rule.GetCallIdRegex())
			if err != nil {
				return nil, fmt.Errorf("invalid call ID regex %q: %w", rule.GetCallIdRegex(), err)
			}
		}

		if rule.GetArgumentRegex() != "" {
			argRe, err = regexp.Compile(rule.GetArgumentRegex())
			if err != nil {
				return nil, fmt.Errorf("invalid argument regex %q: %w", rule.GetArgumentRegex(), err)
			}
		}

		compiledRules[i] = compiledCallPolicyRule{
			nameRegex:     nameRe,
			callIDRegex:   callIDRe,
			argumentRegex: argRe,
			rule:          rule,
		}
	}
	return &CompiledCallPolicy{
		policy:        policy,
		compiledRules: compiledRules,
	}, nil
}

// EvaluateCompiledCallPolicy checks if a call should be allowed based on the compiled policies.
//
// policies is the policies.
// toolName is the toolName.
// callID is the callID.
// arguments is the arguments.
//
// Returns true if successful.
// Returns an error if the operation fails.
//
// Parameters:
//   - policies: []*CompiledCallPolicy.
//   - callID: string.
//   - arguments: []byte.
//
// Returns:
//   - bool:
//   - error:
func EvaluateCompiledCallPolicy(policies []*CompiledCallPolicy, toolName, callID string, arguments []byte) (bool, error) {
	for _, policy := range policies {
		policyBlocked := false
		matchedRule := false
		for _, cRule := range policy.compiledRules {
			matched := true
			rule := cRule.rule

			if rule.GetNameRegex() != "" {
				if cRule.nameRegex == nil || !cRule.nameRegex.MatchString(toolName) {
					matched = false
				}
			}
			if matched && rule.GetCallIdRegex() != "" {
				if cRule.callIDRegex == nil || !cRule.callIDRegex.MatchString(callID) {
					matched = false
				}
			}
			if matched && rule.GetArgumentRegex() != "" {
				if arguments == nil {
					matched = false
				} else if cRule.argumentRegex == nil || !cRule.argumentRegex.Match(arguments) {
					matched = false
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
			if policy.policy.GetDefaultAction() == configv1.CallPolicy_DENY {
				policyBlocked = true
			}
		}

		if policyBlocked {
			return false, nil
		}
	}
	return true, nil
}
