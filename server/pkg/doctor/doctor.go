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
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql" // Register MySQL driver
	_ "github.com/lib/pq"              // Register Postgres driver
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/pkg/validation"
	_ "modernc.org/sqlite" // Register SQLite driver
)

// Status represents the status of a check.
//
// Summary: represents the status of a check.
type Status string

const (
	// StatusOk indicates the check passed successfully.
	StatusOk Status = "OK"
	// StatusWarning indicates a partial failure or non-critical issue that should be investigated.
	StatusWarning Status = "WARNING"
	// StatusError indicates a critical failure that prevents the service from functioning correctly.
	StatusError Status = "ERROR"
	// StatusSkipped indicates the check was skipped, usually due to configuration (e.g., disabled service).
	StatusSkipped Status = "SKIPPED"
)

// CheckResult represents the result of a single service check.
//
// Summary: represents the result of a single service check.
type CheckResult struct {
	// ServiceName is the name of the service being checked.
	ServiceName string
	// Status is the outcome of the check (OK, WARNING, ERROR, SKIPPED).
	Status Status
	// Message provides human-readable details about the check result.
	Message string
	// Error contains the underlying error object if the check failed.
	Error error
}

// RunChecks performs connectivity and health checks on the provided configuration.
//
// Summary: performs connectivity and health checks on the provided configuration.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - config: *configv1.McpAnyServerConfig. The config.
//
// Returns:
//   - []CheckResult: The []CheckResult.
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

		res := CheckService(ctx, service)
		res.ServiceName = service.GetName()
		results = append(results, res)
	}

	return results
}

