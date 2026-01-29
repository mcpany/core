package config

import (
	"context"
	"strings"
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestValidate_InvalidJsonSchema(t *testing.T) {
	// Create a schema that is syntactically correct protobuf, but semantically invalid JSON Schema
	// e.g. "minimum" should be a number, but we give it a string "ten"
	invalidSchema, err := structpb.NewStruct(map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"age": map[string]interface{}{
				"type":    "integer",
				"minimum": "ten", // Invalid! Should be a number.
			},
		},
	})
	assert.NoError(t, err)


	name := "test-service"
	address := "http://example.com"

	config := func() *configv1.McpAnyServerConfig {
		cfg := &configv1.McpAnyServerConfig{}
		svc := &configv1.UpstreamServiceConfig{}
		svc.SetName(name)

		httpSvc := &configv1.HttpUpstreamService{}
		httpSvc.SetAddress(address)
		httpSvc.SetCalls(map[string]*configv1.HttpCallDefinition{
			"test-call": {
				InputSchema: invalidSchema,
			},
		})
		svc.SetHttpService(httpSvc)

		cfg.SetUpstreamServices([]*configv1.UpstreamServiceConfig{svc})
		return cfg
	}()

	// Validate should return errors because the schema is invalid
	errs := Validate(context.Background(), config, Server)

	assert.NotEmpty(t, errs, "Validation should fail for invalid JSON schema")

	found := false
	for _, e := range errs {
		if e.ServiceName == "test-service" {
			t.Logf("Found error: %v", e.Err)
			if strings.Contains(e.Err.Error(), "invalid JSON schema") {
				found = true
			}
		}
	}
	assert.True(t, found, "Expected error 'invalid JSON schema' not found")
}
