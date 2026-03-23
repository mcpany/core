import re
import os

path = "server/pkg/lint/linter.go"
with open(path, "r") as f:
    content = f.read()

def fix_block(match):
    block = match.group(0)
    # Check sections: Parameters, Returns, Errors, Side Effects
    sections = ["Parameters", "Returns", "Errors", "Side Effects"]
    for s in sections:
        if f"// {s}:" not in block:
            # Insert before the next section or the end
            # This is tricky with regex. Let's do it simply.
            pass
    return block

# Standard block structure:
# // Summary: ...
# //
# // Parameters:
# //   - ...
# //
# // Returns:
# //   - ...
# //
# // Errors:
# //   - ...
# //
# // Side Effects:
# //   - ...

def get_standard_block(summary, params=["None."], returns=["None."], errors=["None."], side_effects=["None."]):
    res = [f"// Summary: {summary}"]
    res.append("//")
    res.append("// Parameters:")
    for p in params: res.append(f"//   - {p}")
    res.append("//")
    res.append("// Returns:")
    for r in returns: res.append(f"//   - {r}")
    res.append("//")
    res.append("// Errors:")
    for e in errors: res.append(f"//   - {e}")
    res.append("//")
    res.append("// Side Effects:")
    for s in side_effects: res.append(f"//   - {s}")
    return "\n".join(res) + "\n"

# Manually rebuild linter.go with perfect blocks
new_content = """// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Summary: Package lint provides configuration analysis tools.
//
// Package lint provides functionality for analyzing configuration files
// to detect potential security issues and best practice violations.
package lint

import (
	"context"
	"fmt"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
)

// Summary: Represents a Severity level.
//
// Severity indicates the importance of a linting result.
//
// It is used to categorize findings based on their impact and urgency.
type Severity int

const (
	// Summary: Defines the Error severity level.
	//
	// Error indicates a critical issue that must be fixed.
	Error Severity = iota
	// Summary: Defines the Warning severity level.
	//
	// Warning indicates a potential issue or best practice violation.
	Warning
	// Summary: Defines the Info severity level.
	//
	// Info indicates a suggestion or informational message.
	Info
)

"""

new_content += get_standard_block("Executes String operation.", returns=["string: The string representation (e.g., \"ERROR\")."])
new_content += "func (s Severity) String() string {\n"
new_content += """	switch s {
	case Error:
		return "ERROR"
	case Warning:
		return "WARNING"
	case Info:
		return "INFO"
	default:
		return "UNKNOWN"
	}
}

// Summary: Represents a linting finding.
//
// Result represents a single linting finding.
//
// It encapsulates details about a detected issue.
type Result struct {
	// Severity indicates how critical the finding is.
	Severity Severity
	// ServiceName is the name of the service associated with the finding.
	ServiceName string
	// Message is the descriptive text of the finding.
	Message string
	// Path is the location in the configuration where the issue was found.
	Path string
}

"""

new_content += get_standard_block("Executes String operation.", returns=["string: A formatted string containing result details."])
new_content += "func (r Result) String() string {\n"
new_content += """	pathStr := ""
	if r.Path != "" {
		pathStr = fmt.Sprintf(" at %s", r.Path)
	}
	serviceStr := ""
	if r.ServiceName != "" {
		serviceStr = fmt.Sprintf(" (service: %s)", r.ServiceName)
	}
	return fmt.Sprintf("[%s]%s%s: %s", r.Severity, serviceStr, pathStr, r.Message)
}

// Summary: Represents a configuration linter.
//
// Linter performs static analysis on the configuration.
type Linter struct {
	cfg *configv1.McpAnyServerConfig
}

"""

new_content += get_standard_block("Initializes NewLinter operation.", params=["cfg (*configv1.McpAnyServerConfig): The configuration to analyze."], returns=["*Linter: A new Linter instance."])
new_content += "func NewLinter(cfg *configv1.McpAnyServerConfig) *Linter {\n"
new_content += """	return &Linter{cfg: cfg}
}

"""

new_content += get_standard_block("Executes Run operation.", params=["ctx (context.Context): The context for the operation."], returns=["[]Result: A slice of linting findings.", "error: Encounters a fatal issue."])
new_content += "func (l *Linter) Run(ctx context.Context) ([]Result, error) {\n"
new_content += """	results := make([]Result, 0, 10)

	validationErrors := config.Validate(ctx, l.cfg, config.Server)
	for _, err := range validationErrors {
		results = append(results, Result{
			Severity:    Error,
			ServiceName: err.ServiceName,
			Message:     err.Err.Error(),
		})
	}

	results = append(results, l.checkPlainTextSecrets()...)
	results = append(results, l.checkShellInjection()...)
	results = append(results, l.checkInsecureHTTP()...)
	results = append(results, l.checkCacheSettings()...)

	return results, nil
}

"""

