package middleware

import (
	"testing"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/pkg/tool"
)

func TestInitStandardMiddlewares_RegistersNewMiddlewares(t *testing.T) {
	auditConfig := &configv1.AuditConfig{}

	stdMws, err := InitStandardMiddlewares(
		nil,                 // authManager
		nil,                 // toolManager
		auditConfig,         // auditConfig
		&CachingMiddleware{},// cachingMiddleware
		nil,                 // globalRateLimitConfig
		nil,                 // dlpConfig
		nil,                 // contextOptimizerConfig
		nil,                 // debuggerConfig
		nil,                 // smartRecoveryConfig
		nil,                 // cfiaConfig
		nil,                 // discoverySandboxConfig
	)

	if err != nil {
		t.Fatalf("Expected nil err, got %v", err)
	}

	if stdMws == nil {
		t.Fatal("Expected non-nil standard middlewares")
	}

	if stdMws.Scopes == nil {
		t.Error("Expected Scopes middleware to be initialized")
	}

	if stdMws.Blackboard == nil {
		t.Error("Expected Blackboard to be initialized")
	}

	if stdMws.LazyMCP == nil {
		t.Error("Expected LazyMCP middleware to be initialized")
	}
}
