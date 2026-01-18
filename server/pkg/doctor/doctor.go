// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

// Package doctor provides functionality for checking the health and status of upstream services.
package doctor

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // Register MySQL driver
	_ "github.com/lib/pq"              // Register Postgres driver
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/validation"
	_ "modernc.org/sqlite" // Register SQLite driver
)

// Status represents the status of a check.
type Status string

const (
	// StatusOk indicates the check passed.
	StatusOk Status = "OK"
	// StatusWarning indicates a partial failure or non-critical issue.
	StatusWarning Status = "WARNING"
	// StatusError indicates a critical failure.
	StatusError Status = "ERROR"
	// StatusSkipped indicates the check was skipped.
	StatusSkipped Status = "SKIPPED"
)

// CheckResult represents the result of a single service check.
type CheckResult struct {
	ServiceName string
	Status      Status
	Message     string
	Error       error
}

// RunChecks performs connectivity and health checks on the provided configuration.
//
// ctx is the context for the request.
// config holds the configuration settings.
//
// Returns the result.
func RunChecks(ctx context.Context, config *configv1.McpAnyServerConfig) []CheckResult {
	// Using 'services' variable to support existing loop
	services := config.GetUpstreamServices()
	results := make([]CheckResult, 0, len(services))

	// Check upstream services
	for _, service := range services {
		if service.GetDisable() {
			results = append(results, CheckResult{
				ServiceName: service.GetName(),
				Status:      StatusSkipped,
				Message:     "Service is disabled",
			})
			continue
		}

		res := checkService(ctx, service)
		res.ServiceName = service.GetName()
		results = append(results, res)
	}

	return results
}

func checkService(ctx context.Context, service *configv1.UpstreamServiceConfig) CheckResult {
	// 5 second timeout for checks
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var authMsg string

	// Check authentication first if present
	if auth := service.GetUpstreamAuth(); auth != nil {
		authRes := checkAuthentication(ctx, auth)
		if authRes.Status != StatusOk {
			return CheckResult{
				Status:  authRes.Status,
				Message: fmt.Sprintf("Auth check failed: %s", authRes.Message),
				Error:   authRes.Error,
			}
		}
		authMsg = fmt.Sprintf("Auth [%s]. ", authRes.Message)
	}

	var res CheckResult
	switch service.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_HttpService_case:
		res = checkHTTPService(ctx, service.GetHttpService())
	case configv1.UpstreamServiceConfig_GrpcService_case:
		res = checkGRPCService(ctx, service.GetGrpcService())
	case configv1.UpstreamServiceConfig_OpenapiService_case:
		res = checkOpenAPIService(ctx, service.GetOpenapiService())
	case configv1.UpstreamServiceConfig_SqlService_case:
		res = checkSQLService(ctx, service.GetSqlService())
	case configv1.UpstreamServiceConfig_GraphqlService_case:
		res = checkGraphQLService(ctx, service.GetGraphqlService())
	case configv1.UpstreamServiceConfig_McpService_case:
		res = checkMCPService(ctx, service.GetMcpService())
	case configv1.UpstreamServiceConfig_CommandLineService_case:
		res = checkCommandLineService(ctx, service.GetCommandLineService())
	case configv1.UpstreamServiceConfig_WebsocketService_case:
		res = checkWebSocketService(ctx, service.GetWebsocketService())
	case configv1.UpstreamServiceConfig_WebrtcService_case:
		res = checkWebRTCService(ctx, service.GetWebrtcService())
	case configv1.UpstreamServiceConfig_FilesystemService_case:
		res = checkFilesystemService(ctx, service.GetFilesystemService())
	default:
		res = CheckResult{
			Status:  StatusSkipped,
			Message: "No check implementation for this service type",
		}
	}

	if authMsg != "" {
		res.Message = authMsg + res.Message
	}
	return res
}

func checkAuthentication(ctx context.Context, auth *configv1.Authentication) CheckResult {
	switch auth.WhichAuthMethod() {
	case configv1.Authentication_Oauth2_case:
		return checkOAuth2Reachability(ctx, auth.GetOauth2())
	case configv1.Authentication_Oidc_case:
		return checkOIDCReachability(ctx, auth.GetOidc())
	}
	// Other auth methods (API Key, Basic, etc.) don't have a separate endpoint to check
	// (they are checked implicitly by the service call, but we don't do a full auth'd call in doctor yet).
	return CheckResult{Status: StatusOk, Message: "Verified"}
}

