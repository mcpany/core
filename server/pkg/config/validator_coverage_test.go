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
	err := validateSecretValue(sv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "empty http_url")

	// Invalid URL scheme
	sv.Value = &configv1.SecretValue_RemoteContent{
		RemoteContent: &configv1.RemoteContent{
			HttpUrl: proto.String("ftp://example.com/secret"),
		},
	}
	err = validateSecretValue(sv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid http_url scheme")
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
	err := validateUpstreamAuthentication(ctx, auth)
	assert.Error(t, err)

	// Bearer Token errors
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BearerToken{
			BearerToken: &configv1.BearerTokenAuth{
				Token: nil, // Error: token empty
			},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth)
	assert.Error(t, err)

	// Basic Auth errors
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BasicAuth{
			BasicAuth: &configv1.BasicAuth{
				Username: proto.String(""), // Error
			},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth)
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
	err := validateUpstreamAuthentication(ctx, auth)
	assert.Error(t, err)

	// Mtls
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Mtls{
			Mtls: &configv1.MTLSAuth{ClientCertPath: proto.String("")},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth)
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
	err := validateUpstreamAuthentication(ctx, auth)
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
	err = validateUpstreamAuthentication(ctx, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "resolved api key value is empty")
}

func TestValidateMcpService_BundleErrors(t *testing.T) {
	// Empty Bundle Path
	s := &configv1.McpUpstreamService{
		ConnectionType: &configv1.McpUpstreamService_BundleConnection{
			BundleConnection: &configv1.McpBundleConnection{BundlePath: proto.String("")},
		},
	}
	err := validateMcpService(s)
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
				ParamName: proto.String(""), // Error
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
				Token: nil, // Error: token empty
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
