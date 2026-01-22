// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidateFileExists(t *testing.T) {
	// Create a temporary file
	f, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	// Case 1: File exists
	err = validateFileExists(f.Name(), "")
	assert.NoError(t, err)

	// Case 2: File does not exist
	err = validateFileExists("/path/to/non/existent/file", "")
	assert.Error(t, err)

	// Case 3: Directory
	d, err := os.MkdirTemp("", "testdir")
	require.NoError(t, err)
	defer os.Remove(d)

	err = validateFileExists(d, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}

func TestValidateSecretValue_Nil(t *testing.T) {
	err := validateSecretValue(context.Background(), nil)
	assert.NoError(t, err)
}

func TestValidateSecretValue_EmptyEnvVar(t *testing.T) {
	secret := &configv1.SecretValue{
		Value: &configv1.SecretValue_EnvironmentVariable{
			EnvironmentVariable: "NON_EXISTENT_VAR_XYZ",
		},
	}
	err := validateSecretValue(context.Background(), secret)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable \"NON_EXISTENT_VAR_XYZ\" error")
}

func TestValidateAuditConfig(t *testing.T) {
	// Case 1: Nil config
	assert.NoError(t, validateAuditConfig(nil))

	// Case 2: Disabled
	assert.NoError(t, validateAuditConfig(&configv1.AuditConfig{Enabled: proto.Bool(false)}))

	// Case 3: File storage without path
	stFile := configv1.AuditConfig_STORAGE_TYPE_FILE
	err := validateAuditConfig(&configv1.AuditConfig{
		Enabled:     proto.Bool(true),
		StorageType: &stFile,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output_path is required")

	// Case 4: Webhook storage without URL
	stWebhook := configv1.AuditConfig_STORAGE_TYPE_WEBHOOK
	err = validateAuditConfig(&configv1.AuditConfig{
		Enabled:     proto.Bool(true),
		StorageType: &stWebhook,
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook_url is required")

	// Case 5: Invalid webhook URL
	err = validateAuditConfig(&configv1.AuditConfig{
		Enabled:     proto.Bool(true),
		StorageType: &stWebhook,
		WebhookUrl:  proto.String("not-a-url"),
	})
	assert.Error(t, err)
}

func TestValidateDLPConfig(t *testing.T) {
	// Case 1: Nil
	assert.NoError(t, validateDLPConfig(nil))

	// Case 2: Valid patterns
	err := validateDLPConfig(&configv1.DLPConfig{
		CustomPatterns: []string{"^abc$", "[0-9]+"},
	})
	assert.NoError(t, err)

	// Case 3: Invalid pattern
	err = validateDLPConfig(&configv1.DLPConfig{
		CustomPatterns: []string{"["},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestValidateSecretValue_RemoteContent_Errors(t *testing.T) {
	// Empty URL
	sv := &configv1.SecretValue{
		Value: &configv1.SecretValue_RemoteContent{
			RemoteContent: &configv1.RemoteContent{
				HttpUrl: proto.String(""),
			},
		},
	}
	err := validateSecretValue(context.Background(), sv)
	assert.Error(t, err)
	// Error message comes from http.NewRequestWithContext or ResolveSecret
}

func TestValidateContainerEnvironment_Errors(t *testing.T) {
	// Empty host path
	env := &configv1.ContainerEnvironment{
		Image: proto.String("alpine"),
		Volumes: map[string]string{
			"": "/container/path",
		},
	}
	err := validateContainerEnvironment(env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host path is empty")

	// Empty container path
	env.Volumes = map[string]string{
		"/host/path": "",
	}
	err = validateContainerEnvironment(env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container path is empty")
}

func TestValidateUpstreamAuthentication_Errors(t *testing.T) {
	ctx := context.Background()

	// API Key errors
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String(""), // Error
			},
		},
	}
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)

	// Bearer Token errors
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BearerToken{
			BearerToken: &configv1.BearerTokenAuth{
				Token: nil, // Error: token empty
			},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)

	// Basic Auth errors
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BasicAuth{
			BasicAuth: &configv1.BasicAuth{
				Username: proto.String(""), // Error
			},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
}

func TestValidateSQLService_Errors(t *testing.T) {
	// Empty Driver
	s := &configv1.SqlUpstreamService{Driver: proto.String("")}
	err := validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver")

	// Empty DSN
	s = &configv1.SqlUpstreamService{Driver: proto.String("postgres"), Dsn: proto.String("")}
	err = validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dsn")

	// Call with empty query
	s = &configv1.SqlUpstreamService{
		Driver: proto.String("postgres"),
		Dsn:    proto.String("postgres://"),
		Calls: map[string]*configv1.SqlCallDefinition{
			"call1": {Query: proto.String("")},
		},
	}
	err = validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query is empty")
}

func TestValidateGraphQLService_Errors(t *testing.T) {
	// Empty Address
	s := &configv1.GraphQLUpstreamService{Address: proto.String("")}
	err := validateGraphQLService(s)
	assert.Error(t, err)

	// Invalid URL
	s = &configv1.GraphQLUpstreamService{Address: proto.String("not-url")}
	err = validateGraphQLService(s)
	assert.Error(t, err)
}

func TestValidateWebrtcService_Errors(t *testing.T) {
	// Empty Address
	s := &configv1.WebrtcUpstreamService{Address: proto.String("")}
	err := validateWebrtcService(s)
	assert.Error(t, err)

	// Invalid URL
	s = &configv1.WebrtcUpstreamService{Address: proto.String("not-url")}
	err = validateWebrtcService(s)
	assert.Error(t, err)
}

func TestValidateOAuth2Auth_Errors(t *testing.T) {
	ctx := context.Background()
	// Empty Token URL
	auth := &configv1.OAuth2Auth{TokenUrl: proto.String("")}
	err := validateOAuth2Auth(ctx, auth)
	assert.Error(t, err)

	// Invalid URL
	auth = &configv1.OAuth2Auth{TokenUrl: proto.String("not-url")}
	err = validateOAuth2Auth(ctx, auth)
	assert.Error(t, err)
}

func TestValidateUpstreamAuthentication_AllTypes(t *testing.T) {
	ctx := context.Background()
	// Oauth2
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{TokenUrl: proto.String("")},
		},
	}
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)

	// Mtls
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Mtls{
			Mtls: &configv1.MTLSAuth{ClientCertPath: proto.String("")},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
}

func TestValidateGCSettings(t *testing.T) {
	// Case 1: Invalid interval
	gc := &configv1.GCSettings{Interval: proto.String("invalid")}
	err := validateGCSettings(gc)
	assert.Error(t, err)

	// Case 2: Invalid TTL
	gc = &configv1.GCSettings{Ttl: proto.String("invalid")}
	err = validateGCSettings(gc)
	assert.Error(t, err)

	// Case 3: Empty path in Paths
	gc = &configv1.GCSettings{
		Enabled: proto.Bool(true),
		Paths:   []string{""},
	}
	err = validateGCSettings(gc)
	assert.Error(t, err)

	// Case 4: Relative path
	gc = &configv1.GCSettings{
		Enabled: proto.Bool(true),
		Paths:   []string{"relative/path"},
	}
	err = validateGCSettings(gc)
	assert.Error(t, err)
}

func TestValidateHTTPService_SchemaErrors(t *testing.T) {
	// Invalid Input Schema
	s := &configv1.HttpUpstreamService{
		Address: proto.String("http://example.com"),
		Calls: map[string]*configv1.HttpCallDefinition{
			"call1": {
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
					},
				},
			},
		},
	}
	err := validateHTTPService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input_schema")
}

func TestValidateAPIKeyAuth_Errors(t *testing.T) {
	ctx := context.Background()
	// Value missing
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String("api_key"),
			},
		},
	}
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required for outgoing")

	// Value resolves to empty
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String("api_key"),
				Value: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: ""},
				},
			},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolved api key value is empty")
}

