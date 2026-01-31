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
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidateFileExists(t *testing.T) {
	// Create a temporary file
	f, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	defer os.Remove(f.Name())

	// Case 1: File exists
	err = validateFileExists(context.Background(), f.Name(), "")
	assert.NoError(t, err)

	// Case 2: File does not exist
	err = validateFileExists(context.Background(), "/path/to/non/existent/file", "")
	assert.Error(t, err)

	// Case 3: Directory
	d, err := os.MkdirTemp("", "testdir")
	require.NoError(t, err)
	defer os.Remove(d)

	err = validateFileExists(context.Background(), d, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}

func TestValidateAuditConfig(t *testing.T) {
	// Case 1: Nil config
	assert.NoError(t, validateAuditConfig(nil))

	// Case 2: Disabled
	assert.NoError(t, validateAuditConfig(configv1.AuditConfig_builder{
		Enabled: proto.Bool(false),
	}.Build()))

	err := validateAuditConfig(configv1.AuditConfig_builder{
		Enabled:     proto.Bool(true),
		StorageType: configv1.AuditConfig_STORAGE_TYPE_FILE.Enum(),
	}.Build())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output_path is required")

	err = validateAuditConfig(configv1.AuditConfig_builder{
		Enabled:     proto.Bool(true),
		StorageType: configv1.AuditConfig_STORAGE_TYPE_WEBHOOK.Enum(),
	}.Build())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook_url is required")

	// Case 5: Invalid webhook URL
	err = validateAuditConfig(configv1.AuditConfig_builder{
		Enabled:     proto.Bool(true),
		StorageType: configv1.AuditConfig_STORAGE_TYPE_WEBHOOK.Enum(),
		WebhookUrl:  proto.String("not-a-url"),
	}.Build())
	assert.Error(t, err)
}

func TestValidateDLPConfig(t *testing.T) {
	// Case 1: Nil
	assert.NoError(t, validateDLPConfig(nil))

	// Case 2: Valid patterns
	err := validateDLPConfig(configv1.DLPConfig_builder{
		CustomPatterns: []string{"^abc$", "[0-9]+"},
	}.Build())
	assert.NoError(t, err)

	// Case 3: Invalid pattern
	err = validateDLPConfig(configv1.DLPConfig_builder{
		CustomPatterns: []string{"["},
	}.Build())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid regex pattern")
}

func TestValidateSecretValue_RemoteContent_Errors(t *testing.T) {
	// Empty URL
	sv := configv1.SecretValue_builder{
		RemoteContent: configv1.RemoteContent_builder{
			HttpUrl: proto.String(""),
		}.Build(),
	}.Build()
	err := validateSecretValue(context.Background(), sv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty http_url")

	// Invalid URL scheme
	sv = configv1.SecretValue_builder{
		RemoteContent: configv1.RemoteContent_builder{
			HttpUrl: proto.String("ftp://example.com/secret"),
		}.Build(),
	}.Build()
	err = validateSecretValue(context.Background(), sv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http_url scheme")
}

func TestValidateContainerEnvironment_Errors(t *testing.T) {
	// Empty host path
	env := configv1.ContainerEnvironment_builder{
		Image: proto.String("alpine"),
		Volumes: map[string]string{
			"": "/container/path",
		},
	}.Build()
	err := validateContainerEnvironment(context.Background(), env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "host path is empty")

	// Empty container path
	env.SetVolumes(map[string]string{
		"/host/path": "",
	})
	err = validateContainerEnvironment(context.Background(), env)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "container path is empty")
}

func TestValidateUpstreamAuthentication_Errors(t *testing.T) {
	ctx := context.Background()

	// API Key errors
	auth := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String(""),
		}.Build(),
	}.Build()
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)

	// Bearer Token errors
	auth = configv1.Authentication_builder{
		BearerToken: configv1.BearerTokenAuth_builder{
			Token: nil,
		}.Build(),
	}.Build()
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)

	// Basic Auth errors
	auth = configv1.Authentication_builder{
		BasicAuth: configv1.BasicAuth_builder{
			Username: proto.String(""),
		}.Build(),
	}.Build()
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
}

func TestValidateSQLService_Errors(t *testing.T) {
	// Empty Driver
	s := configv1.SqlUpstreamService_builder{}.Build()
	s.SetDriver("")
	err := validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver")

	// Empty DSN
	s = configv1.SqlUpstreamService_builder{}.Build()
	s.SetDriver("postgres")
	s.SetDsn("")
	err = validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "dsn")

	// Call with empty query
	s = configv1.SqlUpstreamService_builder{
		Driver: proto.String("postgres"),
		Dsn:    proto.String("postgres://"),
		Calls: map[string]*configv1.SqlCallDefinition{
			"call1": configv1.SqlCallDefinition_builder{
				Query: proto.String(""),
			}.Build(),
		},
	}.Build()
	err = validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "query is empty")
}

