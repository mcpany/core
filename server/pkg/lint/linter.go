// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package lint provides configuration analysis tools.
//
// Summary: Package lint provides configuration analysis tools.
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
// Summary: Represents a Severity level.
type Severity int

const (
	// Error indicates a critical issue.
	// Summary: Defines Error.
	Error Severity = iota
	// Warning indicates a potential issue.
	// Summary: Defines Warning.
	Warning
	// Info indicates a suggestion.
	// Summary: Defines Info.
	Info
)

// String returns the string representation of the severity.
//
// Summary: Executes String operation.
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
// Summary: Represents a Result.
type Result struct {
	// Severity indicates how critical the finding is.
	Severity Severity
	// ServiceName is the name of the service.
	ServiceName string
	// Message is the descriptive text.
	Message string
	// Path is the location in the configuration.
	Path string
}

// String returns the human-readable representation of the result.
//
// Summary: Executes String operation.
//
// Parameters:
//   - r (Result): The result instance to format.
//
// Returns:
//   - string: A formatted string.
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
//
// Summary: Represents a Linter.
type Linter struct {
	cfg *configv1.McpAnyServerConfig
}

// NewLinter creates a new Linter instance.
//
// Summary: Initializes NewLinter operation.
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
// Summary: Executes Run operation.
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

// checkPlainTextSecrets checks for secrets stored in plain text.
//
// Summary: Executes checkPlainTextSecrets operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of findings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
				Message:     "Secret is stored in plain text.",
				Path:        path,
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
				path := fmt.Sprintf("command_line_service.env[%s]", k)
				checkSecret(v, path, s.GetName())
			}
			if ce := cmd.GetContainerEnvironment(); ce != nil {
				for k, v := range ce.GetEnv() {
					path := "command_line_service." +
						"container_environment.env[%s]"
					checkSecret(v, fmt.Sprintf(path, k), s.GetName())
				}
			}
		case configv1.UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			switch mcp.WhichConnectionType() {
			case configv1.McpUpstreamService_StdioConnection_case:
				stdio := mcp.GetStdioConnection()
				for k, v := range stdio.GetEnv() {
					path := fmt.Sprintf("mcp_service.stdio.env[%s]", k)
					checkSecret(v, path, s.GetName())
				}
			case configv1.McpUpstreamService_BundleConnection_case:
				bundle := mcp.GetBundleConnection()
				for k, v := range bundle.GetEnv() {
					path := fmt.Sprintf("mcp_service.bundle.env[%s]", k)
					checkSecret(v, path, s.GetName())
				}
			}
		}
	}

	return results
}

// checkShellInjection checks for shell injection risks.
//
// Summary: Executes checkShellInjection operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of findings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) checkShellInjection() []Result {
	var results []Result
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
						Message:     fmt.Sprintf("Command uses %q.", pattern),
						Path:        "command",
					})
				}
			}
		}
	}
	return results
}

// checkInsecureHTTP checks for insecure HTTP connections.
//
// Summary: Executes checkInsecureHTTP operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of findings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) checkInsecureHTTP() []Result {
	var results []Result
	for _, s := range l.cfg.GetUpstreamServices() {
		checkInsecure := func(url, path string) {
			if url != "" && strings.HasPrefix(strings.ToLower(url), "http://") {
				isLocal := strings.Contains(url, "localhost") ||
					strings.Contains(url, "127.0.0.1")
				if !isLocal {
					results = append(results, Result{
						Severity:    Warning,
						ServiceName: s.GetName(),
						Message:     fmt.Sprintf("Insecure HTTP: %q.", url),
						Path:        path,
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
				path := "mcp_service.http_connection.http_address"
				checkInsecure(mcp.GetHttpConnection().GetHttpAddress(), path)
			}
		}
	}
	return results
}

// checkCacheSettings checks for cache settings.
//
// Summary: Executes checkCacheSettings operation.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of findings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
				Message:     "Cache has 0 TTL.",
				Path:        "cache.ttl",
			})
		}
	}
	return results
}
