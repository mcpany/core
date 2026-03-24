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
// Summary: Returns the string representation of the severity.
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
// Summary: Represents a linting finding.
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
// Summary: Returns the human-readable representation of the result.
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
//
// Summary: Performs static analysis on the configuration.
type Linter struct {
	cfg *configv1.McpAnyServerConfig
}

// NewLinter creates a new Linter instance.
//
// Summary: Creates a new Linter instance.
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
// Summary: Executes all configured linting checks.
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

	results = append(results, l.checkSecrets()...)
	results = append(results, l.checkShellInjection()...)
	results = append(results, l.checkInsecureHTTP()...)
	results = append(results, l.checkCacheSettings()...)

	return results, nil
}

// checkSecrets checks for secrets stored in plain text.
//
// Summary: Checks for secrets stored in plain text.
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
func (l *Linter) checkSecrets() []Result {
	var results []Result

	for _, s := range l.cfg.GetUpstreamServices() {
		results = append(results, l.lintServiceAuthSecrets(s)...)
		results = append(results, l.lintServiceEnvSecrets(s)...)
	}

	return results
}

// lintServiceAuthSecrets lints authentication secrets.
//
// Summary: Lints authentication secrets.
//
// Parameters:
//   - s (*configv1.UpstreamServiceConfig): The service.
//
// Returns:
//   - []Result: A slice of findings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) lintServiceAuthSecrets(
	s *configv1.UpstreamServiceConfig,
) []Result {
	var results []Result
	auth := s.GetUpstreamAuth()
	if auth == nil {
		return nil
	}

	var sv *configv1.SecretValue
	var path string

	switch auth.WhichAuthMethod() {
	case configv1.Authentication_ApiKey_case:
		sv = auth.GetApiKey().GetValue()
		path = "upstream_auth.api_key.value"
	case configv1.Authentication_BearerToken_case:
		sv = auth.GetBearerToken().GetToken()
		path = "upstream_auth.bearer_token.token"
	case configv1.Authentication_BasicAuth_case:
		sv = auth.GetBasicAuth().GetPassword()
		path = "upstream_auth.basic_auth.password"
	case configv1.Authentication_Oauth2_case:
		sv = auth.GetOauth2().GetClientSecret()
		path = "upstream_auth.oauth2.client_secret"
	}

	if sv != nil && sv.WhichValue() == configv1.SecretValue_PlainText_case {
		results = append(results, Result{
			Severity:    Warning,
			ServiceName: s.GetName(),
			Message:     "Secret is stored in plain text.",
			Path:        path,
		})
	}

	return results
}

// lintServiceEnvSecrets lints environment variable secrets.
//
// Summary: Lints environment variable secrets.
//
// Parameters:
//   - s (*configv1.UpstreamServiceConfig): The service.
//
// Returns:
//   - []Result: A slice of findings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) lintServiceEnvSecrets(
	s *configv1.UpstreamServiceConfig,
) []Result {
	var results []Result

	check := func(env map[string]*configv1.SecretValue, prefix string) {
		for k, v := range env {
			if v != nil && v.WhichValue() == configv1.SecretValue_PlainText_case {
				results = append(results, Result{
					Severity:    Warning,
					ServiceName: s.GetName(),
					Message:     "Secret is stored in plain text.",
					Path:        fmt.Sprintf("%s.env[%s]", prefix, k),
				})
			}
		}
	}

	switch s.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_CommandLineService_case:
		cmd := s.GetCommandLineService()
		check(cmd.GetEnv(), "command_line_service")
		if ce := cmd.GetContainerEnvironment(); ce != nil {
			check(ce.GetEnv(), "command_line_service.container_environment")
		}
	case configv1.UpstreamServiceConfig_McpService_case:
		mcp := s.GetMcpService()
		switch mcp.WhichConnectionType() {
		case configv1.McpUpstreamService_StdioConnection_case:
			check(mcp.GetStdioConnection().GetEnv(), "mcp_service.stdio")
		case configv1.McpUpstreamService_BundleConnection_case:
			check(mcp.GetBundleConnection().GetEnv(), "mcp_service.bundle")
		}
	}

	return results
}

// checkShellInjection checks for shell injection risks.
//
// Summary: Checks for shell injection risks.
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
	patterns := []string{"sh -c", "bash -c", "cmd /c", "powershell -c"}

	for _, s := range l.cfg.GetUpstreamServices() {
		var cmd string
		switch s.WhichServiceConfig() {
		case configv1.UpstreamServiceConfig_CommandLineService_case:
			cmd = s.GetCommandLineService().GetCommand()
		case configv1.UpstreamServiceConfig_McpService_case:
			mcp := s.GetMcpService()
			if mcp.WhichConnectionType() ==
				configv1.McpUpstreamService_StdioConnection_case {
				cmd = mcp.GetStdioConnection().GetCommand()
			}
		}

		if cmd == "" {
			continue
		}

		for _, p := range patterns {
			if strings.Contains(strings.ToLower(cmd), p) {
				results = append(results, Result{
					Severity:    Warning,
					ServiceName: s.GetName(),
					Message:     fmt.Sprintf("Command uses %q.", p),
					Path:        "command",
				})
			}
		}
	}
	return results
}

// checkInsecureHTTP checks for insecure HTTP connections.
//
// Summary: Checks for insecure HTTP connections.
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
		results = append(results, l.lintServiceHTTP(s)...)
	}
	return results
}

// lintServiceHTTP lints HTTP connections.
//
// Summary: Lints HTTP connections.
//
// Parameters:
//   - s (*configv1.UpstreamServiceConfig): The service.
//
// Returns:
//   - []Result: A slice of findings.
//
// Errors:
//   - None.
//
// Side Effects:
//   - None.
func (l *Linter) lintServiceHTTP(s *configv1.UpstreamServiceConfig) []Result {
	var results []Result
	check := func(url, path string) {
		if url == "" || !strings.HasPrefix(strings.ToLower(url), "http://") {
			return
		}
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

	switch s.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_HttpService_case:
		check(s.GetHttpService().GetAddress(), "http_service.address")
	case configv1.UpstreamServiceConfig_OpenapiService_case:
		openapi := s.GetOpenapiService()
		check(openapi.GetAddress(), "openapi_service.address")
		check(openapi.GetSpecUrl(), "openapi_service.spec_url")
	case configv1.UpstreamServiceConfig_McpService_case:
		mcp := s.GetMcpService()
		if mcp.WhichConnectionType() ==
			configv1.McpUpstreamService_HttpConnection_case {
			check(mcp.GetHttpConnection().GetHttpAddress(),
				"mcp_service.http_connection.http_address")
		}
	}
	return results
}

// checkCacheSettings checks for cache settings.
//
// Summary: Checks for cache settings.
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