func TestValidateGraphQLService_Errors(t *testing.T) {
	// Empty Address
	s := configv1.GraphQLUpstreamService_builder{
		Address: proto.String(""),
	}.Build()
	err := validateGraphQLService(s)
	assert.Error(t, err)

	// Invalid URL
	s = configv1.GraphQLUpstreamService_builder{
		Address: proto.String("not-url"),
	}.Build()
	err = validateGraphQLService(s)
	assert.Error(t, err)
}

func TestValidateWebrtcService_Errors(t *testing.T) {
	// Empty Address
	s := configv1.WebrtcUpstreamService_builder{
		Address: proto.String(""),
	}.Build()
	err := validateWebrtcService(s)
	assert.Error(t, err)

	// Invalid URL
	s = configv1.WebrtcUpstreamService_builder{
		Address: proto.String("not-url"),
	}.Build()
	err = validateWebrtcService(s)
	assert.Error(t, err)
}

func TestValidateOAuth2Auth_Errors(t *testing.T) {
	ctx := context.Background()
	auth := configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			TokenUrl: proto.String(""),
		}.Build(),
	}.Build()
	err := validateOAuth2Auth(ctx, auth.GetOauth2())
	assert.Error(t, err)

	auth2 := configv1.OAuth2Auth_builder{
		TokenUrl: proto.String("not-url"),
	}.Build()
	err = validateOAuth2Auth(ctx, auth2)
	assert.Error(t, err)
}

func TestValidateUpstreamAuthentication_AllTypes(t *testing.T) {
	ctx := context.Background()
	var auth *configv1.Authentication
	// Oauth2
	auth = configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			TokenUrl: proto.String(""),
		}.Build(),
	}.Build()
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)

	// Mtls
	auth = configv1.Authentication_builder{
		Mtls: configv1.MTLSAuth_builder{
			ClientCertPath: proto.String(""),
		}.Build(),
	}.Build()
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
}

func TestValidateGCSettings(t *testing.T) {
	// Case 1: Invalid interval
	gc := configv1.GCSettings_builder{}.Build()
	gc.SetInterval("invalid")
	err := validateGCSettings(context.Background(), gc)
	assert.Error(t, err)

	// Case 2: Invalid TTL
	gc = configv1.GCSettings_builder{}.Build()
	gc.SetTtl("invalid")
	err = validateGCSettings(context.Background(), gc)
	assert.Error(t, err)

	// Case 3: Empty path in Paths
	gc = configv1.GCSettings_builder{
		Enabled: proto.Bool(true),
		Paths:   []string{""},
	}.Build()
	err = validateGCSettings(context.Background(), gc)
	assert.Error(t, err)

	// Case 4: Relative path
	gc = configv1.GCSettings_builder{
		Enabled: proto.Bool(true),
		Paths:   []string{"relative/path"},
	}.Build()
	err = validateGCSettings(context.Background(), gc)
	assert.Error(t, err)
}

func TestValidateHTTPService_SchemaErrors(t *testing.T) {
	// Invalid Input Schema
	s := configv1.HttpUpstreamService_builder{
		Address: proto.String("http://example.com"),
		Calls: map[string]*configv1.HttpCallDefinition{
			"call1": configv1.HttpCallDefinition_builder{
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
					},
				},
			}.Build(),
		},
	}.Build()
	err := validateHTTPService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input_schema")
}

func TestValidateAPIKeyAuth_Errors(t *testing.T) {
	ctx := context.Background()
	// Value missing
	auth := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("api_key"),
		}.Build(),
	}.Build()
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "required for outgoing")

	// Value resolves to empty
	auth = configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("api_key"),
			Value: configv1.SecretValue_builder{
				PlainText: proto.String(""),
			}.Build(),
		}.Build(),
	}.Build()
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolved api key value is empty")
}

func TestValidateAPIKeyAuth_Incoming_Errors(t *testing.T) {
	ctx := context.Background()
	// Both Value and VerificationValue missing
	auth := configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("api_key"),
		}.Build(),
	}.Build()
	// Use validateAuthentication to access internal logic via authCtx
	err := validateAuthentication(ctx, auth, AuthValidationContextIncoming)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "api key configuration is empty")

	// VerificationValue present (Valid for Incoming)
	auth.GetApiKey().SetVerificationValue("static-key")
	err = validateAuthentication(ctx, auth, AuthValidationContextIncoming)
	assert.NoError(t, err)
}