func checkOAuth2Reachability(ctx context.Context, oauth *configv1.OAuth2Auth) CheckResult {
	tokenURL := oauth.GetTokenUrl()
	if tokenURL == "" {
		// Should have been caught by validator, but safe check
		return CheckResult{Status: StatusError, Message: "OAuth2 token_url is empty"}
	}

	// Attempt a POST request to the token URL.
	// We don't necessarily need to send valid credentials to check reachability.
	// Sending an empty POST or one with dummy grant_type usually triggers a 400 Bad Request
	// with a JSON error, which confirms the server is listening and is an OAuth endpoint.

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return CheckResult{Status: StatusError, Message: fmt.Sprintf("Failed to create request: %v", err), Error: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Failed to connect to OAuth2 token URL: %v", err),
			Error:   err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	// Analyze response code
	if resp.StatusCode == 200 {
		return CheckResult{Status: StatusOk, Message: "OAuth2 Reachable (200)"}
	}

	// 400 Bad Request is VERY common for OAuth2 endpoints when sending incomplete credentials.
	// This confirms the endpoint exists and is processing OAuth requests.
	if resp.StatusCode == 400 {
		return CheckResult{Status: StatusOk, Message: "OAuth2 Reachable (400)"}
	}

	// 401 Unauthorized is also a good sign of reachability (WAF or strict auth).
	if resp.StatusCode == 401 {
		return CheckResult{Status: StatusOk, Message: "OAuth2 Reachable (401)"}
	}

	if resp.StatusCode == 404 {
		return CheckResult{Status: StatusError, Message: fmt.Sprintf("OAuth2 token URL not found (404): %s", tokenURL)}
	}

	if resp.StatusCode >= 500 {
		return CheckResult{Status: StatusError, Message: fmt.Sprintf("OAuth2 token URL returned server error: %s", resp.Status)}
	}

	// Other 4xx codes?
	return CheckResult{Status: StatusWarning, Message: fmt.Sprintf("OAuth2 token URL returned unexpected status: %s", resp.Status)}
}

func checkOIDCReachability(ctx context.Context, oidc *configv1.OIDCAuth) CheckResult {
	issuer := oidc.GetIssuer()
	if issuer == "" {
		return CheckResult{Status: StatusError, Message: "OIDC issuer is empty"}
	}

	// OIDC discovery endpoint
	discoveryURL := strings.TrimSuffix(issuer, "/") + "/.well-known/openid-configuration"

	return checkURL(ctx, discoveryURL)
}

func checkHTTPService(ctx context.Context, s *configv1.HttpUpstreamService) CheckResult {
	return checkURL(ctx, s.GetAddress())
}

func checkGraphQLService(ctx context.Context, s *configv1.GraphQLUpstreamService) CheckResult {
	return checkURL(ctx, s.GetAddress())
}

func checkWebRTCService(ctx context.Context, s *configv1.WebrtcUpstreamService) CheckResult {
	return checkURL(ctx, s.GetAddress())
}

func checkWebSocketService(ctx context.Context, s *configv1.WebsocketUpstreamService) CheckResult {
	// For WebSocket, we can try to dial the TCP connection, or do an HTTP request if it supports upgrade.
	// Since checkURL assumes HTTP/HTTPS, we might need to handle ws/wss explicitly.
	addr := s.GetAddress()
	if strings.HasPrefix(addr, "ws://") {
		addr = "http://" + strings.TrimPrefix(addr, "ws://")
	} else if strings.HasPrefix(addr, "wss://") {
		addr = "https://" + strings.TrimPrefix(addr, "wss://")
	}

	// Try a simple HTTP GET. Most WS servers will respond to HTTP GET (usually with Upgrade required or 404/200).
	// If the server is up, we should get a response.
	return checkURL(ctx, addr)
}

func checkURL(ctx context.Context, urlStr string) CheckResult {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Invalid URL: %v", err),
			Error:   err,
		}
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSHandshakeTimeout: 5 * time.Second,
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Failed to connect: %v", err),
			Error:   err,
		}
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 && resp.StatusCode != 404 && resp.StatusCode != 405 && resp.StatusCode != 426 { // 426 Upgrade Required is fine for WS
		// We consider 4xx a warning because the service is technically reachable, just maybe not at this path.
		// However, 5xx is an error.
		if resp.StatusCode >= 500 {
			return CheckResult{
				Status:  StatusError,
				Message: fmt.Sprintf("Server returned error: %s", resp.Status),
			}
		}
		return CheckResult{
			Status:  StatusWarning,
			Message: fmt.Sprintf("Service reachable but returned: %s", resp.Status),
		}
	}

	return CheckResult{
		Status:  StatusOk,
		Message: fmt.Sprintf("Service reachable (%s)", resp.Status),
	}
}

