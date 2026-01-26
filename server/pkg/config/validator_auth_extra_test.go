package config

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

// Mock helpers
func mockOSStat(exists bool, isDir bool) func(string) (os.FileInfo, error) {
	return func(name string) (os.FileInfo, error) {
		if !exists {
			return nil, os.ErrNotExist
		}
		// Use mockFileInfo from validator_test.go
		return &mockFileInfo{isDir: isDir}, nil
	}
}

func TestValidateAuthentication_OIDC(t *testing.T) {
	// Valid OIDC
	oidc := &configv1.OIDCAuth{}
	oidc.SetIssuer("https://accounts.google.com")

	auth := &configv1.Authentication{}
	auth.SetOidc(oidc)

	err := validateAuthentication(context.Background(), auth, AuthValidationContextIncoming)
	assert.NoError(t, err)

	// Invalid OIDC
	oidcInvalid := &configv1.OIDCAuth{}
	oidcInvalid.SetIssuer("not-a-url")

	authInvalid := &configv1.Authentication{}
	authInvalid.SetOidc(oidcInvalid)

	err = validateAuthentication(context.Background(), authInvalid, AuthValidationContextIncoming)
	assert.Error(t, err)
}

func TestValidateAuthentication_TrustedHeader(t *testing.T) {
	// Valid TrustedHeader
	th := &configv1.TrustedHeaderAuth{}
	th.SetHeaderName("X-User")
	th.SetHeaderValue("admin")

	auth := &configv1.Authentication{}
	auth.SetTrustedHeader(th)

	err := validateAuthentication(context.Background(), auth, AuthValidationContextIncoming)
	assert.NoError(t, err)

	// Invalid TrustedHeader (empty value)
	thInvalid := &configv1.TrustedHeaderAuth{}
	thInvalid.SetHeaderName("X-User")
	thInvalid.SetHeaderValue("")

	authInvalid := &configv1.Authentication{}
	authInvalid.SetTrustedHeader(thInvalid)

	err = validateAuthentication(context.Background(), authInvalid, AuthValidationContextIncoming)
	assert.Error(t, err)
}