// CheckService performs a connectivity check for a single service.
//
// Summary: performs a connectivity check for a single service.
//
// Parameters:
//   - ctx: context.Context. The context for the operation.
//   - service: *configv1.UpstreamServiceConfig. The service.
//
// Returns:
//   - CheckResult: The CheckResult.
func CheckService(ctx context.Context, service *configv1.UpstreamServiceConfig) CheckResult {
	// 5 second timeout for checks
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var authMsg string
	var upstreamAuth *configv1.Authentication

	// Check authentication first if present
	if auth := service.GetUpstreamAuth(); auth != nil {
		upstreamAuth = auth
		// Only check OAuth2/OIDC reachability here, as simple API keys are checked during service connection
		if auth.GetOauth2() != nil || auth.GetOidc() != nil {
			authRes := checkAuthentication(ctx, auth)
			if authRes.Status != StatusOk {
				return CheckResult{
					Status:  authRes.Status,
					Message: fmt.Sprintf("Auth check failed: %s", authRes.Message),
					Error:   authRes.Error,
				}
			}
			authMsg = fmt.Sprintf("Auth Config [%s]. ", authRes.Message)
		}
	}

	var res CheckResult
	switch service.WhichServiceConfig() {
	case configv1.UpstreamServiceConfig_HttpService_case:
		res = checkHTTPService(ctx, service.GetHttpService(), upstreamAuth)
	case configv1.UpstreamServiceConfig_GrpcService_case:
		res = checkGRPCService(ctx, service.GetGrpcService())
	case configv1.UpstreamServiceConfig_OpenapiService_case:
		res = checkOpenAPIService(ctx, service.GetOpenapiService(), upstreamAuth)
	case configv1.UpstreamServiceConfig_SqlService_case:
		res = checkSQLService(ctx, service.GetSqlService())
	case configv1.UpstreamServiceConfig_GraphqlService_case:
		res = checkGraphQLService(ctx, service.GetGraphqlService(), upstreamAuth)
	case configv1.UpstreamServiceConfig_McpService_case:
		res = checkMCPService(ctx, service.GetMcpService(), upstreamAuth)
	case configv1.UpstreamServiceConfig_CommandLineService_case:
		res = checkCommandLineService(ctx, service.GetCommandLineService())
	case configv1.UpstreamServiceConfig_WebsocketService_case:
		res = checkWebSocketService(ctx, service.GetWebsocketService(), upstreamAuth)
	case configv1.UpstreamServiceConfig_WebrtcService_case:
		res = checkWebRTCService(ctx, service.GetWebrtcService(), upstreamAuth)
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
	// (they are checked implicitly by the service call).
	return CheckResult{Status: StatusOk, Message: "Verified"}
}

func checkOAuth2Reachability(ctx context.Context, oauth *configv1.OAuth2Auth) CheckResult {
	tokenURL := oauth.GetTokenUrl()
	if tokenURL == "" {
		// Should have been caught by validator, but safe check
		return CheckResult{Status: StatusError, Message: "OAuth2 token_url is empty"}
	}

	// Attempt a POST request to the token URL.
	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader("grant_type=client_credentials"))
	if err != nil {
		return CheckResult{Status: StatusError, Message: fmt.Sprintf("Failed to create request: %v", err), Error: err}
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := util.NewSafeHTTPClient()
	client.Timeout = 5 * time.Second

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

	if resp.StatusCode == 400 {
		return CheckResult{Status: StatusOk, Message: "OAuth2 Reachable (400)"}
	}

	if resp.StatusCode == 401 {
		return CheckResult{Status: StatusOk, Message: "OAuth2 Reachable (401)"}
	}

	if resp.StatusCode == 404 {
		return CheckResult{Status: StatusError, Message: fmt.Sprintf("OAuth2 token URL not found (404): %s", tokenURL)}
	}

	if resp.StatusCode >= 500 {
		return CheckResult{Status: StatusError, Message: fmt.Sprintf("OAuth2 token URL returned server error: %s", resp.Status)}
	}

	return CheckResult{Status: StatusWarning, Message: fmt.Sprintf("OAuth2 token URL returned unexpected status: %s", resp.Status)}
}

func checkOIDCReachability(ctx context.Context, oidc *configv1.OIDCAuth) CheckResult {
	issuer := oidc.GetIssuer()
	if issuer == "" {
		return CheckResult{Status: StatusError, Message: "OIDC issuer is empty"}
	}

	// OIDC discovery endpoint
	discoveryURL := strings.TrimSuffix(issuer, "/") + "/.well-known/openid-configuration"

	return checkURL(ctx, discoveryURL, nil)
}

func checkHTTPService(ctx context.Context, s *configv1.HttpUpstreamService, auth *configv1.Authentication) CheckResult {
	return checkURL(ctx, s.GetAddress(), auth)
}

func checkGraphQLService(ctx context.Context, s *configv1.GraphQLUpstreamService, auth *configv1.Authentication) CheckResult {
	return checkURL(ctx, s.GetAddress(), auth)
}

func checkWebRTCService(ctx context.Context, s *configv1.WebrtcUpstreamService, auth *configv1.Authentication) CheckResult {
	return checkURL(ctx, s.GetAddress(), auth)
}

func checkWebSocketService(ctx context.Context, s *configv1.WebsocketUpstreamService, auth *configv1.Authentication) CheckResult {
	// For WebSocket, we can try to dial the TCP connection, or do an HTTP request if it supports upgrade.
	addr := s.GetAddress()
	if strings.HasPrefix(addr, "ws://") {
		addr = "http://" + strings.TrimPrefix(addr, "ws://")
	} else if strings.HasPrefix(addr, "wss://") {
		addr = "https://" + strings.TrimPrefix(addr, "wss://")
	}

	return checkURL(ctx, addr, auth)
}

func checkURL(ctx context.Context, urlStr string, auth *configv1.Authentication) CheckResult {
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Invalid URL: %v", util.RedactDSN(err.Error())),
			Error:   err,
		}
	}

	// Apply authentication if provided
	if auth != nil {
		if err := applyAuthentication(ctx, req, auth); err != nil {
			return CheckResult{
				Status:  StatusError,
				Message: fmt.Sprintf("Failed to apply authentication: %v", err),
				Error:   err,
			}
		}
	}

	client := util.NewSafeHTTPClient()
	if transport, ok := client.Transport.(*http.Transport); ok {
		transport.TLSHandshakeTimeout = 5 * time.Second
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

	// If Auth was provided, 401/403 are errors
	if auth != nil && (resp.StatusCode == 401 || resp.StatusCode == 403) {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Authentication failed (%s): check your credentials", resp.Status),
		}
	}

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

	dialer := util.NewSafeDialer()
	// Check environment variables to allow unsafe connections if configured (consistent with other components)
	if os.Getenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS") == util.TrueStr || os.Getenv("MCPANY_ALLOW_LOOPBACK_RESOURCES") == util.TrueStr {
		dialer.AllowLoopback = true
		dialer.AllowPrivate = true
	}
	if os.Getenv("MCPANY_ALLOW_PRIVATE_NETWORK_RESOURCES") == util.TrueStr {
		dialer.AllowPrivate = true
	}
	dialer.Dialer = &net.Dialer{Timeout: timeout}

	conn, err := dialer.DialContext(ctx, "tcp", net.JoinHostPort(host, port))
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

