package config

import (
	"testing"

	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUnmarshalServices_JSON(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	data := []byte(`{"services": [{"name": "s1", "http_service": {"address": "http://127.0.0.1"}}]}`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "application/json")
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "s1", services[0].GetName())
}

func TestUnmarshalServices_YAML(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	data := []byte(`
services:
  - name: s1
    http_service:
      address: http://127.0.0.1
`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "application/yaml")
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "s1", services[0].GetName())
}

func TestUnmarshalServices_ProtoText(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	// Protobuf text format
	data := []byte(`
services {
  name: "s1"
  http_service {
    address: "http://127.0.0.1"
  }
}
`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "text/plain")
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "s1", services[0].GetName())
}

func TestUnmarshalServices_SingleService_JSON(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	data := []byte(`{"name": "s1", "http_service": {"address": "http://127.0.0.1"}}`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "application/json")
	require.NoError(t, err)
	assert.Len(t, services, 1)
	assert.Equal(t, "s1", services[0].GetName())
}

func TestUnmarshalServices_Invalid(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	data := []byte(`invalid json`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "application/json")
	assert.Error(t, err)
}

func TestUnmarshalServices_InvalidProto(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	data := []byte(`invalid proto`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "text/plain")
	assert.Error(t, err)
}

func TestUnmarshalServices_Version(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	data := []byte(`{"version": "1.0.0", "services": []}`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "application/json")
	require.NoError(t, err)
}

func TestUnmarshalServices_InvalidVersion(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	data := []byte(`{"version": "invalid", "services": []}`)
	var services []*configv1.UpstreamServiceConfig
	err := m.unmarshalServices(data, &services, "application/json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid semantic version")
}

func TestAddService_Overrides(t *testing.T) {
	m := NewUpstreamServiceManager([]string{"prod"})
	// Setup overrides manually
	enabled := false
	override := configv1.ProfileServiceConfig_builder{
		Enabled: proto.Bool(enabled),
	}.Build()
	m.profileServiceOverrides["s1"] = override

	svc := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("s1"),
	}.Build()
	err := m.addService(svc, 0)
	require.NoError(t, err)

	// Should be skipped (not added to services map) because it's disabled
	assert.Nil(t, m.services["s1"])
}

func TestAddService_ConfigError(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	svc := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String("s1"),
		ConfigError: proto.String("some error"),
	}.Build()
	err := m.addService(svc, 0)
	require.NoError(t, err)

	// It should be added but marked disabled
	loaded := m.services["s1"]
	assert.NotNil(t, loaded)
	assert.True(t, loaded.GetDisable())
}

func TestAddService_Duplicate(t *testing.T) {
	m := NewUpstreamServiceManager(nil)
	svc1 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("s1"),
	}.Build()
	svc2 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("s1"),
	}.Build()

	err := m.addService(svc1, 10)
	require.NoError(t, err)
	assert.Equal(t, int32(10), m.servicePriorities["s1"])

	// Same priority, ignored
	err = m.addService(svc2, 10)
	require.NoError(t, err)

	// Lower priority, ignored
	svc3 := configv1.UpstreamServiceConfig_builder{
		Name: proto.String("s1"),
	}.Build()
	err = m.addService(svc3, 20) // Higher number = Lower priority?
	// Logic says:
	// case priority < existingPriority: Replace
	// case priority == existingPriority: Ignore
	// default (priority > existingPriority): Ignore
	// Wait, code says:
	/*
		case priority < existingPriority:
			// New service has higher priority, replace the old one
			// ...
		case priority == existingPriority:
			// Same priority, this is a duplicate
			// ...
		default:
			// lower priority, do nothing
	*/
	// So Lower Number = Higher Priority.

	require.NoError(t, err)
	assert.Equal(t, int32(10), m.servicePriorities["s1"])

	// Higher priority (lower number), replace
	svc4 := configv1.UpstreamServiceConfig_builder{}.Build()
	svc4.SetName("s1")
	err = m.addService(svc4, 5)
	require.NoError(t, err)
	assert.Equal(t, int32(5), m.servicePriorities["s1"])
}
