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
	// Error indicates a critical issue that must be fixed.
	Error Severity = iota
	// Warning indicates a potential issue or best practice violation.
	Warning
	// Info indicates a suggestion or informational message.
	Info
)

// String returns the string representation of the severity.
//
// Parameters:
//   - s (Severity): The severity level to convert.
//
// Returns:
//   - string: The string representation (e.g., "ERROR").
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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

// String returns the human-readable representation of the result.
//
// Parameters:
//   - r (Result): The result instance to format.
//
// Returns:
//   - string: A formatted string containing result details.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
type Linter struct {
	cfg *configv1.McpAnyServerConfig
}

// NewLinter creates a new Linter instance.
//
// Parameters:
//   - cfg (*configv1.McpAnyServerConfig): The configuration to analyze.
//
// Returns:
//   - *Linter: A new Linter instance.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func NewLinter(cfg *configv1.McpAnyServerConfig) *Linter {
	return &Linter{cfg: cfg}
}

// Run executes all configured linting checks.
//
// Parameters:
//   - ctx (context.Context): The context for the operation.
//
// Returns:
//   - []Result: A slice of linting findings.
//   - error: Encounters a fatal issue.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) Run(ctx context.Context) ([]Result, error) {
	results := make([]Result, 0, 10)

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
				Message:     "Secret is stored in plain text. Use environment variables for better security.",
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

		if cmd := s.GetCommandLineService(); cmd != nil {
			for k, v := range cmd.GetEnv() {
				checkSecret(v, fmt.Sprintf("command_line_service.env[%s]", k), s.GetName())
			}
		}

		if mcp := s.GetMcpService(); mcp != nil {
			if stdio := mcp.GetStdioConnection(); stdio != nil {
				for k, v := range stdio.GetEnv() {
					checkSecret(v, fmt.Sprintf("mcp_service.stdio.env[%s]", k), s.GetName())
				}
			}
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
						Message:     fmt.Sprintf("Command uses shell invocation (%q).", pattern),
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
			path = "openapi_service.address"
		}

		if url != "" && strings.HasPrefix(strings.ToLower(url), "http://") {
			if !strings.Contains(url, "localhost") && !strings.Contains(url, "127.0.0.1") {
				results = append(results, Result{
					Severity:    Warning,
					ServiceName: s.GetName(),
					Message:     fmt.Sprintf("Service uses insecure HTTP connection to %q.", url),
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
		if s.GetCache() == nil {
			continue
		}

		if s.GetCache().GetTtl() == nil || s.GetCache().GetTtl().GetSeconds() == 0 {
			results = append(results, Result{
				Severity:    Info,
				ServiceName: s.GetName(),
				Message:     "Cache is configured but has 0 TTL.",
				Path:        "cache.ttl",
			})
		}
	}
	return results
}
