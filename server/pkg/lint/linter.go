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
// Summary: Represents the severity level of a linting result.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Severity int

const (
	// Error indicates a critical issue that must be fixed.
	//
	// Summary: Defines the Error severity level.
	Error Severity = iota
	// Warning indicates a potential issue or best practice violation.
	//
	// Summary: Defines the Warning severity level.
	Warning
	// Info indicates a suggestion or informational message.
	//
	// Summary: Defines the Info severity level.
	Info
)

// String returns the string representation of the severity.
//
// Summary: Converts Severity to a string.
//
// Parameters:
//   - s (Severity): The severity level.
//
// Returns:
//   - string: The string representation (e.g., "ERROR", "WARNING", "INFO").
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
// Summary: Encapsulates a single issue found during config analysis.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
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
// Summary: Returns a formatted string for the linting result.
//
// Parameters:
//   - r (Result): The result to format.
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
		serviceStr = fmt.Sprintf(" (svc: %s)", r.ServiceName)
	}
	return fmt.Sprintf("[%s]%s%s: %s",
		r.Severity, serviceStr, pathStr, r.Message)
}

// Linter performs static analysis on the configuration.
//
// Summary: A linter that scans configuration for security issues.
//
// Parameters:
//   - None.
//
// Returns:
//   - None.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
type Linter struct {
	cfg *configv1.McpAnyServerConfig
}

// NewLinter creates a new Linter instance.
//
// Summary: Initializes a new Linter.
//
// Parameters:
//   - cfg (*configv1.McpAnyServerConfig): The configuration to analyze.
//
// Returns:
//   - *Linter: A pointer to the initialized Linter.
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
// Summary: Executes the full suite of linting checks.
//
// Parameters:
//   - ctx (context.Context): Context for managing the analysis lifecycle.
//
// Returns:
//   - []Result: A slice containing all detected linting issues.
//   - error: An error if the analysis process itself fails.
//
// Errors:
//   - None.
//
// Side Effects:
//   - Performs multiple security and best-practice scans.
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
// Summary: Scans for plain-text secrets in the configuration.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of warnings for detected plain-text secrets.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) checkPlainTextSecrets() []Result {
	var results []Result

	checkSecret := func(sv *configv1.SecretValue, path, sName string) {
		if sv == nil {
			return
		}
		if sv.WhichValue() == configv1.SecretValue_PlainText_case {
			results = append(results, Result{
				Severity:    Warning,
				ServiceName: sName,
				Message: "Secret is stored in plain text. " +
					"Use env vars or file refs for " +
					"better security.",
				Path: path,
			})
		}
	}

	for _, s := range l.cfg.GetUpstreamServices() {
		sId := s.GetId()
		if auth := s.GetUpstreamAuth(); auth != nil {
			switch auth.WhichAuthMethod() {
			case configv1.Authentication_ApiKey_case:
				checkSecret(auth.GetApiKey().GetValue(),
					"upstream_auth.api_key.value", sId)
			case configv1.Authentication_BearerToken_case:
				checkSecret(auth.GetBearerToken().GetToken(),
					"upstream_auth.bearer_token.token", sId)
			case configv1.Authentication_BasicAuth_case:
				checkSecret(auth.GetBasicAuth().GetPassword(),
					"upstream_auth.basic_auth.password",
					sId)
			case configv1.Authentication_Oauth2_case:
				checkSecret(auth.GetOauth2().GetClientSecret(),
					"upstream_auth.oauth2.client_secret",
					sId)
			}
		}

		switch s.WhichServiceConfig() {
		case configv1.
			UpstreamServiceConfig_CommandLineService_case:
			cmd := s.GetCommandLineService()
			for k, v := range cmd.GetEnv() {
				p := fmt.Sprintf(
					"command_line_service.env[%s]", k)
				checkSecret(v, p, sId)
			}
			if ce := cmd.GetContainerEnvironment(); ce != nil {
				for k, v := range ce.GetEnv() {
					p := fmt.Sprintf(
						"command_line_service." +
							"container_env." +
							"env[%s]", k)
					checkSecret(v, p, sId)
				}
			}
		case configv1.UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			switch mcp.WhichConnectionType() {
			case configv1.McpUpstreamService_StdioConnection_case:
				stdio := mcp.GetStdioConnection()
				for k, v := range stdio.GetEnv() {
					p := fmt.Sprintf(
						"mcp_service.stdio.env[%s]", k)
					checkSecret(v, p, sId)
				}
			case configv1.McpUpstreamService_BundleConnection_case:
				bundle := mcp.GetBundleConnection()
				for k, v := range bundle.GetEnv() {
					p := fmt.Sprintf(
						"mcp_service.bundle.env[%s]", k)
					checkSecret(v, p, sId)
				}
			}
		}
	}

	return results
}

