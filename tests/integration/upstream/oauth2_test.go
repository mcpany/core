/*
 * Copyright 2025 Author(s) of MCP-XY
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpxy/core/tests/framework"
	"github.com/mcpxy/core/tests/integration"
	apiv1 "github.com/mcpxy/core/proto/api/v1"
	configv1 "github.com/mcpxy/core/proto/config/v1"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/mcpxy/core/pkg/util"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
)

const (
	oauth2TestClientID     = "test-client"
	oauth2TestClientSecret = "test-secret"
	oauth2TestWellKnown    = "/.well-known/openid-configuration"
	oauth2TestTokenPath    = "/token"
	oauth2TestJWKSPath     = "/jwks"
)

type mockOAuth2Server struct {
	*httptest.Server
	issuer string
	signer *jwksSigner
}

func newMockOAuth2Server(t *testing.T) *mockOAuth2Server {
	t.Helper()

	signer, err := newJwksSigner()
	require.NoError(t, err)

	server := &mockOAuth2Server{
		signer: signer,
	}

	mux := http.NewServeMux()
	mux.HandleFunc(oauth2TestWellKnown, server.handleWellKnown)
	mux.HandleFunc(oauth2TestTokenPath, server.handleToken)
	mux.HandleFunc(oauth2TestJWKSPath, server.handleJWKS)
	server.Server = httptest.NewServer(mux)
	server.issuer = server.URL

	return server
}

func (s *mockOAuth2Server) handleWellKnown(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"issuer":                 s.issuer,
		"token_endpoint":         s.URL + oauth2TestTokenPath,
		"jwks_uri":               s.URL + oauth2TestJWKSPath,
		"grant_types_supported":  []string{"client_credentials"},
		"token_endpoint_auth_methods_supported": []string{"client_secret_basic"},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *mockOAuth2Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	clientID, clientSecret, ok := r.BasicAuth()
	if !ok {
		clientID = r.Form.Get("client_id")
		clientSecret = r.Form.Get("client_secret")
	}

	if clientID != oauth2TestClientID || clientSecret != oauth2TestClientSecret {
		http.Error(w, "invalid client credentials", http.StatusUnauthorized)
		return
	}
	if r.Form.Get("grant_type") != "client_credentials" {
		http.Error(w, "unsupported grant type", http.StatusBadRequest)
		return
	}

	token, err := s.signer.newJWT(s.issuer, []string{"test-audience"})
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to create token: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   3600,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *mockOAuth2Server) handleJWKS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(s.signer.jwks())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildMCPXYAuthedServer(t *testing.T, issuer, audience string) *integration.MCPXYTestServerInfo {
	t.Helper()

	config := fmt.Sprintf(`
auth:
  oauth2:
    issuer: %s
    audience: [%s]
`, issuer, audience)
	return integration.StartMCPXYServerWithConfig(t, "mcpxy_oauth2_test", config)
}

func TestUpstreamService_HTTP_WithOAuth2(t *testing.T) {
	oauth2Server := newMockOAuth2Server(t)
	defer oauth2Server.Close()

	testCase := &framework.E2ETestCase{
		Name:                "Authenticated HTTP Echo Server with OAuth2",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPAuthedEchoServer,
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			const serviceID = "e2e_http_oauth2_echo"
			tokenURL := oauth2Server.URL + oauth2TestTokenPath
			clientID := oauth2TestClientID
			clientSecret := oauth2TestClientSecret
			oauth2AuthConfig := configv1.UpstreamOAuth2Auth_builder{
				TokenUrl:     &tokenURL,
				ClientId:     &clientID,
				ClientSecret: &clientSecret,
			}.Build()
			authConfig := configv1.UpstreamAuthentication_builder{
				Oauth2: oauth2AuthConfig,
			}.Build()
			integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, authConfig)
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint}, nil)
			require.NoError(t, err)
			defer cs.Close()

			const echoServiceID = "e2e_http_oauth2_echo"
			serviceID, _ := util.SanitizeServiceName(echoServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from oauth2 protected upstream"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.NoError(t, err, "Error calling echo tool with correct auth")
			require.NotNil(t, res, "Nil response from echo tool with correct auth")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, echoMessage, content.Text, "The echoed message does not match the original")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestUpstreamService_MCPXY_WithOAuth2(t *testing.T) {
	oauth2Server := newMockOAuth2Server(t)
	defer oauth2Server.Close()

	testCase := &framework.E2ETestCase{
		Name:                "MCPXY with OAuth2 Authentication",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPEchoServer,
		RegisterUpstream:    framework.RegisterHTTPEchoService,
		StartMCPXYServer: func(t *testing.T, testName string, extraArgs ...string) *integration.MCPXYTestServerInfo {
			return buildMCPXYAuthedServer(t, oauth2Server.issuer, "test-audience")
		},
		InvokeAIClient: func(t *testing.T, mcpxyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			conf := &clientcredentials.Config{
				ClientID:     oauth2TestClientID,
				ClientSecret: oauth2TestClientSecret,
				TokenURL:     oauth2Server.URL + oauth2TestTokenPath,
			}
			tokenSource := conf.TokenSource(ctx)

			httpClient := &http.Client{
				Transport: &oauth2.Transport{
					Source: tokenSource,
				},
			}

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpxyEndpoint, HTTPClient: httpClient}, nil)
			require.NoError(t, err)
			defer cs.Close()

			const echoServiceID = "e2e_http_echo"
			serviceID, _ := util.SanitizeServiceName(echoServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from oauth2 protected mcpxy"}`
			res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.NoError(t, err, "Error calling echo tool with correct auth")
			require.NotNil(t, res, "Nil response from echo tool with correct auth")
			switch content := res.Content[0].(type) {
			case *mcp.TextContent:
				require.JSONEq(t, echoMessage, content.Text, "The echoed message does not match the original")
			default:
				t.Fatalf("Unexpected content type: %T", content)
			}
		},
	}

	framework.RunE2ETest(t, testCase)
}
