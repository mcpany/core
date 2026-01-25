// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"os"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func strPtrLocal(s string) *string {
	return &s
}

func TestValidateGraphQLService_ExtraCoverage(t *testing.T) {
	// 1. Hidden whitespace
	s := &configv1.GraphQLUpstreamService{Address: strPtrLocal(" http://example.com/graphql ")}
	err := validateGraphQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hidden whitespace")

	// 2. Invalid Scheme
	s = &configv1.GraphQLUpstreamService{Address: strPtrLocal("ftp://example.com/graphql")}
	err = validateGraphQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "scheme")

	// 3. Output Schema Error
	invalidSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
		},
	}
	s = &configv1.GraphQLUpstreamService{
		Address: strPtrLocal("http://example.com/graphql"),
		Calls: map[string]*configv1.GraphQLCallDefinition{
			"bad-output": {
				Id:           strPtrLocal("bad-output"),
				Query:        strPtrLocal("query { hello }"),
				OutputSchema: invalidSchema,
			},
		},
	}
	err = validateGraphQLService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output_schema error")
}

func TestValidateSchema_ExtraCoverage(t *testing.T) {
	// 1. Nil schema
	assert.NoError(t, validateSchema(nil))

	// 2. Invalid Type (not a string)
	invalidTypeSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}},
		},
	}
	err := validateSchema(invalidTypeSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "schema 'type' must be a string")

	// 3. Invalid Schema Logic (compilation error)
	invalidLogicSchema := &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"type":    {Kind: &structpb.Value_StringValue{StringValue: "number"}},
			"minimum": {Kind: &structpb.Value_StringValue{StringValue: "not-a-number"}},
		},
	}
	err = validateSchema(invalidLogicSchema)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid JSON schema")
}

func TestExpand_Coverage(t *testing.T) {
	// 1. Invalid variable start
	in := []byte("foo $1var bar")
	out, err := expand(in)
	assert.NoError(t, err)
	assert.Equal(t, "foo $1var bar", string(out))

	// 2. Variable at end
	in = []byte("foo $")
	out, err = expand(in)
	assert.NoError(t, err)
	assert.Equal(t, "foo $", string(out))

	// 3. Simple var missing
	in = []byte("foo $MISSING_VAR_SIMPLE bar")
	os.Unsetenv("MISSING_VAR_SIMPLE")
	_, err = expand(in)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "MISSING_VAR_SIMPLE is missing")

	// 4. Restricted variable
	// Assuming default policy blocks MCPANY_*
	in = []byte("foo $MCPANY_SECRET bar")
	_, err = expand(in)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "variable MCPANY_SECRET is restricted")
}

func TestResolveEnvValue_Extra(t *testing.T) {
	cfg := &configv1.McpAnyServerConfig{}

	// 1. Field not found
	val := resolveEnvValue(cfg, []string{"non_existent"}, "val")
	assert.Equal(t, "val", val)

	// 2. Path mismatch (scalar field but path continues)
	// name is scalar string
	val = resolveEnvValue(cfg, []string{"name", "subfield"}, "val")
	assert.Equal(t, "val", val)

	// 3. Scalar list index traversal (NOT last part) -> Error/Mismatch
	// tags is repeated string
	val = resolveEnvValue(cfg, []string{"tags", "0", "invalid_subfield"}, "val")
	assert.Equal(t, "val", val)

	// 4. Message list traversal
	// upstream_services is repeated message
	// path: upstream_services.0.name
	val = resolveEnvValue(cfg, []string{"upstream_services", "0", "name"}, "my-service")
	assert.Equal(t, "my-service", val)
}

func TestYamlUnmarshal_TabError(t *testing.T) {
	e := &yamlEngine{}
	b := []byte("key:\n\tvalue: 1")
	err := e.Unmarshal(b, &configv1.McpAnyServerConfig{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot contain tabs")
}

func TestAuth_SecretValidationErrors(t *testing.T) {
	ctx := context.Background()

	// API Key with invalid secret (non-existent file)
	auth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: strPtrLocal("key"),
				Value: &configv1.SecretValue{
					Value: &configv1.SecretValue_FilePath{FilePath: "nonexistent_secret_file"},
				},
			},
		},
	}
	err := validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "secret file \"nonexistent_secret_file\" does not exist")

	// Bearer Token with invalid secret (missing env)
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BearerToken{
			BearerToken: &configv1.BearerTokenAuth{
				Token: &configv1.SecretValue{
					Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING_BEARER_ENV"},
				},
			},
		},
	}
	os.Unsetenv("MISSING_BEARER_ENV")
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable \"MISSING_BEARER_ENV\" is not set")

	// Basic Auth with invalid password secret (bad remote url)
	auth = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BasicAuth{
			BasicAuth: &configv1.BasicAuth{
				Username: strPtrLocal("user"),
				Password: &configv1.SecretValue{
					Value: &configv1.SecretValue_RemoteContent{
						RemoteContent: &configv1.RemoteContent{HttpUrl: strPtrLocal("not-a-url")},
					},
				},
			},
		},
	}
	err = validateUpstreamAuthentication(ctx, auth, AuthValidationContextOutgoing)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "basic auth password validation failed")
}