func checkOpenAPIService(ctx context.Context, s *configv1.OpenapiUpstreamService, auth *configv1.Authentication) CheckResult {
	if s.GetSpecUrl() != "" {
		// Check if we can fetch the spec (passing auth if needed? Usually spec might be public, but API isn't)
		// For now, let's assume spec URL might need auth too if it's on the same server.
		res := checkURL(ctx, s.GetSpecUrl(), auth)
		if res.Status != StatusOk {
			return res
		}
	}

	if s.GetAddress() != "" {
		return checkURL(ctx, s.GetAddress(), auth)
	}

	return CheckResult{
		Status:  StatusOk,
		Message: "OpenAPI definition seems accessible",
	}
}

func checkSQLService(ctx context.Context, s *configv1.SqlUpstreamService) CheckResult {
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
			Message: fmt.Sprintf("Failed to initialize SQL driver: %v", util.RedactDSN(err.Error())),
			Error:   err,
		}
	}
	defer func() { _ = db.Close() }()

	// Try to ping
	err = db.PingContext(ctx)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: fmt.Sprintf("Failed to ping database: %v", util.RedactDSN(err.Error())),
			Error:   err,
		}
	}

	return CheckResult{
		Status:  StatusOk,
		Message: "Database connection successful",
	}
}

func checkMCPService(ctx context.Context, s *configv1.McpUpstreamService, auth *configv1.Authentication) CheckResult {
	switch s.WhichConnectionType() {
	case configv1.McpUpstreamService_HttpConnection_case:
		return checkURL(ctx, s.GetHttpConnection().GetHttpAddress(), auth)
	case configv1.McpUpstreamService_StdioConnection_case:
		cmd := s.GetStdioConnection().GetCommand()
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

	if len(parts) > 1 {
		arg := parts[1]
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

// applyAuthentication applies the given authentication configuration to the request.
// It resolves secrets using util.ResolveSecret.
func applyAuthentication(ctx context.Context, req *http.Request, auth *configv1.Authentication) error {
	if auth == nil {
		return nil
	}

	if apiKey := auth.GetApiKey(); apiKey != nil {
		apiKeyValue, err := util.ResolveSecret(ctx, apiKey.GetValue())
		if err != nil {
			return err
		}
		req.Header.Set(apiKey.GetParamName(), apiKeyValue)
	} else if bearerToken := auth.GetBearerToken(); bearerToken != nil {
		tokenValue, err := util.ResolveSecret(ctx, bearerToken.GetToken())
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+tokenValue)
	} else if basicAuth := auth.GetBasicAuth(); basicAuth != nil {
		passwordValue, err := util.ResolveSecret(ctx, basicAuth.GetPassword())
		if err != nil {
			return err
		}
		req.SetBasicAuth(basicAuth.GetUsername(), passwordValue)
	}
	// Note: OAuth2 logic is complex and usually requires a separate token exchange,
	// which is handled by oauth2 library. For simpler "Doctor" checks, we might verify
	// the token endpoint reachability (done separately) or we could try to get a token if we have client credentials.
	// For now, we skip full OAuth2 flow injection here unless we want to do full token negotiation.

	return nil
}