new_content += get_standard_block("Executes checkPlainTextSecrets operation.", returns=["[]Result: A slice of findings."])
new_content += "func (l *Linter) checkPlainTextSecrets() []Result {\n"
new_content += """	var results []Result

	checkSecret := func(sv *configv1.SecretValue, path, serviceName string) {
		if sv == nil {
			return
		}
		if sv.WhichValue() == configv1.SecretValue_PlainText_case {
			results = append(results, Result{
				Severity:    Warning,
				ServiceName: serviceName,
				Message: "Secret is stored in plain text. Use environment " +
					"variables or file references for better security.",
				Path: path,
			})
		}
	}

	for _, s := range l.cfg.GetUpstreamServices() {
		if auth := s.GetUpstreamAuth(); auth != nil {
			switch auth.WhichAuthMethod() {
			case configv1.Authentication_ApiKey_case:
				checkSecret(auth.GetApiKey().GetValue(),
					"upstream_auth.api_key.value", s.GetName())
			case configv1.Authentication_BearerToken_case:
				checkSecret(auth.GetBearerToken().GetToken(),
					"upstream_auth.bearer_token.token", s.GetName())
			case configv1.Authentication_BasicAuth_case:
				checkSecret(auth.GetBasicAuth().GetPassword(),
					"upstream_auth.basic_auth.password", s.GetName())
			case configv1.Authentication_Oauth2_case:
				checkSecret(auth.GetOauth2().GetClientSecret(),
					"upstream_auth.oauth2.client_secret", s.GetName())
			}
		}

		switch s.WhichServiceConfig() {
		case configv1.UpstreamServiceConfig_CommandLineService_case:
			cmd := s.GetCommandLineService()
			for k, v := range cmd.GetEnv() {
				checkSecret(v, fmt.Sprintf("command_line_service.env[%s]", k),
					s.GetName())
			}
			if ce := cmd.GetContainerEnvironment(); ce != nil {
				for k, v := range ce.GetEnv() {
					checkSecret(v, fmt.Sprintf("command_line_service."+
						"container_environment.env[%s]", k), s.GetName())
				}
			}
		case configv1.UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			switch mcp.WhichConnectionType() {
			case configv1.McpUpstreamService_StdioConnection_case:
				stdio := mcp.GetStdioConnection()
				for k, v := range stdio.GetEnv() {
					checkSecret(v, fmt.Sprintf("mcp_service.stdio.env[%s]", k),
						s.GetName())
				}
			case configv1.McpUpstreamService_BundleConnection_case:
				bundle := mcp.GetBundleConnection()
				for k, v := range bundle.GetEnv() {
					checkSecret(v, fmt.Sprintf("mcp_service.bundle.env[%s]", k),
						s.GetName())
				}
			}
		}
	}

	return results
}

"""

new_content += get_standard_block("Executes checkShellInjection operation.", returns=["[]Result: A slice of findings."])
new_content += "func (l *Linter) checkShellInjection() []Result {\n"
new_content += """	var results []Result
	shellRiskPatterns := []string{"sh -c", "bash -c", "cmd /c", "powershell -c"}

	for _, s := range l.cfg.GetUpstreamServices() {
		var command string
		switch s.WhichServiceConfig() {
		case configv1.UpstreamServiceConfig_CommandLineService_case:
			command = s.GetCommandLineService().GetCommand()
		case configv1.UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			if mcp.WhichConnectionType() ==
				configv1.McpUpstreamService_StdioConnection_case {
				command = mcp.GetStdioConnection().GetCommand()
			}
		}

		if command != "" {
			for _, pattern := range shellRiskPatterns {
				if strings.Contains(strings.ToLower(command), pattern) {
					results = append(results, Result{
						Severity:    Warning,
						ServiceName: s.GetName(),
						Message: fmt.Sprintf("Command uses shell invocation "+
							"(%q). Ensure inputs are properly sanitized to "+
							"prevent shell injection.", pattern),
						Path: "command",
					})
				}
			}
		}
	}
	return results
}

"""

new_content += get_standard_block("Executes checkInsecureHTTP operation.", returns=["[]Result: A slice of findings."])
new_content += "func (l *Linter) checkInsecureHTTP() []Result {\n"
new_content += """	var results []Result
	for _, s := range l.cfg.GetUpstreamServices() {
		checkInsecure := func(url, path string) {
			if url != "" && strings.HasPrefix(strings.ToLower(url), "http://") {
				if !strings.Contains(url, "localhost") &&
					!strings.Contains(url, "127.0.0.1") {
					results = append(results, Result{
						Severity:    Warning,
						ServiceName: s.GetName(),
						Message: fmt.Sprintf("Service uses insecure HTTP "+
							"connection to %q. Consider using HTTPS.", url),
						Path: path,
					})
				}
			}
		}

		switch s.WhichServiceConfig() {
		case configv1.UpstreamServiceConfig_HttpService_case:
			checkInsecure(s.GetHttpService().GetAddress(),
				"http_service.address")
		case configv1.UpstreamServiceConfig_OpenapiService_case:
			openapi := s.GetOpenapiService()
			checkInsecure(openapi.GetAddress(), "openapi_service.address")
			checkInsecure(openapi.GetSpecUrl(), "openapi_service.spec_url")
		case configv1.UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			if mcp.WhichConnectionType() ==
				configv1.McpUpstreamService_HttpConnection_case {
				checkInsecure(mcp.GetHttpConnection().GetHttpAddress(),
					"mcp_service.http_connection.http_address")
			}
		}
	}
	return results
}

"""

new_content += get_standard_block("Executes checkCacheSettings operation.", returns=["[]Result: A slice of findings."])
new_content += "func (l *Linter) checkCacheSettings() []Result {\n"
new_content += """	var results []Result
	for _, s := range l.cfg.GetUpstreamServices() {
		if s.GetCache() == nil {
			continue
		}

		if s.GetCache().GetTtl() == nil || s.GetCache().GetTtl().GetSeconds() == 0 {
			results = append(results, Result{
				Severity:    Info,
				ServiceName: s.GetName(),
				Message: "Cache is configured but has 0 TTL (infinite or " +
					"disabled depending on implementation). Verify this is " +
					"intended.",
				Path: "cache.ttl",
			})
		}
	}
	return results
}
"""

with open(path, "w") as f:
    f.write(new_content)
