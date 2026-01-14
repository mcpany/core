// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
    busv1 "github.com/mcpany/core/proto/bus"
	"google.golang.org/protobuf/proto"
    "google.golang.org/protobuf/types/known/structpb"
    "github.com/stretchr/testify/assert"
)

func TestValidateUpstreamServiceCollection_Coverage(t *testing.T) {
	ctx := context.Background()

	// 1. Invalid Name
	coll := &configv1.UpstreamServiceCollection{
		Name:    proto.String(""),
		HttpUrl: proto.String("http://example.com/collection.json"),
	}
	if err := validateUpstreamServiceCollection(ctx, coll); err == nil {
		t.Error("Expected error for empty name")
	}

	// 2. Invalid URL
	coll.Name = proto.String("valid-name")
	coll.HttpUrl = proto.String("not-a-url")
	if err := validateUpstreamServiceCollection(ctx, coll); err == nil {
		t.Error("Expected error for invalid URL")
	}

	// 3. Invalid Scheme
	coll.HttpUrl = proto.String("ftp://example.com")
	if err := validateUpstreamServiceCollection(ctx, coll); err == nil {
		t.Error("Expected error for invalid scheme")
	}

	// 4. Valid with Auth (ApiKey)
	coll.HttpUrl = proto.String("http://example.com/collection.json")
	coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String("X-API-Key"),
				Value: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
			},
		},
	}
	if err := validateUpstreamServiceCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

    // Invalid API Key
    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String(""), // Empty ParamName
				Value: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
			},
		},
	}
    assert.Error(t, validateUpstreamServiceCollection(ctx, coll))

    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_ApiKey{
			ApiKey: &configv1.APIKeyAuth{
				ParamName: proto.String("header"),
				// Missing Value
			},
		},
	}
    assert.Error(t, validateUpstreamServiceCollection(ctx, coll))

    // 5. Valid with Bearer
    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BearerToken{
			BearerToken: &configv1.BearerTokenAuth{
				Token: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
			},
		},
	}
    if err := validateUpstreamServiceCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

    // 6. Valid with Basic Auth
    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_BasicAuth{
			BasicAuth: &configv1.BasicAuth{
				Username: proto.String("user"),
				Password: &configv1.SecretValue{
					Value: &configv1.SecretValue_PlainText{PlainText: "secret"},
				},
			},
		},
	}
    if err := validateUpstreamServiceCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

    // 7. Valid with mTLS (failing due to missing files, but covering code path)
    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Mtls{
			Mtls: &configv1.MTLSAuth{
				ClientCertPath: proto.String("/tmp/nonexistent_cert.pem"),
				ClientKeyPath: proto.String("/tmp/nonexistent_key.pem"),
			},
		},
	}
    // Should fail file check
    if err := validateUpstreamServiceCollection(ctx, coll); err == nil {
		t.Error("Expected error for missing mTLS files")
	}

    // 8. Valid with OAuth2
    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				TokenUrl: proto.String("https://example.com/token"),
                ClientId: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "id"}},
                ClientSecret: &configv1.SecretValue{Value: &configv1.SecretValue_PlainText{PlainText: "secret"}},
			},
		},
	}
    if err := validateUpstreamServiceCollection(ctx, coll); err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

    // Invalid OAuth2
    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				TokenUrl: proto.String(""), // Empty URL
			},
		},
	}
    assert.Error(t, validateUpstreamServiceCollection(ctx, coll))

    coll.Authentication = &configv1.Authentication{
		AuthMethod: &configv1.Authentication_Oauth2{
			Oauth2: &configv1.OAuth2Auth{
				TokenUrl: proto.String("not-url"), // Invalid URL
			},
		},
	}
    assert.Error(t, validateUpstreamServiceCollection(ctx, coll))
}

