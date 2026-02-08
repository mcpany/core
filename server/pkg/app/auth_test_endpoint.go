// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/logging"
	"github.com/mcpany/core/server/pkg/validation"
	"google.golang.org/protobuf/encoding/protojson"
)

// AuthTestRequest defines the structure for an authentication test request.
//
// Summary: defines the structure for an authentication test request.
type AuthTestRequest struct {
	CredentialID  string         `json:"credential_id"`
	ServiceType   string         `json:"service_type"`
	ServiceConfig map[string]any `json:"service_config"`
}

// AuthTestResponse defines the structure for an authentication test response.
//
// Summary: defines the structure for an authentication test response.
type AuthTestResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (a *Application) handleAuthTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AuthTestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
		defer cancel()

		logger := logging.GetLogger()
		logger.Info("Received auth test request", "credential_id", req.CredentialID, "type", req.ServiceType)

		// 1. Resolve Credential
		var cred *configv1.Credential
		if req.CredentialID != "" && req.CredentialID != "none" && a.Storage != nil {
			var err error
			cred, err = a.Storage.GetCredential(ctx, req.CredentialID)
			if err != nil {
				writeAuthResponse(w, false, fmt.Sprintf("Failed to load credential: %v", err))
				return
			}
			if cred == nil {
				writeAuthResponse(w, false, "Credential not found")
				return
			}
		}

		// 2. Parse Config
		var svcConfig configv1.UpstreamServiceConfig
		if req.ServiceConfig != nil {
			bytes, err := json.Marshal(req.ServiceConfig)
			if err == nil {
				_ = protojson.Unmarshal(bytes, &svcConfig)
			}
		}

		// 3. Test based on type
		var err error
		switch strings.ToUpper(req.ServiceType) { // Normalize service type to uppercase for consistent matching
		case "HTTP":
			err = testHTTPConnection(ctx, &svcConfig, cred)
		case "COMMAND_LINE", "CMD":
			err = testCommandConnection(ctx, &svcConfig, cred)
		default:
			// For generic/unknown types, we just verify the credential exists (if requested)
			if cred != nil {
				err = nil // Credential loaded fine, considered "connected" for unsupported types for now
			}
			// if credential was requested but not found, we already returned early
		}

		if err != nil {
			writeAuthResponse(w, false, err.Error())
		} else {
			writeAuthResponse(w, true, "Connection verification successful")
		}
	}
}

func writeAuthResponse(w http.ResponseWriter, success bool, message string) {
	resp := AuthTestResponse{
		Success: success,
		Message: message,
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func testHTTPConnection(ctx context.Context, cfg *configv1.UpstreamServiceConfig, cred *configv1.Credential) error {
	httpSvc := cfg.GetHttpService()
	var url string
	if httpSvc != nil && httpSvc.GetAddress() != "" {
		url = httpSvc.GetAddress()
	} else if mcpSvc := cfg.GetMcpService(); mcpSvc != nil {
		if httpConn := mcpSvc.GetHttpConnection(); httpConn != nil {
			url = httpConn.GetHttpAddress()
		}
	}

	if url == "" {
		return fmt.Errorf("missing http_service configuration or address")
	}

	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	if err := validation.IsSafeURL(url); err != nil {
		return fmt.Errorf("unsafe url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Apply Credential
	if cred != nil && cred.GetAuthentication() != nil {
		auth := cred.GetAuthentication()
		if apiKey := auth.GetApiKey(); apiKey != nil {
			// Handle API Key
			val := ""
			if sv := apiKey.GetValue(); sv != nil {
				if sv.WhichValue() == configv1.SecretValue_PlainText_case {
					val = sv.GetPlainText()
				} else if sv.WhichValue() == configv1.SecretValue_EnvironmentVariable_case {
					val = "env:" + sv.GetEnvironmentVariable()
				}
			}

			if val != "" {
				name := apiKey.GetParamName()
				if name == "" {
					name = "Authorization" // Default? Or should be required?
				}
				if apiKey.GetIn() == configv1.APIKeyAuth_HEADER {
					req.Header.Set(name, val)
				} else if apiKey.GetIn() == configv1.APIKeyAuth_QUERY {
					q := req.URL.Query()
					q.Add(name, val)
					req.URL.RawQuery = q.Encode()
				}
			}
		} else if token := auth.GetBearerToken(); token != nil {
			// Handle Bearer Token
			val := ""
			if t := token.GetToken(); t != nil {
				if t.WhichValue() == configv1.SecretValue_PlainText_case {
					val = t.GetPlainText()
				}
			}
			if val != "" {
				req.Header.Set("Authorization", "Bearer "+val)
			}
		}
		// Add other auth types as needed (Basic, etc.)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logging.GetLogger().Warn("failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned error status: %d", resp.StatusCode)
	}

	return nil
}

func testCommandConnection(_ context.Context, cfg *configv1.UpstreamServiceConfig, _ *configv1.Credential) error {
	cmdSvc := cfg.GetCommandLineService()
	if cmdSvc == nil || cmdSvc.GetCommand() == "" {
		return fmt.Errorf("missing command_line_service configuration")
	}

	// Safety check: Don't run arbitrary commands from an unverified public endpoint if exposed?
	// This is a debug endpoint, usually guarded by auth or strict access.
	// For "test", maybe we just check if binary exists?
	// Executing the command might define it start the service, which might block.
	// We should just check if the executable is in path or run with --version / --help.

	cmdStr := cmdSvc.GetCommand()
	// Split command? config usually has full command string or args.
	// CommandLineService has `command` (string) and potential args in `mcp_stdio_connection` but here it's simple.

	args := strings.Fields(cmdStr)
	if len(args) == 0 {
		return fmt.Errorf("empty command")
	}

	executable := args[0]
	path, err := exec.LookPath(executable)
	if err != nil {
		return fmt.Errorf("executable '%s' not found in PATH: %w", executable, err)
	}

	// We found the executable. We could try running it, but that might have side effects.
	// Just confirming it exists is a good "connection" check for local commands.
	_ = path
	return nil
}