func TestValidateAPIKeyAuth_Incoming_Errors(t *testing.T) {
	ctx := context.Background()
	// Both Value and VerificationValue missing
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String("api_key"),
			},
		},
	}
	// Use validateAuthentication to access internal logic via authCtx
	err := validateAuthentication(ctx, auth, AuthValidationContextIncoming)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api key configuration is empty")

	// VerificationValue present (Valid for Incoming)
	auth.GetApiKey().VerificationValue = proto.String("static-key")
	err = validateAuthentication(ctx, auth, AuthValidationContextIncoming)
	assert.NoError(t, err)
}

func TestValidateMcpService_BundleErrors(t *testing.T) {
	// Empty Bundle Path
	s := &configv1.McpUpstreamService{
		ConnectionType: &configv1.McpUpstreamService_BundleConnection{
			BundleConnection: &configv1.McpBundleConnection{BundlePath: proto.String("")},
		},
	}
	err := validateMcpService(context.Background(), s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty bundle_path")
}

func TestValidateSQLService_SchemaErrors(t *testing.T) {
	// Invalid Input Schema
	s := &configv1.SqlUpstreamService{
		Driver: proto.String("postgres"),
		Dsn:    proto.String("postgres://"),
		Calls: map[string]*configv1.SqlCallDefinition{
			"call1": {
				Query: proto.String("SELECT 1"),
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
					},
				},
			},
		},
	}
	err := validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input_schema")
}