func TestValidateSQLService_Coverage(t *testing.T) {
    // 1. Empty Driver
    svc := &configv1.SqlUpstreamService{
        Driver: proto.String(""),
        Dsn: proto.String("postgres://user:pass@localhost:5432/db"),
    }
    if err := validateSQLService(svc); err == nil {
        t.Error("Expected error for empty driver")
    }

    // 2. Empty DSN
    svc.Driver = proto.String("postgres")
    svc.Dsn = proto.String("")
    if err := validateSQLService(svc); err == nil {
        t.Error("Expected error for empty DSN")
    }

    // 3. Call with Empty Query
    svc.Dsn = proto.String("postgres://...")
    svc.Calls = map[string]*configv1.SqlCallDefinition{
        "call1": {
            Query: proto.String(""),
        },
    }
    if err := validateSQLService(svc); err == nil {
        t.Error("Expected error for empty query")
    }

    // 4. Call with Invalid Input Schema
    svc.Calls["call1"].Query = proto.String("SELECT * FROM table")
    svc.Calls["call1"].InputSchema = &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}}, // Invalid type
        },
    }
    if err := validateSQLService(svc); err == nil {
        t.Error("Expected error for invalid input schema")
    }

    // 5. Call with Invalid Output Schema
    svc.Calls["call1"].InputSchema = nil
    svc.Calls["call1"].OutputSchema = &structpb.Struct{
        Fields: map[string]*structpb.Value{
            "type": {Kind: &structpb.Value_NumberValue{NumberValue: 123}}, // Invalid type
        },
    }
    if err := validateSQLService(svc); err == nil {
        t.Error("Expected error for invalid output schema")
    }

    // 6. Valid
    svc.Calls["call1"].OutputSchema = nil
    if err := validateSQLService(svc); err != nil {
        t.Errorf("Unexpected error: %v", err)
    }
}

func TestValidateGCSettings_Coverage(t *testing.T) {
    // Invalid interval
    gc := &configv1.GCSettings{
        Interval: proto.String("invalid"),
    }
    assert.Error(t, validateGCSettings(gc))

    // Invalid TTL
    gc.Interval = proto.String("1h")
    gc.Ttl = proto.String("invalid")
    assert.Error(t, validateGCSettings(gc))

    // Enabled with empty path
    gc.Ttl = proto.String("24h")
    gc.Enabled = proto.Bool(true)
    gc.Paths = []string{""}
    assert.Error(t, validateGCSettings(gc))

    // Enabled with relative path (should fail IsAbs check)
    gc.Paths = []string{"relative/path"}
    // IsAllowedPath might fail first, but if allowed paths are not set, it might allow anything?
    // validateGCSettings calls IsAllowedPath then IsAbs.
    assert.Error(t, validateGCSettings(gc))
}

func TestValidateDLPConfig_Coverage(t *testing.T) {
    dlp := &configv1.DLPConfig{
        CustomPatterns: []string{"["}, // Invalid regex
    }
    assert.Error(t, validateDLPConfig(dlp))
}

func TestValidateGlobalSettings_Coverage(t *testing.T) {
    gs := &configv1.GlobalSettings{
        McpListenAddress: proto.String("invalid"),
    }
    assert.Error(t, validateGlobalSettings(gs, Server))

    gs.McpListenAddress = proto.String(":50050")

    // Construct MessageBus using Setters
    mb := &busv1.MessageBus{}
    rb := &busv1.RedisBus{}
    rb.SetAddress("") // Empty address
    mb.SetRedis(rb)
    gs.MessageBus = mb

    assert.Error(t, validateGlobalSettings(gs, Server))

    // Client with short API Key
    gs.MessageBus = nil
    gs.ApiKey = proto.String("short")
    assert.Error(t, validateGlobalSettings(gs, Client))

    // Profile definition duplicate
    gs.ApiKey = proto.String("")
    gs.ProfileDefinitions = []*configv1.ProfileDefinition{
        {Name: proto.String("p1")},
        {Name: proto.String("p1")},
    }
    assert.Error(t, validateGlobalSettings(gs, Server))

    // Empty profile name
     gs.ProfileDefinitions = []*configv1.ProfileDefinition{
        {Name: proto.String("")},
    }
    assert.Error(t, validateGlobalSettings(gs, Server))
}

func TestValidateWebrtcService_Coverage(t *testing.T) {
     svc := &configv1.WebrtcUpstreamService{
         Address: proto.String(""),
     }
     assert.Error(t, validateWebrtcService(svc))

     svc.Address = proto.String("not-url")
     assert.Error(t, validateWebrtcService(svc))

     svc.Address = proto.String("ftp://example.com")
     assert.Error(t, validateWebrtcService(svc))
}

func TestValidateGraphQLService_Coverage(t *testing.T) {
     svc := &configv1.GraphQLUpstreamService{
         Address: proto.String(""),
     }
     assert.Error(t, validateGraphQLService(svc))

     svc.Address = proto.String("not-url")
     assert.Error(t, validateGraphQLService(svc))

     svc.Address = proto.String("ftp://example.com")
     assert.Error(t, validateGraphQLService(svc))
}