func checkGRPCService(ctx context.Context, s *configv1.GrpcUpstreamService) CheckResult {
	// Basic TCP check for gRPC address
	host, port, err := net.SplitHostPort(s.GetAddress())
	if err != nil {
		// Try to see if it's just missing a port, though gRPC usually needs one.
		// If using a scheme like dns:///, it might be different.
		// For now, assume host:port
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Invalid gRPC address format: %v", err),
			Error:   err,
		}
	}

	timeout := 5 * time.Second
	if deadline, ok := ctx.Deadline(); ok {
		timeout = time.Until(deadline)
	}

	d := &net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Failed to connect to gRPC endpoint: %v", err),
			Error:   err,
		}
	}
	defer func() { _ = conn.Close() }()

	return CheckResult{
		Status:  StatusOk,
		Message: "TCP connection successful",
	}
}

func checkOpenAPIService(ctx context.Context, s *configv1.OpenapiUpstreamService) CheckResult {
	if s.GetSpecUrl() != "" {
		// Check if we can fetch the spec
		res := checkURL(ctx, s.GetSpecUrl())
		if res.Status != StatusOk {
			return res
		}
	}

	if s.GetAddress() != "" {
		return checkURL(ctx, s.GetAddress())
	}

	return CheckResult{
		Status:  StatusOk,
		Message: "OpenAPI definition seems accessible",
	}
}

func checkSQLService(ctx context.Context, s *configv1.SqlUpstreamService) CheckResult {
	// We need to resolve secrets in DSN if they exist, but that's handled by `util.ResolveSecret`
	// Since we don't have access to the secret store easily here without app initialization,
	// this might be tricky if the DSN relies on secret injection (e.g. ${SECRET_NAME}).
	// For now, we assume the DSN is raw or we'd need to mock the secret resolver.
	// Given this is a simple "Doctor", we might skip full DSN parsing if it looks like a template.

	if strings.Contains(s.GetDsn(), "${") {
		return CheckResult{
			Status:  StatusWarning,
			Message: "Cannot validate SQL connection with secret variables in DSN",
		}
	}

	db, err := sql.Open(s.GetDriver(), s.GetDsn())
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Failed to initialize SQL driver: %v", err),
			Error:   err,
		}
	}
	defer func() { _ = db.Close() }()

	// Try to ping
	err = db.PingContext(ctx)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Failed to ping database: %v", err),
			Error:   err,
		}
	}

	return CheckResult{
		Status:  StatusOk,
		Message: "Database connection successful",
	}
}

func checkMCPService(ctx context.Context, s *configv1.McpUpstreamService) CheckResult {
	switch s.WhichConnectionType() {
	case configv1.McpUpstreamService_HttpConnection_case:
		return checkURL(ctx, s.GetHttpConnection().GetHttpAddress())
	case configv1.McpUpstreamService_StdioConnection_case:
		// Check if command exists
		cmd := s.GetStdioConnection().GetCommand()
		// Reuse validation logic if possible, or just LookPath
		_, err := exec.LookPath(cmd)
		if err != nil {
			return CheckResult{
				Status:  StatusError,
				Message: fmt.Sprintf("Command not found: %s", cmd),
				Error:   err,
			}
		}
		return CheckResult{
			Status:  StatusOk,
			Message: "Command executable found",
		}
	default:
		return CheckResult{
			Status:  StatusSkipped,
			Message: "Unknown MCP connection type",
		}
	}
}

func checkCommandLineService(_ context.Context, s *configv1.CommandLineUpstreamService) CheckResult {
	if s.GetContainerEnvironment().GetImage() != "" {
		return CheckResult{
			Status:  StatusSkipped,
			Message: "Skipping containerized command check",
		}
	}

	cmd := s.GetCommand()
	// Check executable
	// We might need to handle args? "python script.py" -> check python and script.py
	// But GetCommand usually is just the executable or the full line.
	// If it's a full line, we need to split.
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return CheckResult{
			Status:  StatusError,
			Message: "Empty command",
		}
	}
	executable := parts[0]

	_, err := exec.LookPath(executable)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Command not found: %s", executable),
			Error:   err,
		}
	}

	// Maybe check if script file exists if provided?
	if len(parts) > 1 {
		arg := parts[1]
		// Simple heuristic: if it has an extension and exists, good.
		if validation.FileExists(arg) == nil {
			return CheckResult{
				Status:  StatusOk,
				Message: fmt.Sprintf("Executable and script found (%s %s)", executable, arg),
			}
		}
	}

	return CheckResult{
		Status:  StatusOk,
		Message: "Command executable found",
	}
}

func checkFilesystemService(_ context.Context, s *configv1.FilesystemUpstreamService) CheckResult {
	for vPath, hostPath := range s.GetRootPaths() {
		if err := validation.FileExists(hostPath); err != nil {
			return CheckResult{
				Status:  StatusError,
				Message: fmt.Sprintf("Root path %q -> %q not found or inaccessible: %v", vPath, hostPath, err),
				Error:   err,
			}
		}
	}
	return CheckResult{
		Status:  StatusOk,
		Message: "All root paths exist",
	}
}
