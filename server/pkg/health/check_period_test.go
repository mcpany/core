// Copyright 2026 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package health

import (
	"testing"
	"time"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestNewChecker_WithInterval(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://localhost:1234"),
			HealthCheck: configv1.HttpHealthCheck_builder{
				Url:      proto.String("http://localhost:1234/health"),
				Interval: durationpb.New(1 * time.Second),
				Timeout:  durationpb.New(1 * time.Second),
			}.Build(),
		}.Build(),
	}.Build()

	// This should use WithPeriodicCheck
	checker := NewChecker(config)
	assert.NotNil(t, checker)

	// Check Stop exists
	if c, ok := checker.(interface{ Stop() }); ok {
		c.Stop()
	} else {
		// Should have stop
		t.Log("Checker does not implement Stop()")
	}
}

func TestNewChecker_WithoutInterval(t *testing.T) {
	config := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("test-service-no-interval"),
		HttpService: configv1.HttpUpstreamService_builder{
			Address: proto.String("http://localhost:1234"),
			HealthCheck: configv1.HttpHealthCheck_builder{
				Url:      proto.String("http://localhost:1234/health"),
				// No interval
			}.Build(),
		}.Build(),
	}.Build()

	// This should use WithCheck
	checker := NewChecker(config)
	assert.NotNil(t, checker)

	if c, ok := checker.(interface{ Stop() }); ok {
		c.Stop()
	}
}
