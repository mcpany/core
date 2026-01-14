package config

import (
	"context"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestValidateGlobalSettings_Client(t *testing.T) {
	ctx := context.Background()

	// Client with short API key
	cfg := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			ApiKey: proto.String("short"),
		},
	}
	errs := Validate(ctx, cfg, Client)
	assert.NotEmpty(t, errs)
	assert.Contains(t, errs[0].Err.Error(), "API key must be at least 16 characters long")

    // Client with valid API key
    cfgValid := &configv1.McpAnyServerConfig{
		GlobalSettings: &configv1.GlobalSettings{
			ApiKey: proto.String("1234567890123456"),
		},
	}
    errs = Validate(ctx, cfgValid, Client)
    assert.Empty(t, errs)
}

func TestValidateGlobalSettings_Audit(t *testing.T) {
    ctx := context.Background()
    stFile := configv1.AuditConfig_STORAGE_TYPE_FILE
    // File storage without path
    cfg := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            Audit: &configv1.AuditConfig{
                Enabled: proto.Bool(true),
                StorageType: &stFile,
                // Missing OutputPath
            },
        },
    }
    errs := Validate(ctx, cfg, Server)
    assert.NotEmpty(t, errs)

    stWebhook := configv1.AuditConfig_STORAGE_TYPE_WEBHOOK
    // Webhook without URL
    cfg2 := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            Audit: &configv1.AuditConfig{
                Enabled: proto.Bool(true),
                StorageType: &stWebhook,
            },
        },
    }
    errs = Validate(ctx, cfg2, Server)
    assert.NotEmpty(t, errs)
}

func TestValidateGlobalSettings_DLP(t *testing.T) {
    ctx := context.Background()
    cfg := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            Dlp: &configv1.DLPConfig{
                Enabled: proto.Bool(true),
                CustomPatterns: []string{"["}, // Invalid Regex
            },
        },
    }
    errs := Validate(ctx, cfg, Server)
    assert.NotEmpty(t, errs)
}

func TestValidateGlobalSettings_GC(t *testing.T) {
    ctx := context.Background()
    cfg := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            GcSettings: &configv1.GCSettings{
                Interval: proto.String("invalid"),
            },
        },
    }
    errs := Validate(ctx, cfg, Server)
    assert.NotEmpty(t, errs)

    cfg2 := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            GcSettings: &configv1.GCSettings{
                Ttl: proto.String("invalid"),
            },
        },
    }
    errs = Validate(ctx, cfg2, Server)
    assert.NotEmpty(t, errs)

    cfg3 := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            GcSettings: &configv1.GCSettings{
                Enabled: proto.Bool(true),
                Paths: []string{""}, // Empty path
            },
        },
    }
    errs = Validate(ctx, cfg3, Server)
    assert.NotEmpty(t, errs)
}

func TestValidateGlobalSettings_Profile(t *testing.T) {
    ctx := context.Background()
    // Empty profile name
    cfg := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            ProfileDefinitions: []*configv1.ProfileDefinition{
                {Name: proto.String("")},
            },
        },
    }
    errs := Validate(ctx, cfg, Server)
    assert.NotEmpty(t, errs)

    // Duplicate profile name
    cfg2 := &configv1.McpAnyServerConfig{
        GlobalSettings: &configv1.GlobalSettings{
            ProfileDefinitions: []*configv1.ProfileDefinition{
                {Name: proto.String("p1")},
                {Name: proto.String("p1")},
            },
        },
    }
    errs = Validate(ctx, cfg2, Server)
    assert.NotEmpty(t, errs)
}

func TestValidateUser_Duplicate(t *testing.T) {
    ctx := context.Background()
    cfg := &configv1.McpAnyServerConfig{
        Users: []*configv1.User{
            {Id: proto.String("u1")},
            {Id: proto.String("u1")},
        },
    }
    errs := Validate(ctx, cfg, Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "duplicate user id")

    cfgEmpty := &configv1.McpAnyServerConfig{
        Users: []*configv1.User{
            {Id: proto.String("")},
        },
    }
    errs = Validate(ctx, cfgEmpty, Server)
    assert.NotEmpty(t, errs)
    assert.Contains(t, errs[0].Err.Error(), "user has empty id")
}