func TestOAuth2_SecretValidationErrors(t *testing.T) {
	ctx := context.Background()
	// Invalid Client ID secret
	auth := &configv1.OAuth2Auth{
		TokenUrl: strPtrLocal("https://token.url"),
		ClientId: &configv1.SecretValue{
			Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING_CLIENT_ID"},
		},
	}
	os.Unsetenv("MISSING_CLIENT_ID")
	err := validateOAuth2Auth(ctx, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oauth2 client_id validation failed")

	// Invalid Client Secret secret
	auth = &configv1.OAuth2Auth{
		TokenUrl: strPtrLocal("https://token.url"),
		ClientId: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}},
		ClientSecret: &configv1.SecretValue{
			Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "MISSING_CLIENT_SECRET"},
		},
	}
	os.Unsetenv("MISSING_CLIENT_SECRET")
	err = validateOAuth2Auth(ctx, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oauth2 client_secret validation failed")
}

func TestValidateCommand_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	// 1. Abs path is dir
	err := validateCommandExists(tmpDir, "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")

	// 2. Relative path is dir
	err = validateCommandExists("./.", tmpDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is a directory")
}

func TestValidateDirectory_File(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "file")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	err = validateDirectoryExists(tmpFile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not a directory")
}

func TestFindKeyLine_Error(t *testing.T) {
	line := findKeyLine([]byte(": invalid yaml"), "key")
	assert.Equal(t, 0, line)
}

func TestMcpService_NoConnectionType(t *testing.T) {
	s := &configv1.McpUpstreamService{
		// ConnectionType is nil
	}
	err := validateMcpService(s)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "has no connection_type")
}

func TestValidateOAuth2Auth_IssuerUrl_Error(t *testing.T) {
	ctx := context.Background()
	// Empty token url AND empty issuer url
	auth := &configv1.OAuth2Auth{
		IssuerUrl: strPtrLocal(""),
	}
	err := validateOAuth2Auth(ctx, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "oauth2 token_url is empty and no issuer_url provided")

	// Invalid issuer url
	auth = &configv1.OAuth2Auth{
		IssuerUrl: strPtrLocal("not-url"),
	}
	err = validateOAuth2Auth(ctx, auth)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid oauth2 issuer_url")
}

// Tests from validator_graphql_test.go merged here

func TestValidateGraphQLService_MissingValidation(t *testing.T) {
	ctx := context.Background()

	invalidSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": 123, // Invalid type, should be string
	})

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtrLocal("graphql-service"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{
					GraphqlService: &configv1.GraphQLUpstreamService{
						Address: strPtrLocal("http://example.com/graphql"),
						Calls: map[string]*configv1.GraphQLCallDefinition{
							"bad-call": {
								Id:          strPtrLocal("bad-call"),
								Query:       strPtrLocal("query { hello }"),
								InputSchema: invalidSchema,
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(ctx, config, Server)
	assert.NotEmpty(t, errs, "Expected validation errors for invalid GraphQL call schema")
}

func TestValidateGraphQLService_Valid(t *testing.T) {
	ctx := context.Background()

	validSchema, _ := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"foo": map[string]interface{}{
				"type": "string",
			},
		},
	})

	config := &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: strPtrLocal("graphql-service-valid"),
				ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{
					GraphqlService: &configv1.GraphQLUpstreamService{
						Address: strPtrLocal("http://example.com/graphql"),
						Calls: map[string]*configv1.GraphQLCallDefinition{
							"good-call": {
								Id:          strPtrLocal("good-call"),
								Query:       strPtrLocal("query { hello }"),
								InputSchema: validSchema,
							},
						},
					},
				},
			},
		},
	}

	errs := Validate(ctx, config, Server)
	assert.Empty(t, errs, "Expected no validation errors for valid GraphQL service")
}
