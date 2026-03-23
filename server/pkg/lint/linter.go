// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

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

// Severity indicates the importance of a linting result.
//
// It is used to categorize findings based on their impact and urgency.
type Severity int

const (
	// Error indicates a critical issue that must be fixed for the system to function correctly or securely.
	Error Severity = iota
	// Warning indicates a potential issue or best practice violation that should be addressed.
	Warning
	// Info indicates a suggestion or informational message for optimization or clarity.
	Info
)

// String returns the string representation of the severity.
//
// It converts the Severity enum to its string counterpart (ERROR, WARNING, INFO).
//
// Returns:
//   - string: The string representation of the severity.
func (s Severity) String() string {
	switch s {
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

// Result represents a single linting finding.
//
// It encapsulates all details about a detected issue, including its severity, location, and description.
type Result struct {
	// Severity indicates how critical the finding is (Error, Warning, Info).
	Severity Severity
	// ServiceName is the name of the service associated with the finding, if any.
	ServiceName string
	// Message is the descriptive text of the finding.
	Message string
	// Path is the location in the configuration where the issue was found (e.g., "upstream_services[0].auth").
	Path string
}

// String returns the string representation of the result.
//
// It formats the result into a human-readable string suitable for CLI output.
//
// Returns:
//   - string: A formatted string containing severity, service, path, and message.
func (r Result) String() string {
	pathStr := ""
	if r.Path != "" {
		pathStr = fmt.Sprintf(" at %s", r.Path)
	}
	serviceStr := ""
	if r.ServiceName != "" {
		serviceStr = fmt.Sprintf(" (service: %s)", r.ServiceName)
	}
	return fmt.Sprintf("[%s]%s%s: %s", r.Severity, serviceStr, pathStr, r.Message)
}

// Linter performs static analysis on the configuration.
//
// It holds the configuration to be analyzed and provides methods to execute various checks.
type Linter struct {
	cfg *configv1.McpAnyServerConfig
}

// NewLinter creates a new Linter instance.
//
// Parameters:
//   - cfg: *configv1.McpAnyServerConfig. The server configuration to be linted.
//
// Returns:
//   - *Linter: A new Linter instance initialized with the provided configuration.
func NewLinter(cfg *configv1.McpAnyServerConfig) *Linter {
	return &Linter{cfg: cfg}
}

// Run executes all linting checks.
//
// It aggregates results from multiple check categories including standard validation,
// secret usage, shell injection risks, insecure HTTP, and cache settings.
//
// Parameters:
//   - ctx: context.Context. The context for the request (currently unused but reserved for future async checks).
//
// Returns:
//   - []Result: A list of linting findings.
//   - error: An error if the linting process encounters a fatal issue (currently always nil).
func (l *Linter) Run(ctx context.Context) ([]Result, error) {
	// Pre-allocate to avoid performance warnings, though initial size is a guess.
	results := make([]Result, 0, 10)

	// 1. Run standard validation first (Errors)
	validationErrors := config.Validate(ctx, l.cfg, config.Server)
	for _, err := range validationErrors {
		results = append(results, Result{
			Severity:    Error,
			ServiceName: err.ServiceName,
			Message:     err.Err.Error(),
		})
	}

	// 2. Check for Plain Text Secrets (Warning)
	results = append(results, l.checkPlainTextSecrets()...)

	// 3. Check for Shell Injection Risks (Warning)
	results = append(results, l.checkShellInjection()...)

	// 4. Check for Insecure HTTP (Warning)
	results = append(results, l.checkInsecureHTTP()...)

	// 5. Check for Missing Cache TTL (Info)
	results = append(results, l.checkCacheSettings()...)

	return results, nil
}

func (l *Linter) checkPlainTextSecrets() []Result {
	var results []Result

	checkSecret := func(sv *configv1.SecretValue, path, serviceName string) {
		if sv == nil {
			return
		}
		if sv.WhichValue() == configv1.SecretValue_PlainText_case {
			results = append(results, Result{
				Severity:    Warning,
				ServiceName: serviceName,
				Message:     "Secret is stored in plain text. Use environment variables or file references for better security.",
				Path:        path,
			})
		}
	}

	for _, s := range l.cfg.GetUpstreamServices() {
		if auth := s.GetUpstreamAuth(); auth != nil {
			if apiKey := auth.GetApiKey(); apiKey != nil {
				checkSecret(apiKey.GetValue(), "upstream_auth.api_key.value", s.GetName())
			}
			if bearer := auth.GetBearerToken(); bearer != nil {
				checkSecret(bearer.GetToken(), "upstream_auth.bearer_token.token", s.GetName())
			}
			if basic := auth.GetBasicAuth(); basic != nil {
				checkSecret(basic.GetPassword(), "upstream_auth.basic_auth.password", s.GetName())
			}
			if oauth := auth.GetOauth2(); oauth != nil {
				checkSecret(oauth.GetClientSecret(), "upstream_auth.oauth2.client_secret", s.GetName())
			}
		}

		// Check command env vars
		if cmd := s.GetCommandLineService(); cmd != nil {
			for k, v := range cmd.GetEnv() {
				checkSecret(v, fmt.Sprintf("command_line_service.env[%s]", k), s.GetName())
			}
			if ce := cmd.GetContainerEnvironment(); ce != nil {
				for k, v := range ce.GetEnv() {
					checkSecret(v, fmt.Sprintf("command_line_service.container_environment.env[%s]", k), s.GetName())
				}
			}
		}

		if mcp := s.GetMcpService(); mcp != nil {
			if stdio := mcp.GetStdioConnection(); stdio != nil {
				for k, v := range stdio.GetEnv() {
					checkSecret(v, fmt.Sprintf("mcp_service.stdio.env[%s]", k), s.GetName())
				}
			}
			if bundle := mcp.GetBundleConnection(); bundle != nil {
				for k, v := range bundle.GetEnv() {
					checkSecret(v, fmt.Sprintf("mcp_service.bundle.env[%s]", k), s.GetName())
				}
			}
		}
	}

	// Check users
	for _, u := range l.cfg.GetUsers() {
		if auth := u.GetAuthentication(); auth != nil {
			// Similar checks for user auth if applicable
			// (Assuming user auth structure mirrors upstream auth mostly or has secrets)
			// ...
			_ = auth
		}
	}

	return results
}

func (l *Linter) checkShellInjection() []Result {
	var results []Result
	shellRiskPatterns := []string{"sh -c", "bash -c", "cmd /c", "powershell -c"}

	for _, s := range l.cfg.GetUpstreamServices() {
		var command string
		if cmd := s.GetCommandLineService(); cmd != nil {
			command = cmd.GetCommand()
		} else if mcp := s.GetMcpService(); mcp != nil {
			if stdio := mcp.GetStdioConnection(); stdio != nil {
				command = stdio.GetCommand()
			}
		}

		if command != "" {
			for _, pattern := range shellRiskPatterns {
				if strings.Contains(strings.ToLower(command), pattern) {
					results = append(results, Result{
						Severity:    Warning,
						ServiceName: s.GetName(),
						Message:     fmt.Sprintf("Command uses shell invocation (%q). Ensure inputs are properly sanitized to prevent shell injection.", pattern),
						Path:        "command",
					})
				}
			}
		}
	}
	return results
}

func (l *Linter) checkInsecureHTTP() []Result {
	var results []Result
	for _, s := range l.cfg.GetUpstreamServices() {
		var url string
		var path string

		if http := s.GetHttpService(); http != nil {
			url = http.GetAddress()
			path = "http_service.address"
		} else if openapi := s.GetOpenapiService(); openapi != nil {
			url = openapi.GetAddress()
			if url == "" {
				url = openapi.GetSpecUrl()
				path = "openapi_service.spec_url"
			} else {
				path = "openapi_service.address"
			}
		} else if mcp := s.GetMcpService(); mcp != nil {
			if httpConn := mcp.GetHttpConnection(); httpConn != nil {
				url = httpConn.GetHttpAddress()
				path = "mcp_service.http_connection.http_address"
			}
		}

		if url != "" && strings.HasPrefix(strings.ToLower(url), "http://") {
			// Whitelist localhost/127.0.0.1
			if !strings.Contains(url, "localhost") && !strings.Contains(url, "127.0.0.1") {
				results = append(results, Result{
					Severity:    Warning,
					ServiceName: s.GetName(),
					Message:     fmt.Sprintf("Service uses insecure HTTP connection to %q. Consider using HTTPS.", url),
					Path:        path,
				})
			}
		}
	}
	return results
}

func (l *Linter) checkCacheSettings() []Result {
	var results []Result
	for _, s := range l.cfg.GetUpstreamServices() {
		// If cache is enabled but no TTL or default TTL
		// Logic here depends on how cache is defined.
		// `config.Validate` checks for negative TTL.
		// Here we want to warn if TTL is missing for HTTP services maybe?
		if s.GetCache() == nil {
			// No cache configured at all. Maybe okay.
			continue
		}

		if s.GetCache().GetTtl() == nil || s.GetCache().GetTtl().GetSeconds() == 0 {
			results = append(results, Result{
				Severity:    Info,
				ServiceName: s.GetName(),
				Message:     "Cache is configured but has 0 TTL (infinite or disabled depending on implementation). Verify this is intended.",
				Path:        "cache.ttl",
			})
		}
	}
	return results
}
