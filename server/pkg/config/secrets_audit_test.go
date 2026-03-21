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
	svc := func() *configv1.UpstreamServiceConfig {
		auth := configv1.Authentication_builder{
			BasicAuth: configv1.BasicAuth_builder{
				Username: proto.String("admin"),
				Password: configv1.SecretValue_builder{
					PlainText: proto.String("incoming-secret"),
				}.Build(),
			}.Build(),
		}.Build()

		cmd := configv1.CommandLineUpstreamService_builder{
			Env: map[string]*configv1.SecretValue{
				"API_KEY": configv1.SecretValue_builder{
					PlainText: proto.String("cmd-secret"),
				}.Build(),
			},
			Calls: map[string]*configv1.CommandLineCallDefinition{
				"call1": configv1.CommandLineCallDefinition_builder{
					Parameters: []*configv1.CommandLineParameterMapping{
						configv1.CommandLineParameterMapping_builder{
							Secret: configv1.SecretValue_builder{
								PlainText: proto.String("param-secret"),
							}.Build(),
						}.Build(),
					},
				}.Build(),
			},
		}.Build()

		hook := configv1.CallHook_builder{
			Webhook: configv1.WebhookConfig_builder{
				Url:           "http://hook.com",
				WebhookSecret: "webhook-secret-string",
			}.Build(),
		}.Build()

		cache := configv1.CacheConfig_builder{
			SemanticConfig: configv1.SemanticCacheConfig_builder{
				ApiKey: configv1.SecretValue_builder{
					PlainText: proto.String("cache-api-key"),
				}.Build(),
				Openai: configv1.OpenAIEmbeddingProviderConfig_builder{
					ApiKey: configv1.SecretValue_builder{
						PlainText: proto.String("openai-api-key"),
					}.Build(),
				}.Build(),
			}.Build(),
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			Name:               proto.String("test-service-comprehensive"),
			UpstreamAuth:       auth,
			CommandLineService: cmd,
			PreCallHooks:       []*configv1.CallHook{hook},
			Cache:              cache,
		}.Build()
	}()

	// Create another service for Vector DB checking
	vectorSvc := func() *configv1.UpstreamServiceConfig {
		vec := configv1.VectorUpstreamService_builder{
			Pinecone: configv1.PineconeVectorDB_builder{
				ApiKey: proto.String("pinecone-secret-string"),
			}.Build(),
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			VectorService: vec,
		}.Build()
	}()

	// Create another service for Filesystem checking
	fsSvc := func() *configv1.UpstreamServiceConfig {
		fs := configv1.FilesystemUpstreamService_builder{
			S3: configv1.S3Fs_builder{
				SecretAccessKey: proto.String("s3-secret-string"),
			}.Build(),
		}.Build()

		return configv1.UpstreamServiceConfig_builder{
			FilesystemService: fs,
		}.Build()
	}()

	StripSecretsFromService(svc)
	StripSecretsFromService(vectorSvc)
	StripSecretsFromService(fsSvc)

	// Verify 1. Incoming Auth
	assert.Empty(t, svc.GetUpstreamAuth().GetBasicAuth().GetPassword().GetPlainText(), "Incoming auth password should be stripped")

	// Verify 2. Env vars
	cmdSvc := svc.GetCommandLineService()
	// Verify 3. Calls
	assert.Empty(t, cmdSvc.GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText(), "Call parameter secret should be stripped")

	// Verify 4. Hooks
	webhook := svc.GetPreCallHooks()[0].GetWebhook()
	assert.Equal(t, "", webhook.GetWebhookSecret(), "Webhook secret string should be empty")

	// Verify 5. Cache
	assert.Empty(t, svc.GetCache().GetSemanticConfig().GetApiKey().GetPlainText(), "Cache API key should be stripped")
	assert.Empty(t, svc.GetCache().GetSemanticConfig().GetOpenai().GetApiKey().GetPlainText(), "OpenAI provider API key should be stripped")

	// Verify Vector DB
	pinecone := vectorSvc.GetVectorService().GetPinecone()
	assert.Equal(t, "", pinecone.GetApiKey(), "Pinecone API key should be empty")

	// Verify Filesystem
	s3 := fsSvc.GetFilesystemService().GetS3()
	assert.Equal(t, "", s3.GetSecretAccessKey(), "S3 Secret Access Key should be empty")
}