func TestValidateAPIKeyAuth_Contexts(t *testing.T) {
	// Outgoing: requires Value
	apiKey := &configv1.APIKeyAuth{}
	apiKey.SetParamName("X-Key")

	sv := &configv1.SecretValue{}
	sv.SetPlainText("secret")
	apiKey.SetValue(sv)

	err := validateAPIKeyAuth(context.Background(), apiKey, AuthValidationContextOutgoing)
	assert.NoError(t, err)

	// Outgoing missing value
	apiKeyMissing := &configv1.APIKeyAuth{}
	apiKeyMissing.SetParamName("X-Key")

	err = validateAPIKeyAuth(context.Background(), apiKeyMissing, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required for outgoing auth")

	// Incoming: requires VerificationValue OR Value
	apiKeyIncoming := &configv1.APIKeyAuth{}
	apiKeyIncoming.SetParamName("X-Key")
	apiKeyIncoming.SetVerificationValue("static-secret")

	err = validateAPIKeyAuth(context.Background(), apiKeyIncoming, AuthValidationContextIncoming)
	assert.NoError(t, err)

	apiKeyIncomingMissing := &configv1.APIKeyAuth{}
	apiKeyIncomingMissing.SetParamName("X-Key")

	err = validateAPIKeyAuth(context.Background(), apiKeyIncomingMissing, AuthValidationContextIncoming)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Set either")
	assert.Contains(t, err.Error(), "verification_value")
}

func TestValidateServiceConfig_AllTypes(t *testing.T) {
	// 1. GraphQL
	gql := &configv1.GraphQLUpstreamService{}
	gql.SetAddress("http://example.com/graphql")
	svc := &configv1.UpstreamServiceConfig{}
	svc.SetGraphqlService(gql)
	assert.NoError(t, validateServiceConfig(svc))

	// 2. WebRTC
	webrtc := &configv1.WebrtcUpstreamService{}
	webrtc.SetAddress("http://example.com/signal")
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetWebrtcService(webrtc)
	assert.NoError(t, validateServiceConfig(svc))

	// 3. SQL
	sql := &configv1.SqlUpstreamService{}
	sql.SetDriver("postgres")
	sql.SetDsn("postgres://...")
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetSqlService(sql)
	assert.NoError(t, validateServiceConfig(svc))

	// 4. HTTP
	http := &configv1.HttpUpstreamService{}
	http.SetAddress("http://example.com")
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetHttpService(http)
	assert.NoError(t, validateServiceConfig(svc))

	// 5. WebSocket
	ws := &configv1.WebsocketUpstreamService{}
	ws.SetAddress("ws://example.com")
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetWebsocketService(ws)
	assert.NoError(t, validateServiceConfig(svc))

	// 6. gRPC
	grpc := &configv1.GrpcUpstreamService{}
	grpc.SetAddress("localhost:50051")
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetGrpcService(grpc)
	assert.NoError(t, validateServiceConfig(svc))

	// 7. OpenAPI
	openapi := &configv1.OpenapiUpstreamService{}
	openapi.SetAddress("http://example.com")
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetOpenapiService(openapi)
	assert.NoError(t, validateServiceConfig(svc))

	// 8. Command Line
	// We need to override LookPath for command line validation if not mocked
	oldLookPath := execLookPath
	execLookPath = func(file string) (string, error) {
		return "/bin/" + file, nil
	}
	defer func() { execLookPath = oldLookPath }()

	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand("ls")
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetCommandLineService(cmd)
	assert.NoError(t, validateServiceConfig(svc))

	// 9. MCP (HTTP)
	mcpHttp := &configv1.McpStreamableHttpConnection{}
	mcpHttp.SetHttpAddress("http://example.com")
	mcp := &configv1.McpUpstreamService{}
	mcp.SetHttpConnection(mcpHttp)
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetMcpService(mcp)
	assert.NoError(t, validateServiceConfig(svc))

	// 10. MCP (Stdio)
	mcpStdio := &configv1.McpStdioConnection{}
	mcpStdio.SetCommand("ls")
	mcp2 := &configv1.McpUpstreamService{}
	mcp2.SetStdioConnection(mcpStdio)
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetMcpService(mcp2)
	assert.NoError(t, validateServiceConfig(svc))

	// 11. Vector Service (Ignored by validation logic, so returns nil)
	vector := &configv1.VectorUpstreamService{}
	pinecone := &configv1.PineconeVectorDB{}
	pinecone.SetApiKey("key")
	vector.SetPinecone(pinecone)
	svc = &configv1.UpstreamServiceConfig{}
	svc.SetVectorService(vector)
	assert.NoError(t, validateServiceConfig(svc))
}

func TestValidateOAuth2_AutoDiscovery(t *testing.T) {
	// Test the case where TokenUrl is empty but IssuerUrl is present (Auto-discovery)
	oauth := &configv1.OAuth2Auth{}
	oauth.SetIssuerUrl("https://accounts.google.com")

	cid := &configv1.SecretValue{}
	cid.SetPlainText("id")
	oauth.SetClientId(cid)

	csec := &configv1.SecretValue{}
	csec.SetPlainText("secret")
	oauth.SetClientSecret(csec)

	err := validateOAuth2Auth(context.Background(), oauth)
	assert.NoError(t, err)

	// Test invalid TokenUrl (if present)
	oauthInvalid := &configv1.OAuth2Auth{}
	oauthInvalid.SetTokenUrl("not-url")
	oauthInvalid.SetClientId(cid)
	oauthInvalid.SetClientSecret(csec)

	err = validateOAuth2Auth(context.Background(), oauthInvalid)
	assert.Error(t, err)
}

func TestValidateCommandLineService_Local(t *testing.T) {
	// Override LookPath
	oldLookPath := execLookPath
	execLookPath = func(file string) (string, error) {
		if file == "ls" {
			return "/bin/ls", nil
		}
		return "", fmt.Errorf("executable file not found in $PATH")
	}
	defer func() { execLookPath = oldLookPath }()

	// Test command existence check (assuming 'ls' exists)
	cmd := &configv1.CommandLineUpstreamService{}
	cmd.SetCommand("ls")

	err := validateCommandLineService(cmd)
	assert.NoError(t, err)

	// Test non-existent
	cmdInvalid := &configv1.CommandLineUpstreamService{}
	cmdInvalid.SetCommand("non-existent-command-xyz")

	err = validateCommandLineService(cmdInvalid)
	assert.Error(t, err)
}

func TestValidateSchema_Invalid(t *testing.T) {
	// structpb with invalid type
	spb, err := structpb.NewStruct(map[string]interface{}{
		"type": 123,
	})
	assert.NoError(t, err)

	err = validateSchema(spb)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must be a string")
}

func TestValidateMtlsAuth_Files(t *testing.T) {
	// Mock osStat
	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()

	osStat = mockOSStat(true, false) // files exist

	mtls := &configv1.MTLSAuth{}
	mtls.SetClientCertPath("cert.pem")
	mtls.SetClientKeyPath("key.pem")
	mtls.SetCaCertPath("ca.pem")

	err := validateMtlsAuth(mtls)
	assert.NoError(t, err)
}

func TestValidateMtlsAuth_MissingFile(t *testing.T) {
	// Mock osStat failure
	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()

	osStat = mockOSStat(false, false)

	mtls := &configv1.MTLSAuth{}
	mtls.SetClientCertPath("missing.pem")
	mtls.SetClientKeyPath("missing.key")

	err := validateMtlsAuth(mtls)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestValidationErrors_ToString(t *testing.T) {
	ve := &ValidationError{ServiceName: "svc", Err: assert.AnError}
	assert.Contains(t, ve.Error(), "service \"svc\"")
}

func TestValidateUser_DuplicateID(t *testing.T) {
	u1 := &configv1.User{}
	u1.SetId("u1")
	u2 := &configv1.User{}
	u2.SetId("u1")

	cfg := &configv1.McpAnyServerConfig{}
	cfg.SetUsers([]*configv1.User{u1, u2})

	errs := Validate(context.Background(), cfg, Server)
	found := false
	for _, e := range errs {
		if e.Err.Error() == "duplicate user id" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestValidateUser_EmptyID(t *testing.T) {
	u1 := &configv1.User{}
	u1.SetId("")

	cfg := &configv1.McpAnyServerConfig{}
	cfg.SetUsers([]*configv1.User{u1})

	errs := Validate(context.Background(), cfg, Server)
	found := false
	for _, e := range errs {
		if e.Err.Error() == "user has empty id" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestValidateCollection_EmptyName(t *testing.T) {
	col := &configv1.Collection{}
	err := validateCollection(context.Background(), col)
	assert.Error(t, err)
}

func TestValidateGlobalSettings_InvalidProfile(t *testing.T) {
	p1 := &configv1.ProfileDefinition{}
	p1.SetName("p1")
	p2 := &configv1.ProfileDefinition{}
	p2.SetName("p1")

	gs := &configv1.GlobalSettings{}
	gs.SetProfileDefinitions([]*configv1.ProfileDefinition{p1, p2})

	cfg := &configv1.McpAnyServerConfig{}
	cfg.SetGlobalSettings(gs)

	errs := Validate(context.Background(), cfg, Server)
	found := false
	for _, e := range errs {
		if e.ServiceName == "global_settings" && e.Err != nil {
			found = true
		}
	}
	assert.True(t, found)
}

func TestValidateGlobalSettings_APIKey(t *testing.T) {
	// Client binary type checks api key length
	gs := &configv1.GlobalSettings{}
	gs.SetApiKey("short")

	err := validateGlobalSettings(gs, Client)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least 16 characters")
}

func TestValidateFileExists_Directory(t *testing.T) {
	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()

	osStat = mockOSStat(true, true) // isDir = true

	err := validateFileExists("somedir", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}

func TestValidateFileExists_PermError(t *testing.T) {
	oldOsStat := osStat
	defer func() { osStat = oldOsStat }()

	osStat = func(name string) (os.FileInfo, error) {
		return nil, errors.New("permission denied")
	}

	err := validateFileExists("file", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}

func TestValidateUpstreamAuthentication_AllTypes_Extra(t *testing.T) {
	ctx := context.Background()

	// API Key
	apiKey := &configv1.APIKeyAuth{}
	apiKey.SetParamName("key")
	sv := &configv1.SecretValue{}
	sv.SetPlainText("s")
	apiKey.SetValue(sv)
	auth := &configv1.Authentication{}
	auth.SetApiKey(apiKey)
	assert.NoError(t, validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing))

	// Bearer Token
	bearer := &configv1.BearerTokenAuth{}
	sv2 := &configv1.SecretValue{}
	sv2.SetPlainText("token")
	bearer.SetToken(sv2)
	auth.SetBearerToken(bearer)
	assert.NoError(t, validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing))

	// Basic Auth
	basic := &configv1.BasicAuth{}
	basic.SetUsername("user")
	sv3 := &configv1.SecretValue{}
	sv3.SetPlainText("pass")
	basic.SetPassword(sv3)
	auth.SetBasicAuth(basic)
	assert.NoError(t, validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing))

	// OAuth2
	oauth := &configv1.OAuth2Auth{}
	oauth.SetIssuerUrl("https://example.com")
	oauth.SetClientId(sv)
	oauth.SetClientSecret(sv)
	auth.SetOauth2(oauth)
	assert.NoError(t, validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing))

	// OIDC (Ignored by validateUpstreamAuthentication)
	oidc := &configv1.OIDCAuth{}
	auth.SetOidc(oidc)
	assert.NoError(t, validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing))
}
