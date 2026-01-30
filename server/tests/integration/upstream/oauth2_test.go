// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/util"
	"github.com/mcpany/core/server/tests/framework"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/protobuf/proto"
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

func (s *mockOAuth2Server) handleWellKnown(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(map[string]interface{}{
		"issuer":                                s.issuer,
		"token_endpoint":                        s.URL + oauth2TestTokenPath,
		"jwks_uri":                              s.URL + oauth2TestJWKSPath,
		"grant_types_supported":                 []string{"client_credentials"},
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

func (s *mockOAuth2Server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(s.signer.jwks())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildMCPANYAuthedServer(t *testing.T, issuer, audience string) *integration.MCPANYTestServerInfo {
	t.Helper()

	config := fmt.Sprintf(`
auth:
  oauth2:
    issuer: %s
    audience: [%s]
`, issuer, audience)
	return integration.StartMCPANYServerWithConfig(t, "mcpany_oauth2_test", config)
}

func TestUpstreamService_HTTP_WithOAuth2(t *testing.T) {
	var tokenURL, clientIDVal, clientSecretVal string
	if os.Getenv("TEST_OAUTH_SERVER_URL") != "" {
		tokenURL = os.Getenv("TEST_OAUTH_TOKEN_URL")
		clientIDVal = os.Getenv("TEST_OAUTH_CLIENT_ID")
		clientSecretVal = os.Getenv("TEST_OAUTH_CLIENT_SECRET")
	} else {
		oauth2Server := newMockOAuth2Server(t)
		defer oauth2Server.Close()
		tokenURL = oauth2Server.URL + oauth2TestTokenPath
		clientIDVal = oauth2TestClientID
		clientSecretVal = oauth2TestClientSecret
	}

	testCase := &framework.E2ETestCase{
		Name:                "Authenticated HTTP Echo Server with OAuth2",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPAuthedEchoServer,
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			const serviceID = "e2e_http_oauth2_echo"
			// tokenURL is already set above
			clientID := configv1.SecretValue_builder{
				PlainText: proto.String(clientIDVal),
			}.Build()
			clientSecret := configv1.SecretValue_builder{
				PlainText: proto.String(clientSecretVal),
			}.Build()
			oauth2AuthConfig := configv1.OAuth2Auth_builder{
				TokenUrl:     proto.String(tokenURL),
				ClientId:     clientID,
				ClientSecret: clientSecret,
			}.Build()
			authConfig := configv1.Authentication_builder{
				Oauth2: oauth2AuthConfig,
			}.Build()
			integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, authConfig)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

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

func TestUpstreamService_HTTP_WithOAuth2_InvalidCredentials(t *testing.T) {
	oauth2Server := newMockOAuth2Server(t)
	defer oauth2Server.Close()

	testCase := &framework.E2ETestCase{
		Name:                "Authenticated HTTP Echo Server with OAuth2 Invalid Credentials",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPAuthedEchoServer,
		RegisterUpstream: func(t *testing.T, registrationClient apiv1.RegistrationServiceClient, upstreamEndpoint string) {
			const serviceID = "e2e_http_oauth2_echo_invalid"
			tokenURL := oauth2Server.URL + oauth2TestTokenPath
			clientID := configv1.SecretValue_builder{
				PlainText: proto.String(oauth2TestClientID),
			}.Build()
			clientSecret := configv1.SecretValue_builder{
				PlainText: proto.String("test-client-secret"),
			}.Build()
			oauth2AuthConfig := configv1.OAuth2Auth_builder{
				TokenUrl:     proto.String(tokenURL),
				ClientId:     clientID,
				ClientSecret: clientSecret,
			}.Build()
			authConfig := configv1.Authentication_builder{
				Oauth2: oauth2AuthConfig,
			}.Build()
			integration.RegisterHTTPService(t, registrationClient, serviceID, upstreamEndpoint, "echo", "/echo", http.MethodPost, authConfig)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			const echoServiceID = "e2e_http_oauth2_echo_invalid"
			serviceID, _ := util.SanitizeServiceName(echoServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from oauth2 protected upstream"}`
			_, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(echoMessage)})
			require.Error(t, err, "Expected an error when calling echo tool with invalid auth")
		},
	}

	framework.RunE2ETest(t, testCase)
}

func TestUpstreamService_MCPANY_WithOAuth2(t *testing.T) {
	var issuer, audience, tokenURL, clientIDVal, clientSecretVal string
	if os.Getenv("TEST_OAUTH_SERVER_URL") != "" {
		issuer = os.Getenv("TEST_OAUTH_SERVER_URL") + "/" // Hydra issuer usually ends with /
		tokenURL = os.Getenv("TEST_OAUTH_TOKEN_URL")
		clientIDVal = os.Getenv("TEST_OAUTH_CLIENT_ID")
		clientSecretVal = os.Getenv("TEST_OAUTH_CLIENT_SECRET")
		audience = "test-client" // Hydra uses client ID as audience for some setups, or we can might need to adjust scope/audience
	} else {
		oauth2Server := newMockOAuth2Server(t)
		defer oauth2Server.Close()
		issuer = oauth2Server.issuer
		tokenURL = oauth2Server.URL + oauth2TestTokenPath
		clientIDVal = oauth2TestClientID
		clientSecretVal = oauth2TestClientSecret
		audience = "test-audience"
	}

	testCase := &framework.E2ETestCase{
		Name:                "MCPANY with OAuth2 Authentication",
		UpstreamServiceType: "http",
		BuildUpstream:       framework.BuildHTTPEchoServer,
		RegisterUpstream:    framework.RegisterHTTPEchoService,
		StartMCPANYServer: func(t *testing.T, _ string, _ ...string) *integration.MCPANYTestServerInfo {
			return buildMCPANYAuthedServer(t, issuer, audience)
		},
		InvokeAIClient: func(t *testing.T, mcpanyEndpoint string) {
			ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeLong)
			defer cancel()

			conf := &clientcredentials.Config{
				ClientID:     clientIDVal,
				ClientSecret: clientSecretVal,
				TokenURL:     tokenURL,
				Scopes:       []string{"openid", "offline"},
			}
			tokenSource := conf.TokenSource(ctx)

			httpClient := &http.Client{
				Transport: &oauth2.Transport{
					Source: tokenSource,
				},
			}

			testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
			cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpanyEndpoint, HTTPClient: httpClient}, nil)
			require.NoError(t, err)
			defer func() { _ = cs.Close() }()

			const echoServiceID = "e2e_http_echo"
			serviceID, _ := util.SanitizeServiceName(echoServiceID)
			sanitizedToolName, _ := util.SanitizeToolName("echo")
			toolName := serviceID + "." + sanitizedToolName
			echoMessage := `{"message": "hello world from oauth2 protected mcpany"}`
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
