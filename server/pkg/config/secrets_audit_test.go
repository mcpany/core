// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func newCmdCall(args []string, params []*configv1.CommandLineParameterMapping) *configv1.CommandLineCallDefinition {
	c := &configv1.CommandLineCallDefinition{}
	c.SetArgs(args)
	c.SetParameters(params)
	return c
}

func newCmdParam(secret *configv1.SecretValue) *configv1.CommandLineParameterMapping {
	p := &configv1.CommandLineParameterMapping{}
	p.SetSecret(secret)
	return p
}

func newWebhook(url, secret string) *configv1.WebhookConfig {
	w := &configv1.WebhookConfig{}
	w.SetUrl(url)
	w.SetWebhookSecret(secret)
	return w
}

func newSemantic(apiKey, openaiKey *configv1.SecretValue) *configv1.SemanticCacheConfig {
	s := &configv1.SemanticCacheConfig{}
	s.SetApiKey(apiKey)

	prov := &configv1.OpenAIEmbeddingProviderConfig{}
	prov.SetApiKey(openaiKey)
	s.SetOpenai(prov)
	return s
}

func newCacheConfig(semantic *configv1.SemanticCacheConfig) *configv1.CacheConfig {
	c := &configv1.CacheConfig{}
	c.SetSemanticConfig(semantic)
	return c
}

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
					"call1": newCmdCall(nil, []*configv1.CommandLineParameterMapping{
						newCmdParam(&configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{PlainText: "param-secret"},
						}),
					}),
				},
			},
		},
		// 4. Hooks
		PreCallHooks: []*configv1.CallHook{
			{
				HookConfig: &configv1.CallHook_Webhook{
					Webhook: newWebhook("http://hook.com", "webhook-secret-string"),
				},
			},
		},
		// 5. Cache
		Cache: newCacheConfig(newSemantic(
			&configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "cache-api-key"}},
			&configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "openai-api-key"}},
		)),
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
	assert.Nil(t, svc.GetAuthentication().GetBasicAuth().GetPassword().GetValue(), "Incoming auth password should be stripped")

	// Verify 2. Env vars
	cmdSvc := svc.GetServiceConfig().(*configv1.UpstreamServiceConfig_CommandLineService).CommandLineService
	assert.Nil(t, cmdSvc.GetEnv()["API_KEY"].GetValue(), "Command line env secret should be stripped")

	// Verify 3. Calls
	assert.Nil(t, cmdSvc.GetCalls()["call1"].GetParameters()[0].GetSecret().GetValue(), "Call parameter secret should be stripped")

	// Verify 4. Hooks
	webhook := svc.GetPreCallHooks()[0].GetWebhook()
	assert.Equal(t, "", webhook.GetWebhookSecret(), "Webhook secret string should be empty")

	// Verify 5. Cache
	assert.Nil(t, svc.GetCache().GetSemanticConfig().GetApiKey().GetValue(), "Cache API key should be stripped")
	assert.Nil(t, svc.GetCache().GetSemanticConfig().GetOpenai().GetApiKey().GetValue(), "OpenAI provider API key should be stripped")

	// Verify Vector DB
	pinecone := vectorSvc.GetServiceConfig().(*configv1.UpstreamServiceConfig_VectorService).VectorService.GetPinecone()
	assert.Equal(t, "", pinecone.GetApiKey(), "Pinecone API key should be empty")

    // Verify Filesystem
    s3 := fsSvc.GetServiceConfig().(*configv1.UpstreamServiceConfig_FilesystemService).FilesystemService.GetS3()
    assert.Equal(t, "", s3.GetSecretAccessKey(), "S3 Secret Access Key should be empty")
}