func TestValidateMcpService_BundleErrors(t *testing.T) {
	// Empty Bundle Path
	s := configv1.McpUpstreamService_builder{
		BundleConnection: configv1.McpBundleConnection_builder{
			BundlePath: proto.String(""),
		}.Build(),
	}.Build()
	err := validateMcpService(context.Background(), s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty bundle_path")
}

func TestValidateSQLService_SchemaErrors(t *testing.T) {
	// Invalid Input Schema
	s := configv1.SqlUpstreamService_builder{
		Driver: proto.String("postgres"),
		Dsn:    proto.String("postgres://"),
		Calls: map[string]*configv1.SqlCallDefinition{
			"call1": configv1.SqlCallDefinition_builder{
				Query: proto.String("SELECT 1"),
				InputSchema: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
					},
				},
			}.Build(),
		},
	}.Build()
	err := validateSQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input_schema")
}

func TestValidate_ClientErrors(t *testing.T) {
	cfg := configv1.McpAnyServerConfig_builder{
		GlobalSettings: configv1.GlobalSettings_builder{
			ApiKey: proto.String("short"),
		}.Build(),
	}.Build()
	errs := Validate(context.Background(), cfg, Client)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Err.Error(), "at least 16 characters")
}

func TestValidateCollection_Coverage(t *testing.T) {
	ctx := context.Background()

	// 1. Invalid Name
	coll := configv1.Collection_builder{
		Name:    proto.String(""),
		HttpUrl: proto.String("http://example.com/collection.json"),
	}.Build()
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for empty name")
	}

	// 2. Invalid URL
	coll.SetName("valid-name")
	coll.SetHttpUrl("not-a-url")
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for invalid URL")
	}

	// 3. Invalid Scheme
	coll.SetHttpUrl("ftp://example.com")
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for invalid scheme")
	}

	// 4. Valid with Auth (ApiKey)
	coll.SetHttpUrl("http://example.com/collection.json")
	coll.SetAuthentication(configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("x-api-key"),
			Value: configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build(),
		}.Build(),
	}.Build())

	if err := validateCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid API Key
	coll.SetAuthentication(configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String(""), // Empty ParamName
			Value: configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build(),
		}.Build(),
	}.Build())
	assert.Error(t, validateCollection(ctx, coll))

	coll.SetAuthentication(configv1.Authentication_builder{
		ApiKey: configv1.APIKeyAuth_builder{
			ParamName: proto.String("header"),
		}.Build(),
	}.Build())
	assert.Error(t, validateCollection(ctx, coll))

	// 5. Valid with Bearer
	coll.SetAuthentication(configv1.Authentication_builder{
		BearerToken: configv1.BearerTokenAuth_builder{
			Token: configv1.SecretValue_builder{
				PlainText: proto.String("token"),
			}.Build(),
		}.Build(),
	}.Build())

	if err := validateCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// 7. Valid with mTLS (failing due to missing files, but covering code path)
	coll.SetAuthentication(configv1.Authentication_builder{
		Mtls: configv1.MTLSAuth_builder{
			ClientCertPath: proto.String("/tmp/nonexistent_cert.pem"),
			ClientKeyPath:  proto.String("/tmp/nonexistent_key.pem"),
		}.Build(),
	}.Build())
	// Should fail file check
	if err := validateCollection(ctx, coll); err == nil {
		t.Error("Expected error for missing mTLS files")
	}

	// 8. Valid with OAuth2
	coll.SetAuthentication(configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			TokenUrl: proto.String("https://example.com/token"),
			ClientId: configv1.SecretValue_builder{
				PlainText: proto.String("id"),
			}.Build(),
			ClientSecret: configv1.SecretValue_builder{
				PlainText: proto.String("secret"),
			}.Build(),
		}.Build(),
	}.Build())
	if err := validateCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Invalid OAuth2
	coll.SetAuthentication(configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			TokenUrl: proto.String(""), // Empty URL
		}.Build(),
	}.Build())
	assert.Error(t, validateCollection(ctx, coll))

	coll.SetAuthentication(configv1.Authentication_builder{
		Oauth2: configv1.OAuth2Auth_builder{
			TokenUrl: proto.String("not-url"), // Invalid URL
		}.Build(),
	}.Build())
	assert.Error(t, validateCollection(ctx, coll))
}