// checkShellInjection checks for shell injection risks in commands.
//
// Summary: Scans for shell injection patterns in configured commands.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of warnings for detected shell injection patterns.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) checkShellInjection() []Result {
	var results []Result
	shellRiskPatterns := []string{
		"sh -c", "bash -c", "cmd /c", "powershell -c"}

	for _, s := range l.cfg.GetUpstreamServices() {
		var command string
		switch s.WhichServiceConfig() {
		case configv1.
			UpstreamServiceConfig_CommandLineService_case:
			command = s.GetCommandLineService().GetCommand()
		case configv1.
			UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			if mcp.WhichConnectionType() == configv1.
				McpUpstreamService_StdioConnection_case {
				command = mcp.GetStdioConnection().GetCommand()
			}
		}

		if command != "" {
			lowCmd := strings.ToLower(command)
			for _, pattern := range shellRiskPatterns {
				if strings.Contains(lowCmd, pattern) {
					results = append(results, Result{
						Severity:    Warning,
						ServiceName: s.GetId(),
						Message: fmt.Sprintf(
							"Command uses "+
								"shell "+
								"invocation "+
								"(%q). "+
								"Ensure "+
								"inputs are "+
								"sanitized "+
								"to prevent "+
								"shell "+
								"injection.",
							pattern),
						Path: "command",
					})
				}
			}
		}
	}
	return results
}

// checkInsecureHTTP checks for insecure HTTP connections.
//
// Summary: Scans for insecure HTTP addresses in the configuration.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of warnings for any detected insecure HTTP.
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
			if url == "" {
				return
			}
			lowURL := strings.ToLower(url)
			if !strings.HasPrefix(lowURL, "http://") {
				return
			}
			if strings.Contains(url, "localhost") ||
				strings.Contains(url, "127.0.0.1") {
				return
			}
			results = append(results, Result{
				Severity:    Warning,
				ServiceName: s.GetId(),
				Message: fmt.Sprintf("Service uses insecure "+
					"HTTP connection to %q. Use HTTPS.",
					url),
				Path: path,
			})
		}

		switch s.WhichServiceConfig() {
		case configv1.UpstreamServiceConfig_HttpService_case:
			checkInsecure(s.GetHttpService().GetAddress(),
				"http_service.address")
		case configv1.UpstreamServiceConfig_OpenapiService_case:
			openapi := s.GetOpenapiService()
			checkInsecure(openapi.GetAddress(),
				"openapi_service.address")
			checkInsecure(openapi.GetSpecUrl(),
				"openapi_service.spec_url")
		case configv1.UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			if mcp.WhichConnectionType() == configv1.
				McpUpstreamService_HttpConnection_case {
				checkInsecure(mcp.GetHttpConnection().
					GetHttpAddress(),
					"mcp_service.http_connection."+
						"http_address")
			}
		}
	}
	return results
}

// checkCacheSettings checks for suspicious cache settings.
//
// Summary: Scans for problematic cache configurations.
//
// Parameters:
//   - None.
//
// Returns:
//   - []Result: A slice of info findings for questionable cache settings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) checkCacheSettings() []Result {
	var results []Result
	for _, s := range l.cfg.GetUpstreamServices() {
		cache := s.GetCache()
		if cache == nil {
			continue
		}

		ttl := cache.GetTtl()
		if ttl == nil || ttl.GetSeconds() == 0 {
			results = append(results, Result{
				Severity:    Info,
				ServiceName: s.GetId(),
				Message: "Cache is configured but has 0 " +
					"TTL (infinite or disabled). " +
					"Verify this is intended.",
				Path: "cache.ttl",
			})
		}
	}
	return results
}
