// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestStripSecretsFromService(t *testing.T) {
	tests := []struct {
		name     string
		input    *configv1.UpstreamServiceConfig
		validate func(t *testing.T, svc *configv1.UpstreamServiceConfig)
	}{
		{
			name:  "nil input",
			input: nil,
			validate: func(t *testing.T, svc *configv1.UpstreamServiceConfig) {
				assert.Nil(t, svc)
			},
		},
		{
			name: "http service with secrets",
			input: &configv1.UpstreamServiceConfig{
				ServiceType: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Calls: []*configv1.HttpCallDefinition{
							{
								Parameters: []*configv1.HttpParameterMapping{
									{
										Secret: &configv1.SecretValue{
											Value: &configv1.SecretValue_PlainText{
												PlainText: "sensitive-api-key",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, svc *configv1.UpstreamServiceConfig) {
				httpSvc := svc.GetHttpService()
				assert.NotNil(t, httpSvc)
				assert.Len(t, httpSvc.GetCalls(), 1)
				assert.Len(t, httpSvc.GetCalls()[0].GetParameters(), 1)
				assert.Empty(t, httpSvc.GetCalls()[0].GetParameters()[0].GetSecret().GetPlainText())
			},
		},
		{
			name: "service with auth",
			input: &configv1.UpstreamServiceConfig{
				UpstreamAuth: &configv1.Authentication{
					AuthType: &configv1.Authentication_ApiKey{
						ApiKey: &configv1.ApiKeyAuth{
							Value: &configv1.SecretValue{
								Value: &configv1.SecretValue_PlainText{
									PlainText: "my-api-key",
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, svc *configv1.UpstreamServiceConfig) {
				auth := svc.GetUpstreamAuth()
				assert.NotNil(t, auth)
				apiKey := auth.GetApiKey()
				assert.NotNil(t, apiKey)
				assert.Empty(t, apiKey.GetValue().GetPlainText())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StripSecretsFromService(tt.input)
			tt.validate(t, tt.input)
		})
	}
}

func TestStripSecretsFromProfile(t *testing.T) {
	tests := []struct {
		name     string
		input    *configv1.ProfileDefinition
		validate func(t *testing.T, profile *configv1.ProfileDefinition)
	}{
		{
			name:  "nil input",
			input: nil,
			validate: func(t *testing.T, profile *configv1.ProfileDefinition) {
				assert.Nil(t, profile)
			},
		},
		{
			name: "profile with collection and service",
			input: &configv1.ProfileDefinition{
				Collections: []*configv1.Collection{
					{
						Services: []*configv1.UpstreamServiceConfig{
							{
								UpstreamAuth: &configv1.Authentication{
									AuthType: &configv1.Authentication_BasicAuth{
										BasicAuth: &configv1.BasicAuth{
											Password: &configv1.SecretValue{
												Value: &configv1.SecretValue_PlainText{
													PlainText: "my-password",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, profile *configv1.ProfileDefinition) {
				assert.NotNil(t, profile)
				assert.Len(t, profile.GetCollections(), 1)
				assert.Len(t, profile.GetCollections()[0].GetServices(), 1)
				auth := profile.GetCollections()[0].GetServices()[0].GetUpstreamAuth()
				assert.NotNil(t, auth)
				basic := auth.GetBasicAuth()
				assert.NotNil(t, basic)
				assert.Empty(t, basic.GetPassword().GetPlainText())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StripSecretsFromProfile(tt.input)
			tt.validate(t, tt.input)
		})
	}
}

func TestStripSecretsFromCollection(t *testing.T) {
	tests := []struct {
		name     string
		input    *configv1.Collection
		validate func(t *testing.T, collection *configv1.Collection)
	}{
		{
			name:  "nil input",
			input: nil,
			validate: func(t *testing.T, collection *configv1.Collection) {
				assert.Nil(t, collection)
			},
		},
		{
			name: "collection with multiple services",
			input: &configv1.Collection{
				Services: []*configv1.UpstreamServiceConfig{
					{
						UpstreamAuth: &configv1.Authentication{
							AuthType: &configv1.Authentication_BearerToken{
								BearerToken: &configv1.BearerAuth{
									Token: &configv1.SecretValue{
										Value: &configv1.SecretValue_PlainText{
											PlainText: "bearer-token-1",
										},
									},
								},
							},
						},
					},
					{
						ServiceType: &configv1.UpstreamServiceConfig_CommandLineService{
							CommandLineService: &configv1.CommandLineUpstreamService{
								Env: map[string]*configv1.SecretValue{
									"API_KEY": {
										Value: &configv1.SecretValue_PlainText{
											PlainText: "secret-env-value",
										},
									},
								},
							},
						},
					},
				},
			},
			validate: func(t *testing.T, collection *configv1.Collection) {
				assert.NotNil(t, collection)
				assert.Len(t, collection.GetServices(), 2)

				bearer := collection.GetServices()[0].GetUpstreamAuth().GetBearerToken()
				assert.NotNil(t, bearer)
				assert.Empty(t, bearer.GetToken().GetPlainText())

				cmd := collection.GetServices()[1].GetCommandLineService()
				assert.NotNil(t, cmd)
				assert.Contains(t, cmd.GetEnv(), "API_KEY")
				assert.Empty(t, cmd.GetEnv()["API_KEY"].GetPlainText())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StripSecretsFromCollection(tt.input)
			tt.validate(t, tt.input)
		})
	}
}

func TestStripSecretsFromAuth(t *testing.T) {
	tests := []struct {
		name     string
		input    *configv1.Authentication
		validate func(t *testing.T, auth *configv1.Authentication)
	}{
		{
			name:  "nil input",
			input: nil,
			validate: func(t *testing.T, auth *configv1.Authentication) {
				assert.Nil(t, auth)
			},
		},
		{
			name: "api key auth with verification value",
			input: &configv1.Authentication{
				AuthType: &configv1.Authentication_ApiKey{
					ApiKey: &configv1.ApiKeyAuth{
						Value: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{
								PlainText: "super-secret-key",
							},
						},
						VerificationValue: "verification-data",
					},
				},
			},
			validate: func(t *testing.T, auth *configv1.Authentication) {
				apiKey := auth.GetApiKey()
				assert.NotNil(t, apiKey)
				assert.Empty(t, apiKey.GetValue().GetPlainText())
				assert.Empty(t, apiKey.GetVerificationValue())
			},
		},
		{
			name: "oauth2 auth",
			input: &configv1.Authentication{
				AuthType: &configv1.Authentication_Oauth2{
					Oauth2: &configv1.OAuth2Config{
						ClientId: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{
								PlainText: "client-id",
							},
						},
						ClientSecret: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{
								PlainText: "client-secret",
							},
						},
					},
				},
			},
			validate: func(t *testing.T, auth *configv1.Authentication) {
				oauth := auth.GetOauth2()
				assert.NotNil(t, oauth)
				assert.Empty(t, oauth.GetClientId().GetPlainText())
				assert.Empty(t, oauth.GetClientSecret().GetPlainText())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StripSecretsFromAuth(tt.input)
			tt.validate(t, tt.input)
		})
	}
}

func TestHydrateSecretsInService(t *testing.T) {
	tests := []struct {
		name     string
		input    *configv1.UpstreamServiceConfig
		secrets  map[string]*configv1.SecretValue
		validate func(t *testing.T, svc *configv1.UpstreamServiceConfig)
	}{
		{
			name:    "nil input",
			input:   nil,
			secrets: map[string]*configv1.SecretValue{},
			validate: func(t *testing.T, svc *configv1.UpstreamServiceConfig) {
				assert.Nil(t, svc)
			},
		},
		{
			name: "hydrate env in command line service",
			input: &configv1.UpstreamServiceConfig{
				ServiceType: &configv1.UpstreamServiceConfig_CommandLineService{
					CommandLineService: &configv1.CommandLineUpstreamService{
						Env: map[string]*configv1.SecretValue{
							"API_TOKEN": {
								Value: &configv1.SecretValue_EnvironmentVariable{
									EnvironmentVariable: "GLOBAL_API_TOKEN",
								},
							},
						},
					},
				},
			},
			secrets: map[string]*configv1.SecretValue{
				"GLOBAL_API_TOKEN": {
					Value: &configv1.SecretValue_PlainText{
						PlainText: "hydrated-token-value",
					},
				},
			},
			validate: func(t *testing.T, svc *configv1.UpstreamServiceConfig) {
				cmd := svc.GetCommandLineService()
				assert.NotNil(t, cmd)
				assert.Contains(t, cmd.GetEnv(), "API_TOKEN")
				// When hydrated, the PlainText value is set.
				assert.Equal(t, "hydrated-token-value", cmd.GetEnv()["API_TOKEN"].GetPlainText())
			},
		},
		{
			name: "hydrate http service params",
			input: &configv1.UpstreamServiceConfig{
				ServiceType: &configv1.UpstreamServiceConfig_HttpService{
					HttpService: &configv1.HttpUpstreamService{
						Calls: []*configv1.HttpCallDefinition{
							{
								Parameters: []*configv1.HttpParameterMapping{
									{
										Secret: &configv1.SecretValue{
											Value: &configv1.SecretValue_EnvironmentVariable{
												EnvironmentVariable: "SECRET_PARAM",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			secrets: map[string]*configv1.SecretValue{
				"SECRET_PARAM": {
					Value: &configv1.SecretValue_PlainText{
						PlainText: "hydrated-secret-param",
					},
				},
			},
			validate: func(t *testing.T, svc *configv1.UpstreamServiceConfig) {
				httpSvc := svc.GetHttpService()
				assert.NotNil(t, httpSvc)
				assert.Len(t, httpSvc.GetCalls(), 1)
				assert.Len(t, httpSvc.GetCalls()[0].GetParameters(), 1)
				assert.Equal(t, "hydrated-secret-param", httpSvc.GetCalls()[0].GetParameters()[0].GetSecret().GetPlainText())
			},
		},
		{
			name: "hydrate auth",
			input: &configv1.UpstreamServiceConfig{
				UpstreamAuth: &configv1.Authentication{
					AuthType: &configv1.Authentication_BasicAuth{
						BasicAuth: &configv1.BasicAuth{
							Password: &configv1.SecretValue{
								Value: &configv1.SecretValue_EnvironmentVariable{
									EnvironmentVariable: "DB_PASSWORD",
								},
							},
						},
					},
				},
			},
			secrets: map[string]*configv1.SecretValue{
				"DB_PASSWORD": {
					Value: &configv1.SecretValue_PlainText{
						PlainText: "my-secure-db-pwd",
					},
				},
			},
			validate: func(t *testing.T, svc *configv1.UpstreamServiceConfig) {
				auth := svc.GetUpstreamAuth()
				assert.NotNil(t, auth)
				basic := auth.GetBasicAuth()
				assert.NotNil(t, basic)
				assert.Equal(t, "my-secure-db-pwd", basic.GetPassword().GetPlainText())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HydrateSecretsInService(tt.input, tt.secrets)
			tt.validate(t, tt.input)
		})
	}
}