func TestValidate_ClientErrors(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			ApiKey: proto.String("short"),
		},
	}
	errs := Validate(context.Background(), cfg, Client)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Err.Error(), "at least 16 characters")
}

func TestValidateCollection_Coverage(t *testing.T) {
	ctx := context.Background()

	// 1. Invalid Name
	coll := &configv1.Collection{
		Name:    proto.String(""),
		HttpUrl: proto.String("http://example.com/collection.json"),
	}
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for empty name")
	}

	// 2. Invalid URL
	coll.Name = proto.String("valid-name")
	coll.HttpUrl = proto.String("not-a-url")
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for invalid URL")
	}

	// 3. Invalid Scheme
	coll.HttpUrl = proto.String("ftp://example.com")
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for invalid scheme")
	}

	// 4. Valid with Auth (ApiKey)
	coll.HttpUrl = proto.String("http://example.com/collection.json")
	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String("x-api-key"),
				Value: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
			},
		},
	}

	if err := validateCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid API Key
	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String(""), // Empty ParamName
				Value: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
			},
		},
	}
	assert.Error(t, validateCollection(ctx, coll))

	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String("header"),
				// Missing Value
			},
		},
	}
	assert.Error(t, validateCollection(ctx, coll))

	// 5. Valid with Bearer
	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BearerToken{
			BearerToken: &configv1.BearerTokenAuth{
				Token: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "token"},
				},
			},
		},
	}

	if err := validateCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 7. Valid with mTLS (failing due to missing files, but covering code path)
	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Mtls{
			Mtls: &configv1.MTLSAuth{
				ClientCertPath: proto.String("/tmp/nonexistent_cert.pem"),
				ClientKeyPath:  proto.String("/tmp/nonexistent_key.pem"),
			},
		},
	}
	// Should fail file check
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for missing mTLS files")
	}

	// 8. Valid with OAuth2
	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				TokenUrl:     proto.String("https://example.com/token"),
				ClientId:     &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}},
				ClientSecret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
			},
		},
	}
	if err := validateCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid OAuth2
	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				TokenUrl: proto.String(""), // Empty URL
			},
		},
	}
	assert.Error(t, validateCollection(ctx, coll))

	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				TokenUrl: proto.String("not-url"), // Invalid URL
			},
		},
	}
	assert.Error(t, validateCollection(ctx, coll))
}

