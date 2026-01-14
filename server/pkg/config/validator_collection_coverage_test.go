package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidateUpstreamServiceCollectionAuth(t *testing.T) {
    ctx := context.Background()

    // Valid Collection with API Key
    coll := &configv1.UpstreamServiceCollection{
        Name: proto.String("coll1"),
        HttpUrl: proto.String("http://example.com"),
        Authentication: &configv1.Authentication{
             AuthMethod: &configv1.Authentication_ApiKey{
                ApiKey: &configv1.APIKeyAuth{
                    ParamName: proto.String("X-Key"),
                    Value: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "key"}},
                },
            },
        },
    }

    err := validateUpstreamServiceCollection(ctx, coll)
    assert.NoError(t, err)

    // Invalid Collection Auth (Missing API Key value)
    collInvalid := &configv1.UpstreamServiceCollection{
        Name: proto.String("coll2"),
        HttpUrl: proto.String("http://example.com"),
        Authentication: &configv1.Authentication{
             AuthMethod: &configv1.Authentication_ApiKey{
                ApiKey: &configv1.APIKeyAuth{
                    ParamName: proto.String("X-Key"),
                    // Missing Value
                },
            },
        },
    }
    err = validateUpstreamServiceCollection(ctx, collInvalid)
    assert.Error(t, err)

    // Test other auth methods for collection to boost coverage of validateUpstreamAuthentication
    // Bearer
    collBearer := &configv1.UpstreamServiceCollection{
        Name: proto.String("collBearer"),
        HttpUrl: proto.String("http://example.com"),
        Authentication: &configv1.Authentication{
             AuthMethod: &configv1.Authentication_BearerToken{
                BearerToken: &configv1.BearerTokenAuth{
                    Token: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "token"}},
                },
            },
        },
    }
    assert.NoError(t, validateUpstreamServiceCollection(ctx, collBearer))

    // Basic
    collBasic := &configv1.UpstreamServiceCollection{
        Name: proto.String("collBasic"),
        HttpUrl: proto.String("http://example.com"),
        Authentication: &configv1.Authentication{
             AuthMethod: &configv1.Authentication_BasicAuth{
                BasicAuth: &configv1.BasicAuth{
                    Username: proto.String("user"),
                    Password: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "pass"}},
                },
            },
        },
    }
    assert.NoError(t, validateUpstreamServiceCollection(ctx, collBasic))
}
