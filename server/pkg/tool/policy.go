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

// PolicyResult represents the result of a policy evaluation.
type PolicyResult struct {
	Allowed         bool
	RequireApproval bool
}

// EvaluateCallPolicy checks if a call should be allowed based on the policies.
// If arguments is nil, it performs a static check (ignoring rules with argument_regex).
// It returns true if the call is allowed, false otherwise.
func EvaluateCallPolicy(policies []*configv1.CallPolicy, toolName, callID string, arguments []byte) (bool, error) {
	result, err := EvaluateCallPolicyWithResult(policies, toolName, callID, arguments)
	if err != nil {
		return false, err
	}
	// For backward compatibility, REQUIRE_APPROVAL is treated as "not allowed" (blocked)
	// by this function, but the caller should use EvaluateCallPolicyWithResult for more details.
	// Actually, if it requires approval, it's not "Allowed" in the sense of "proceed immediately".
	if result.RequireApproval {
		return false, nil
	}
	return result.Allowed, nil
}

// EvaluateCallPolicyWithResult checks if a call should be allowed based on the policies.
func EvaluateCallPolicyWithResult(policies []*configv1.CallPolicy, toolName, callID string, arguments []byte) (PolicyResult, error) {
	// Fallback to slower implementation if not using compiled policies
	for _, policy := range policies {
		if policy == nil {
			continue
		}

		matchedAction := policy.GetDefaultAction()

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
				matchedAction = rule.GetAction()
				break // First match wins
			}
		}

		if matchedAction == configv1.CallPolicy_DENY {
			return PolicyResult{Allowed: false}, nil
		}
		if matchedAction == configv1.CallPolicy_REQUIRE_APPROVAL {
			return PolicyResult{Allowed: false, RequireApproval: true}, nil
		}
	}
	return PolicyResult{Allowed: true}, nil
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
func EvaluateCompiledCallPolicy(policies []*CompiledCallPolicy, toolName, callID string, arguments []byte) (bool, error) {
	result, err := EvaluateCompiledCallPolicyWithResult(policies, toolName, callID, arguments)
	if err != nil {
		return false, err
	}
	if result.RequireApproval {
		return false, nil
	}
	return result.Allowed, nil
}

// EvaluateCompiledCallPolicyWithResult checks if a call should be allowed based on the compiled policies.
func EvaluateCompiledCallPolicyWithResult(policies []*CompiledCallPolicy, toolName, callID string, arguments []byte) (PolicyResult, error) {
	for _, policy := range policies {
		matchedAction := policy.policy.GetDefaultAction()

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
				matchedAction = rule.GetAction()
				break // First match wins
			}
		}

		if matchedAction == configv1.CallPolicy_DENY {
			return PolicyResult{Allowed: false}, nil
		}
		if matchedAction == configv1.CallPolicy_REQUIRE_APPROVAL {
			return PolicyResult{Allowed: false, RequireApproval: true}, nil
		}
	}
	return PolicyResult{Allowed: true}, nil
}