func TestValidate_ExtraServices(t *testing.T) {
	tests := []struct {
		name                string
		config              *configv1.McpAnyServerConfig
		expectedErrorCount  int
		expectedErrorString string
	}{
		{
			name: "valid graphql service",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				// Use struct construction
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("graphql-valid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{
							GraphqlService: &configv1.GraphQLUpstreamService{
								Address: proto.String("http://example.com/graphql"),
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid graphql service - bad scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "graphql-bad-scheme",
							"graphql_service": {
								"address": "ftp://example.com/graphql"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "service \"graphql-bad-scheme\": invalid graphql address scheme: ftp\n\t-> Fix: Use 'http' or 'https' as the scheme.",
		},
		{
			name: "valid webrtc service",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				// Use struct construction
				cfg.UpstreamServices = []*configv1.UpstreamServiceConfig{
					{
						Name: proto.String("webrtc-valid"),
						ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
							WebrtcService: &configv1.WebrtcUpstreamService{
								Address: proto.String("http://example.com/webrtc"),
							},
						},
					},
				}
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid webrtc service - bad scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "webrtc-bad-scheme",
							"webrtc_service": {
								"address": "ftp://example.com/webrtc"
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: "service \"webrtc-bad-scheme\": invalid webrtc address scheme: ftp\n\t-> Fix: Use 'http' or 'https' as the scheme.",
		},
		{
			name: "valid upstream service collection",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-1",
							"http_url": "http://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "invalid upstream service collection - empty name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "",
							"http_url": "http://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "": collection name is empty`,
		},
		{
			name: "invalid upstream service collection - empty url",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-no-url",
							"http_url": ""
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "collection-no-url": collection must have either http_url or inline content (services/skills)`,
		},
		{
			name: "invalid upstream service collection - bad url scheme",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-bad-scheme",
							"http_url": "ftp://example.com/collection"
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "collection-bad-scheme": invalid collection http_url scheme: ftp`,
		},
		{
			name: "valid upstream service collection with auth",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"collections": [
						{
							"name": "collection-auth",
							"http_url": "http://example.com/collection",
							"authentication": {
								"basic_auth": {
									"username": "user",
									"password": { "plainText": "pass" }
								}
							}
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount: 0,
		},
		{
			name: "duplicate service name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "service-1",
							"http_service": { "address": "http://example.com/1" }
						},
						{
							"name": "service-1",
							"http_service": { "address": "http://example.com/2" }
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "service-1": duplicate service name found`,
		},
		{
			name: "invalid upstream service - empty name",
			config: func() *configv1.McpAnyServerConfig {
				cfg := &configv1.McpAnyServerConfig{}
				require.NoError(t, protojson.Unmarshal([]byte(`{
					"upstream_services": [
						{
							"name": "",
							"http_service": { "address": "http://example.com/empty-name" }
						}
					]
				}`), cfg))
				return cfg
			}(),
			expectedErrorCount:  1,
			expectedErrorString: `service "": service name is empty`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validationErrors := Validate(context.Background(), tt.config, Server)
			if tt.expectedErrorCount > 0 {
				require.NotEmpty(t, validationErrors, "expected validation errors but got none")
				found := false
				for _, err := range validationErrors {
					if err.Error() == tt.expectedErrorString {
						found = true
						break
					}
				}
				if !found {
					assert.EqualError(t, &validationErrors[0], tt.expectedErrorString)
				}
			} else {
				assert.Empty(t, validationErrors)
			}
		})
	}
}

func boolPtr(b bool) *bool                                                                { return &b }
func storageTypePtr(t configv1.AuditConfig_StorageType) *configv1.AuditConfig_StorageType { return &t }

func TestValidateUsers(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		users        []*configv1.User
		expectErr    bool
		errSubstring string
	}{
		{
			name: "Valid User",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
					Authentication: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_ApiKey{
							ApiKey: &configv1.APIKeyAuth{
								ParamName:         strPtr("key"),
								VerificationValue: strPtr("secret"),
							},
						},
					},
				},
			},
			expectErr: false,
		},
		{
			name: "Missing ID",
			users: []*configv1.User{
				{
					Id: strPtr(""), // Empty string pointer
				},
			},
			expectErr:    true,
			errSubstring: "user has empty id",
		},
		{
			name: "Duplicate ID",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
					Authentication: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_ApiKey{
							ApiKey: &configv1.APIKeyAuth{ParamName: strPtr("k"), VerificationValue: strPtr("v")},
						},
					},
				},
				{
					Id: strPtr("user1"), // Duplicate
					Authentication: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_ApiKey{
							ApiKey: &configv1.APIKeyAuth{ParamName: strPtr("k"), VerificationValue: strPtr("v")},
						},
					},
				},
			},
			expectErr:    true,
			errSubstring: "duplicate user id",
		},
		{
			name: "Missing Authentication",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
				},
			},
			expectErr: false,
		},
		{
			name: "Invalid OAuth2",
			users: []*configv1.User{
				{
					Id: strPtr("user1"),
					Authentication: &configv1.Authentication{
						AuthMethod: &configv1.Authentication_Oauth2{
							Oauth2: &configv1.OAuth2Auth{
								TokenUrl: strPtr("invalid-url"),
							},
						},
					},
				},
			},
			expectErr:    true,
			errSubstring: "invalid oauth2 token_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configv1.McpAnyServerConfig{
				Users: tt.users,
			}
			errs := Validate(ctx, config, Server)
			if tt.expectErr {
				assert.NotEmpty(t, errs)
				found := false
				for _, e := range errs {
					if assert.Contains(t, e.Err.Error(), tt.errSubstring) {
						found = true
						break
					}
				}
				if !found && len(errs) > 0 {
					// Check if substring match failed but error existed
					// Actually strict check:
					assert.Fail(t, "expected error substring not found", "substring: %s, errors: %v", tt.errSubstring, errs)
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}

func TestValidateGlobalSettings_Extended(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		gs           *configv1.GlobalSettings
		expectErr    bool
		errSubstring string
	}{
		{
			name: "Valid Audit File",
			gs: &configv1.GlobalSettings{
				Audit: &configv1.AuditConfig{
					Enabled:     boolPtr(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_FILE),
					OutputPath:  strPtr("/var/log/audit.log"),
				},
			},
			expectErr: false,
		},
		{
			name: "Audit File Missing Path",
			gs: &configv1.GlobalSettings{
				Audit: &configv1.AuditConfig{
					Enabled:     boolPtr(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_FILE),
				},
			},
			expectErr:    true,
			errSubstring: "output_path is required",
		},
		{
			name: "Audit Webhook Invalid URL",
			gs: &configv1.GlobalSettings{
				Audit: &configv1.AuditConfig{
					Enabled:     boolPtr(true),
					StorageType: storageTypePtr(configv1.AuditConfig_STORAGE_TYPE_WEBHOOK),
					WebhookUrl:  strPtr("not-a-url"),
				},
			},
			expectErr:    true,
			errSubstring: "invalid webhook_url",
		},
		{
			name: "DLP Invalid Regex",
			gs: &configv1.GlobalSettings{
				Dlp: &configv1.DLPConfig{
					Enabled:        boolPtr(true),
					CustomPatterns: []string{"["}, // Invalid regex
				},
			},
			expectErr:    true,
			errSubstring: "invalid regex pattern",
		},
		{
			name: "GC Invalid Interval",
			gs: &configv1.GlobalSettings{
				GcSettings: &configv1.GCSettings{
					Enabled:  boolPtr(true),
					Interval: strPtr("not-a-duration"),
				},
			},
			expectErr:    true,
			errSubstring: "invalid interval",
		},
		{
			name: "GC Insecure Path",
			gs: &configv1.GlobalSettings{
				GcSettings: &configv1.GCSettings{
					Enabled: boolPtr(true),
					Paths:   []string{"../etc"},
				},
			},
			expectErr:    true,
			errSubstring: "not secure",
		},
		{
			name: "GC Relative Path (Not Allowed)",
			gs: &configv1.GlobalSettings{
				GcSettings: &configv1.GCSettings{
					Enabled: boolPtr(true),
					Paths:   []string{"relative/path"},
				},
			},
			expectErr:    true,
			errSubstring: "must be absolute",
		},
		{
			name: "Duplicate Profile Name",
			gs: &configv1.GlobalSettings{
				ProfileDefinitions: []*configv1.ProfileDefinition{
					{Name: strPtr("p1")},
					{Name: strPtr("p1")},
				},
			},
			expectErr:    true,
			errSubstring: "duplicate profile definition name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &configv1.McpAnyServerConfig{
				GlobalSettings: tt.gs,
			}
			errs := Validate(ctx, config, Server)
			if tt.expectErr {
				assert.NotEmpty(t, errs)
				found := false
				for _, e := range errs {
					if len(e.Err.Error()) > 0 && (tt.errSubstring == "" || assert.Contains(t, e.Err.Error(), tt.errSubstring)) {
						found = true
						break
					}
				}
				if !found {
					t.Logf("Errors found: %v", errs)
					assert.Fail(t, "expected error not found")
				}
			} else {
				assert.Empty(t, errs)
			}
		})
	}
}
