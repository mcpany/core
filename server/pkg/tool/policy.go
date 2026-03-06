// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package tool

import (
	"fmt"
	"regexp"
	"sync"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
)

var exportRegexCache sync.Map

// ShouldExport determines whether a named item (tool, prompt, or resource) should be exported. Summary: Checks if an item should be exported based on policy. Parameters: - name: string. The name of the item. - policy: *configv1.ExportPolicy. The export policy to evaluate. Returns: - bool: True if the item should be exported, false otherwise.
//
// Summary: ShouldExport determines whether a named item (tool, prompt, or resource) should be exported. Summary: Checks if an item should be exported based on policy. Parameters: - name: string. The name of the item. - policy: *configv1.ExportPolicy. The export policy to evaluate. Returns: - bool: True if the item should be exported, false otherwise.
//
// Parameters:
//   - name (string): The name parameter used in the operation.
//   - policy (*configv1.ExportPolicy): The policy parameter used in the operation.
//
// Returns:
//   - (bool): A boolean indicating the success or status of the operation.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func ShouldExport(name string, policy *configv1.ExportPolicy) bool {
	if policy == nil {
		// Default to Allow/Export if no policy is present?
		// Usually default is everything is exported unless restricted.
		return true
	}

	// Iterate strict rules first
	for _, rule := range policy.GetRules() {
		pattern := rule.GetNameRegex()
		if pattern == "" {
			continue
		}

		// ⚡ BOLT: Cached regex compilation for ShouldExport to eliminate O(n) regex compilation overhead during frequent export policy evaluations.
		// Randomized Selection from Top 5 High-Impact Targets (CPU/Regex)
		var re *regexp.Regexp
		if cached, ok := exportRegexCache.Load(pattern); ok {
			re = cached.(*regexp.Regexp)
		} else {
			var err error
			re, err = regexp.Compile(pattern)
			if err != nil {
				logging.GetLogger().Error("Invalid regex in export policy", "regex", pattern, "error", err)
				continue
			}
			exportRegexCache.Store(pattern, re)
		}

		if re.MatchString(name) {
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

// EvaluateCallPolicy checks if a call should be allowed based on the policies. Summary: Evaluates call policies against a tool execution. If arguments is nil, it performs a static check (ignoring rules with argument_regex). It returns true if the call is allowed, false otherwise. Parameters: - policies: []*configv1.CallPolicy. The list of policies to evaluate. - toolName: string. The name of the tool being called. - callID: string. The unique ID of the call. - arguments: []byte. The arguments of the call (can be nil). Returns: - bool: True if the call is allowed, false otherwise. - error: An error if evaluation fails.
//
// Summary: EvaluateCallPolicy checks if a call should be allowed based on the policies. Summary: Evaluates call policies against a tool execution. If arguments is nil, it performs a static check (ignoring rules with argument_regex). It returns true if the call is allowed, false otherwise. Parameters: - policies: []*configv1.CallPolicy. The list of policies to evaluate. - toolName: string. The name of the tool being called. - callID: string. The unique ID of the call. - arguments: []byte. The arguments of the call (can be nil). Returns: - bool: True if the call is allowed, false otherwise. - error: An error if evaluation fails.
//
// Parameters:
//   - policies ([]*configv1.CallPolicy): The policies parameter used in the operation.
//   - _ (toolName): An unnamed parameter of type toolName.
//   - callID (string): The unique identifier used to reference the call resource.
//   - arguments ([]byte): The arguments parameter used in the operation.
//
// Returns:
//   - (bool): A boolean indicating the success or status of the operation.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
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

// CompiledCallPolicy - Auto-generated documentation.
//
// Summary: CompiledCallPolicy holds a compiled version of a call policy.
//
// Fields:
//   - Various fields for CompiledCallPolicy.
type CompiledCallPolicy struct {
	policy        *configv1.CallPolicy
	compiledRules []compiledCallPolicyRule
}

// CompileCallPolicies compiles a list of call policies into an efficient runtime format. Summary: Compiles call policies for runtime usage. Parameters: - policies: []*configv1.CallPolicy. The list of policies to compile. Returns: - []*CompiledCallPolicy: The compiled policies. - error: An error if compilation fails (e.g., invalid regex).
//
// Summary: CompileCallPolicies compiles a list of call policies into an efficient runtime format. Summary: Compiles call policies for runtime usage. Parameters: - policies: []*configv1.CallPolicy. The list of policies to compile. Returns: - []*CompiledCallPolicy: The compiled policies. - error: An error if compilation fails (e.g., invalid regex).
//
// Parameters:
//   - policies ([]*configv1.CallPolicy): The policies parameter used in the operation.
//
// Returns:
//   - ([]*CompiledCallPolicy): The resulting []CompiledCallPolicy object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
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

// NewCompiledCallPolicy compiles a single call policy. Summary: Compiles a single call policy. Parameters: - policy: *configv1.CallPolicy. The policy to compile. Returns: - *CompiledCallPolicy: The compiled policy. - error: An error if compilation fails.
//
// Summary: NewCompiledCallPolicy compiles a single call policy. Summary: Compiles a single call policy. Parameters: - policy: *configv1.CallPolicy. The policy to compile. Returns: - *CompiledCallPolicy: The compiled policy. - error: An error if compilation fails.
//
// Parameters:
//   - policy (*configv1.CallPolicy): The policy parameter used in the operation.
//
// Returns:
//   - (*CompiledCallPolicy): The resulting CompiledCallPolicy object containing the requested data.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
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

// EvaluateCompiledCallPolicy checks if a call should be allowed based on the compiled policies. Summary: Evaluates compiled call policies. Parameters: - policies: []*CompiledCallPolicy. The list of compiled policies to evaluate. - toolName: string. The name of the tool being called. - callID: string. The unique ID of the call. - arguments: []byte. The arguments of the call (can be nil). Returns: - bool: True if the call is allowed, false otherwise. - error: An error if evaluation fails.
//
// Summary: EvaluateCompiledCallPolicy checks if a call should be allowed based on the compiled policies. Summary: Evaluates compiled call policies. Parameters: - policies: []*CompiledCallPolicy. The list of compiled policies to evaluate. - toolName: string. The name of the tool being called. - callID: string. The unique ID of the call. - arguments: []byte. The arguments of the call (can be nil). Returns: - bool: True if the call is allowed, false otherwise. - error: An error if evaluation fails.
//
// Parameters:
//   - policies ([]*CompiledCallPolicy): The policies parameter used in the operation.
//   - _ (toolName): An unnamed parameter of type toolName.
//   - callID (string): The unique identifier used to reference the call resource.
//   - arguments ([]byte): The arguments parameter used in the operation.
//
// Returns:
//   - (bool): A boolean indicating the success or status of the operation.
//   - (error): An error object if the operation fails, otherwise nil.
//
// Errors:
//   - Returns an error if the underlying operation fails or encounters invalid input.
//
// Side Effects:
//   - None.
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
