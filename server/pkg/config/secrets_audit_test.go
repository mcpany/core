// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromService_Comprehensive(t *testing.T) {
	svc := &configv1.UpstreamServiceConfig{
		Name: proto.String("test-service-comprehensive"),
		// 1. Incoming Auth
		Authentication: &configv1.Authentication{
			AuthMethod: &configv1.Authentication_BasicAuth{
				BasicAuth: &configv1.BasicAuth{
					Username: proto.String("admin"),
					Password: &configv1.SecretValue{
						Value: &configv1.SecretValue_PlainText{PlainText: "incoming-secret"},
					},
				},
			},
		},
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				// 2. Env vars in CommandLineService
				Env: map[string]*configv1.SecretValue{
					"API_KEY": {Value: &configv1.SecretValue_PlainText{PlainText: "cmd-secret"}},
				},
				// 3. Calls with secrets
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"call1": {
						Parameters: []*configv1.CommandLineParameterMapping{
							{
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "param-secret"},
								},
							},
						},
					},
				},
			},
		},
		// 4. Hooks
		PreCallHooks: []*configv1.CallHook{
			{
				HookConfig: &configv1.CallHook_Webhook{
					Webhook: &configv1.WebhookConfig{
						Url:           "http://hook.com",
						WebhookSecret: "webhook-secret-string",
					},
				},
			},
		},
		// 5. Cache
		Cache: &configv1.CacheConfig{
			SemanticConfig: &configv1.SemanticCacheConfig{
				ApiKey: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "cache-api-key"},
				},
				ProviderConfig: &configv1.SemanticCacheConfig_Openai{
					Openai: &configv1.OpenAIEmbeddingProviderConfig{
						ApiKey: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{PlainText: "openai-api-key"},
						},
					},
				},
			},
		},
	}

	// Create another service for Vector DB checking
	vectorSvc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
			VectorService: &configv1.VectorUpstreamService{
				VectorDbType: &configv1.VectorUpstreamService_Pinecone{
					Pinecone: &configv1.PineconeVectorDB{
						ApiKey: proto.String("pinecone-secret-string"),
					},
				},
			},
		},
	}

    // Create another service for Filesystem checking
    fsSvc := &configv1.UpstreamServiceConfig{
        ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
            FilesystemService: &configv1.FilesystemUpstreamService{
                FilesystemType: &configv1.FilesystemUpstreamService_S3{
                    S3: &configv1.S3Fs{
                        SecretAccessKey: proto.String("s3-secret-string"),
                    },
                },
            },
        },
    }

	StripSecretsFromService(svc)
	StripSecretsFromService(vectorSvc)
    StripSecretsFromService(fsSvc)

	// Verify 1. Incoming Auth
	assert.Nil(t, svc.Authentication.GetBasicAuth().Password.Value, "Incoming auth password should be stripped")

	// Verify 2. Env vars
	cmdSvc := svc.GetServiceConfig().(*configv1.UpstreamServiceConfig_CommandLineService).CommandLineService
	assert.Nil(t, cmdSvc.Env["API_KEY"].Value, "Command line env secret should be stripped")

	// Verify 3. Calls
	assert.Nil(t, cmdSvc.Calls["call1"].Parameters[0].Secret.Value, "Call parameter secret should be stripped")

	// Verify 4. Hooks
	webhook := svc.PreCallHooks[0].GetWebhook()
	assert.Equal(t, "", webhook.WebhookSecret, "Webhook secret string should be empty")

	// Verify 5. Cache
	assert.Nil(t, svc.Cache.SemanticConfig.ApiKey.Value, "Cache API key should be stripped")
	assert.Nil(t, svc.Cache.SemanticConfig.GetOpenai().ApiKey.Value, "OpenAI provider API key should be stripped")

	// Verify Vector DB
	pinecone := vectorSvc.GetServiceConfig().(*configv1.UpstreamServiceConfig_VectorService).VectorService.GetPinecone()
	assert.Equal(t, "", *pinecone.ApiKey, "Pinecone API key should be empty")

    // Verify Filesystem
    s3 := fsSvc.GetServiceConfig().(*configv1.UpstreamServiceConfig_FilesystemService).FilesystemService.GetS3()
    assert.Equal(t, "", *s3.SecretAccessKey, "S3 Secret Access Key should be empty")
}
