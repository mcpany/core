package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
)

func TestStripSecretsFromService_Comprehensive(t *testing.T) {
	svc := func() *configv1.UpstreamServiceConfig {
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName("test-service-comprehensive")

		auth := &configv1.Authentication{}
		basic := &configv1.BasicAuth{}
		basic.SetUsername("admin")
		secret := &configv1.SecretValue{}
		secret.SetPlainText("incoming-secret")
		basic.SetPassword(secret)
		auth.SetBasicAuth(basic)
		svc.SetAuthentication(auth)

		cmd := &configv1.CommandLineUpstreamService{}
		cmdSecret := &configv1.SecretValue{}
		cmdSecret.SetPlainText("cmd-secret")
		cmd.SetEnv(map[string]*configv1.SecretValue{
			"API_KEY": cmdSecret,
		})

		paramSecret := &configv1.SecretValue{}
		paramSecret.SetPlainText("param-secret")
		cmd.SetCalls(map[string]*configv1.CommandLineCallDefinition{
			"call1": {
				Parameters: []*configv1.CommandLineParameterMapping{
					{
						Secret: paramSecret,
					},
				},
			},
		})
		svc.SetCommandLineService(cmd)

		hook := &configv1.CallHook{}
		webhook := &configv1.WebhookConfig{}
		webhook.SetUrl("http://hook.com")
		webhook.SetWebhookSecret("webhook-secret-string")
		hook.SetWebhook(webhook)
		svc.SetPreCallHooks([]*configv1.CallHook{hook})

		cache := &configv1.CacheConfig{}
		semantic := &configv1.SemanticCacheConfig{}
		cacheKey := &configv1.SecretValue{}
		cacheKey.SetPlainText("cache-api-key")
		semantic.SetApiKey(cacheKey)

		openai := &configv1.OpenAIEmbeddingProviderConfig{}
		openaiKey := &configv1.SecretValue{}
		openaiKey.SetPlainText("openai-api-key")
		openai.SetApiKey(openaiKey)
		semantic.SetOpenai(openai)
		cache.SetSemanticConfig(semantic)
		svc.SetCache(cache)

		return svc
	}()

	// Create another service for Vector DB checking
	vectorSvc := func() *configv1.UpstreamServiceConfig {
		svc := &configv1.UpstreamServiceConfig{}
		vec := &configv1.VectorUpstreamService{}
		pine := &configv1.PineconeVectorDB{}
		pine.SetApiKey("pinecone-secret-string")
		vec.SetPinecone(pine)
		svc.SetVectorService(vec)
		return svc
	}()

    // Create another service for Filesystem checking
    fsSvc := func() *configv1.UpstreamServiceConfig {
		svc := &configv1.UpstreamServiceConfig{}
		fs := &configv1.FilesystemUpstreamService{}
		s3 := &configv1.S3Fs{}
		s3.SetSecretAccessKey("s3-secret-string")
		fs.SetS3(s3)
		svc.SetFilesystemService(fs)
		return svc
	}()

	StripSecretsFromService(svc)
	StripSecretsFromService(vectorSvc)
    StripSecretsFromService(fsSvc)

	// Verify 1. Incoming Auth
	// Verify 1. Incoming Auth
	assert.Empty(t, svc.GetAuthentication().GetBasicAuth().GetPassword().GetPlainText(), "Incoming auth password should be stripped")

	// Verify 2. Env vars
	cmdSvc := svc.GetCommandLineService()
	// Verify 3. Calls
	assert.Empty(t, cmdSvc.GetCalls()["call1"].GetParameters()[0].GetSecret().GetPlainText(), "Call parameter secret should be stripped")

	// Verify 4. Hooks
	webhook := svc.GetPreCallHooks()[0].GetWebhook()
	assert.Equal(t, "", webhook.GetWebhookSecret(), "Webhook secret string should be empty")

	// Verify 5. Cache
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
