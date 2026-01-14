// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package lint

import (
	"context"
	"fmt"
	"strings"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/config"
)

// Severity indicates the importance of a linting result.
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
type Result struct {
	Severity    Severity
	ServiceName string
	Message     string
	Path        string // Config path context (e.g., "upstream_services[0].auth")
}

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
func NewLinter(cfg *configv1.McpAnyServerConfig) *Linter {
	return &Linter{cfg: cfg}
}

// Run executes all linting checks.
func (l *Linter) Run(ctx context.Context) ([]Result, error) {
	var results []Result

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
		if _, ok := sv.Value.(*configv1.SecretValue_PlainText); ok {
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
