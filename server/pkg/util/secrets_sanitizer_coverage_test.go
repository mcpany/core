package util

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromService_CommandLine(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_CommandLineService{
			CommandLineService: &configv1.CommandLineUpstreamService{
				Env: map[string]*configv1.SecretValue{
					"API_KEY": {
						Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
					},
				},
				Calls: map[string]*configv1.CommandLineCallDefinition{
					"call1": {
						Parameters: []*configv1.CommandLineParameterMapping{
							{
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "secret_param"},
								},
							},
						},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	cmd := svc.GetCommandLineService()
	assert.Nil(t, cmd.Env["API_KEY"].Value)
	assert.Nil(t, cmd.Calls["call1"].Parameters[0].Secret.Value)
}

func TestStripSecretsFromService_Http(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
			HttpService: &configv1.HttpUpstreamService{
				Calls: map[string]*configv1.HttpCallDefinition{
					"call1": {
						Parameters: []*configv1.HttpParameterMapping{
							{
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "secret_param"},
								},
							},
						},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	httpSvc := svc.GetHttpService()
	assert.Nil(t, httpSvc.Calls["call1"].Parameters[0].Secret.Value)
}

func TestStripSecretsFromService_Mcp(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: &configv1.McpUpstreamService{
				ConnectionType: &configv1.McpUpstreamService_StdioConnection{
					StdioConnection: &configv1.McpStdioConnection{
						Env: map[string]*configv1.SecretValue{
							"API_KEY": {
								Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
							},
						},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	mcp := svc.GetMcpService()
	assert.Nil(t, mcp.GetStdioConnection().Env["API_KEY"].Value)
}

func TestStripSecretsFromService_Filesystem(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_FilesystemService{
			FilesystemService: &configv1.FilesystemUpstreamService{
				FilesystemType: &configv1.FilesystemUpstreamService_S3{
					S3: &configv1.S3Fs{
						SecretAccessKey: proto.String("secret"),
						SessionToken:    proto.String("token"),
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	fs := svc.GetFilesystemService().GetS3()
	assert.Empty(t, fs.GetSecretAccessKey())
	assert.Empty(t, fs.GetSessionToken())
}

func TestStripSecretsFromService_Vector(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
			VectorService: &configv1.VectorUpstreamService{
				VectorDbType: &configv1.VectorUpstreamService_Milvus{
					Milvus: &configv1.MilvusVectorDB{
						ApiKey:   proto.String("key"),
						Password: proto.String("pass"),
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	vector := svc.GetVectorService().GetMilvus()
	assert.Empty(t, vector.GetApiKey())
	assert.Empty(t, vector.GetPassword())
}

func TestStripSecretsFromService_Websocket(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_WebsocketService{
			WebsocketService: &configv1.WebsocketUpstreamService{
				Calls: map[string]*configv1.WebsocketCallDefinition{
					"call1": {
						Parameters: []*configv1.WebsocketParameterMapping{
							{
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
								},
							},
						},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	ws := svc.GetWebsocketService()
	assert.Nil(t, ws.Calls["call1"].Parameters[0].Secret.Value)
}

func TestStripSecretsFromService_Webrtc(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
			WebrtcService: &configv1.WebrtcUpstreamService{
				Calls: map[string]*configv1.WebrtcCallDefinition{
					"call1": {
						Parameters: []*configv1.WebrtcParameterMapping{
							{
								Secret: &configv1.SecretValue{
									Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
								},
							},
						},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	webrtc := svc.GetWebrtcService()
	assert.Nil(t, webrtc.Calls["call1"].Parameters[0].Secret.Value)
}

func TestStripSecretsFromService_HooksAndCache(t *testing.T) {
	t.Parallel()

	svc := &configv1.UpstreamServiceConfig{
		PreCallHooks: []*configv1.CallHook{
			{
				HookConfig: &configv1.CallHook_Webhook{
					Webhook: &configv1.WebhookConfig{
						WebhookSecret: "secret",
					},
				},
			},
		},
		Cache: &configv1.CacheConfig{
			SemanticConfig: &configv1.SemanticCacheConfig{
				ApiKey: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
				ProviderConfig: &configv1.SemanticCacheConfig_Openai{
					Openai: &configv1.OpenAIEmbeddingProviderConfig{
						ApiKey: &configv1.SecretValue{
							Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
						},
					},
				},
			},
		},
	}

	StripSecretsFromService(svc)

	assert.Empty(t, svc.PreCallHooks[0].GetWebhook().WebhookSecret)
	assert.Nil(t, svc.Cache.SemanticConfig.ApiKey.Value)
	assert.Nil(t, svc.Cache.SemanticConfig.GetOpenai().ApiKey.Value)
}
