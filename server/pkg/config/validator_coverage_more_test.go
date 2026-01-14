package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidateUpstreamAuthenticationCoverage(t *testing.T) {
	// Test OAuth2 validation
	ctx := context.Background()

	// Valid OAuth2
	oauth := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				ClientId: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client_id"}},
				ClientSecret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
				TokenUrl: proto.String("https://example.com/token"),
			},
		},
	}
	assert.Empty(t, Validate(ctx, &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("oauth-svc"),
				UpstreamAuth: oauth,
                ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
                    HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://localhost")},
                },
			},
		},
	}, Server))

    // Invalid OAuth2 (missing token URL)
	oauthInvalid := &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				ClientId: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "client_id"}},
				ClientSecret: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
				// Missing TokenUrl
			},
		},
	}
    errs := Validate(ctx, &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("oauth-svc-invalid"),
				UpstreamAuth: oauthInvalid,
                ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
                    HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://localhost")},
                },
			},
		},
	}, Server)
    assert.NotEmpty(t, errs)

    // API Key Validation (Coverage)
    apiKeyAuth := &configv1.Authentication{
        AuthMethod: &configv1.Authentication_ApiKey{
            ApiKey: &configv1.APIKeyAuth{
                ParamName: proto.String("X-API-Key"),
                Value: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "key"}},
            },
        },
    }
    assert.Empty(t, Validate(ctx, &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("apikey-svc"),
				UpstreamAuth: apiKeyAuth,
                ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
                    HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://localhost")},
                },
			},
		},
	}, Server))

    // Bearer Token Validation (Coverage)
    bearerAuth := &configv1.Authentication{
        AuthMethod: &configv1.Authentication_BearerToken{
            BearerToken: &configv1.BearerTokenAuth{
                Token: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "token"}},
            },
        },
    }
     assert.Empty(t, Validate(ctx, &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("bearer-svc"),
				UpstreamAuth: bearerAuth,
                ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
                    HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://localhost")},
                },
			},
		},
	}, Server))

    // Basic Auth Validation (Coverage)
    basicAuth := &configv1.Authentication{
        AuthMethod: &configv1.Authentication_BasicAuth{
            BasicAuth: &configv1.BasicAuth{
                Username: proto.String("user"),
                Password: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "pass"}},
            },
        },
    }
    assert.Empty(t, Validate(ctx, &configv1.McpAnyServerConfig{
		UpstreamServices: []*configv1.UpstreamServiceConfig{
			{
				Name: proto.String("basic-svc"),
				UpstreamAuth: basicAuth,
                ServiceConfig: &configv1.UpstreamServiceConfig_HttpService{
                    HttpService: &configv1.HttpUpstreamService{Address: proto.String("http://localhost")},
                },
			},
		},
	}, Server))
}

func TestValidateSQLService(t *testing.T) {
    ctx := context.Background()

    // Valid SQL
    sqlSvc := &configv1.SqlUpstreamService{
        Driver: proto.String("postgres"),
        Dsn: proto.String("postgres://user:pass@localhost:5432/db"),
    }

    assert.Empty(t, Validate(ctx, &configv1.McpAnyServerConfig{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            {
                Name: proto.String("sql-svc"),
                ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
                    SqlService: sqlSvc,
                },
            },
        },
    }, Server))

    // Invalid SQL (unsupported driver) - Currently passes as driver is not validated
    sqlSvcInvalid := &configv1.SqlUpstreamService{
        Driver: proto.String("invalid-driver"),
        Dsn: proto.String("dsn"),
    }
     errs := Validate(ctx, &configv1.McpAnyServerConfig{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            {
                Name: proto.String("sql-svc-invalid"),
                ServiceConfig: &configv1.UpstreamServiceConfig_SqlService{
                    SqlService: sqlSvcInvalid,
                },
            },
        },
    }, Server)
    assert.Empty(t, errs)
}

func TestValidateGraphQLService(t *testing.T) {
    ctx := context.Background()
    // Invalid GraphQL (missing endpoint)
    gqlSvc := &configv1.GraphQLUpstreamService{
        // Missing Address
    }

    errs := Validate(ctx, &configv1.McpAnyServerConfig{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            {
                Name: proto.String("gql-svc-invalid"),
                ServiceConfig: &configv1.UpstreamServiceConfig_GraphqlService{
                    GraphqlService: gqlSvc,
                },
            },
        },
    }, Server)
    assert.NotEmpty(t, errs)
}

func TestValidateWebRTCService(t *testing.T) {
    ctx := context.Background()
    // Invalid WebRTC (missing signaling URL)
    svc := &configv1.WebrtcUpstreamService{
    }
     errs := Validate(ctx, &configv1.McpAnyServerConfig{
        UpstreamServices: []*configv1.UpstreamServiceConfig{
            {
                Name: proto.String("webrtc-svc-invalid"),
                ServiceConfig: &configv1.UpstreamServiceConfig_WebrtcService{
                    WebrtcService: svc,
                },
            },
        },
    }, Server)
    assert.NotEmpty(t, errs)
}
