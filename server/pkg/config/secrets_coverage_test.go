package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestStripSecretsFromGrpcService(t *testing.T) {
	svc := &configv1.GrpcUpstreamService{
		Address: proto.String("localhost:50051"),
	}

	upstreamSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("grpc-test"),
		ServiceConfig: &configv1.UpstreamServiceConfig_GrpcService{
			GrpcService: svc,
		},
	}

	StripSecretsFromService(upstreamSvc)
	assert.Equal(t, "localhost:50051", upstreamSvc.GetGrpcService().GetAddress())
}

func TestStripSecretsFromOpenapiService(t *testing.T) {
	svc := &configv1.OpenapiUpstreamService{
		SpecSource: &configv1.OpenapiUpstreamService_SpecContent{
             SpecContent: "/path/to/spec",
        },
	}

	upstreamSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("openapi-test"),
		ServiceConfig: &configv1.UpstreamServiceConfig_OpenapiService{
			OpenapiService: svc,
		},
	}

	StripSecretsFromService(upstreamSvc)
	assert.Equal(t, "/path/to/spec", upstreamSvc.GetOpenapiService().GetSpecContent())
}

func TestStripSecretsFromMcpService(t *testing.T) {
	svc := &configv1.McpUpstreamService{
		ConnectionType: &configv1.McpUpstreamService_StdioConnection{
            StdioConnection: &configv1.McpStdioConnection{
                Command: proto.String("npx"),
                Args: []string{"-s", "secret"},
                Env: map[string]*configv1.SecretValue{
                    "API_KEY": {Value: &configv1.SecretValue_PlainText{PlainText: "secret_key"}},
                    "ENV_VAR": {Value: &configv1.SecretValue_EnvironmentVariable{EnvironmentVariable: "SOME_ENV"}},
                },
            },
        },
	}

	upstreamSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("mcp-test"),
		ServiceConfig: &configv1.UpstreamServiceConfig_McpService{
			McpService: svc,
		},
	}

	StripSecretsFromService(upstreamSvc)

    // Check that env vars are scrubbed
    env := upstreamSvc.GetMcpService().GetStdioConnection().GetEnv()
    assert.Nil(t, env["API_KEY"].GetValue())

    // ENV_VAR should be preserved because it is not PlainText
    assert.NotNil(t, env["ENV_VAR"].GetValue())
    assert.Equal(t, "SOME_ENV", env["ENV_VAR"].GetEnvironmentVariable())
}

func TestStripSecretsFromVectorService(t *testing.T) {
    svc := &configv1.VectorUpstreamService{
        VectorDbType: &configv1.VectorUpstreamService_Pinecone{
            Pinecone: &configv1.PineconeVectorDB{
                ApiKey: proto.String("secret-api-key"),
            },
        },
    }

    upstreamSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("vector-test"),
		ServiceConfig: &configv1.UpstreamServiceConfig_VectorService{
			VectorService: svc,
		},
	}

    StripSecretsFromService(upstreamSvc)
    // Should be scrubbed to empty string
    assert.Equal(t, "", upstreamSvc.GetVectorService().GetPinecone().GetApiKey())
}

func TestStripSecretsFromWebrtcService(t *testing.T) {
    svc := &configv1.WebrtcUpstreamService{
        Address: proto.String("wss://secret:pass@host"),
    }
    upstreamSvc := &configv1.UpstreamServiceConfig{
		Name: proto.String("webrtc-test"),
		ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
			WebrtcService: svc,
		},
	}

    StripSecretsFromService(upstreamSvc)
}
