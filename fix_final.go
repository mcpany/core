package main

import (
	"os"
	"strings"
)

func main() {
	content, _ := os.ReadFile("server/tests/integration/e2e_helpers.go")
	s := string(content)

	// RegisterHTTPServiceWithParams
	oldParams := "	method := configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName])\n\n\tcallID := \"call-\" + toolDef.GetName()\n\tcallDef := configv1.HttpCallDefinition_builder{\n\t\tId:           &callID,\n\t\tEndpointPath: &endpointPath,\n\t\tMethod:       &method,\n\t\tParameters:   params,\n\t}.Build()\n\ttoolDef.SetCallId(callID)\n\n\tupstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{\n\t\tName: &serviceID,\n\t\tHttpService: configv1.HttpUpstreamService_builder{\n\t\t\tAddress: &baseURL,\n\t\t\tTools:   []*configv1.ToolDefinition{toolDef},\n\t\t\tCalls:   map[string]*configv1.HttpCallDefinition{callID: callDef},\n\t\t}.Build(),\n\t}"

	newParams := "	callID := \"call-\" + toolDef.GetName()\n\ttoolDef.SetCallId(callID)\n\n\tupstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{\n\t\tName: &serviceID,\n\t\tHttpService: configv1.HttpUpstreamService_builder{\n\t\t\tAddress: &baseURL,\n\t\t\tTools:   []*configv1.ToolDefinition{toolDef},\n\t\t\tCalls: map[string]*configv1.HttpCallDefinition{\n\t\t\t\tcallID: configv1.HttpCallDefinition_builder{\n\t\t\t\t\tId:           &callID,\n\t\t\t\t\tEndpointPath: &endpointPath,\n\t\t\t\t\tMethod:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName]).Enum(),\n\t\t\t\t\tParameters:   params,\n\t\t\t\t}.Build(),\n\t\t\t},\n\t\t}.Build(),\n\t}"

	s = strings.Replace(s, oldParams, newParams, 1)

	// RegisterHTTPServiceWithJSONRPC
	oldJSON := "	method := configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName])\n\n\ttoolDef := configv1.ToolDefinition_builder{Name: &operationID}.Build()\n\tcallID := \"call-\" + toolDef.GetName()\n\tcallDef := configv1.HttpCallDefinition_builder{\n\t\tId:           &callID,\n\t\tEndpointPath: &endpointPath,\n\t\tMethod:       &method,\n\t}.Build()\n\ttoolDef.SetCallId(callID)\n\n\tupstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{\n\t\tName: &serviceID,\n\t\tHttpService: configv1.HttpUpstreamService_builder{\n\t\t\tAddress: &baseURL,\n\t\t\tTools:   []*configv1.ToolDefinition{toolDef},\n\t\t\tCalls:   map[string]*configv1.HttpCallDefinition{callID: callDef},\n\t\t}.Build(),\n\t}"

	newJSON := "	toolDef := configv1.ToolDefinition_builder{Name: &operationID}.Build()\n\tcallID := \"call-\" + toolDef.GetName()\n\ttoolDef.SetCallId(callID)\n\n\tupstreamServiceConfigBuilder := configv1.UpstreamServiceConfig_builder{\n\t\tName: &serviceID,\n\t\tHttpService: configv1.HttpUpstreamService_builder{\n\t\t\tAddress: &baseURL,\n\t\t\tTools:   []*configv1.ToolDefinition{toolDef},\n\t\t\tCalls: map[string]*configv1.HttpCallDefinition{\n\t\t\t\tcallID: configv1.HttpCallDefinition_builder{\n\t\t\t\t\tId:           &callID,\n\t\t\t\t\tEndpointPath: &endpointPath,\n\t\t\t\t\tMethod:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value[httpMethodEnumName]).Enum(),\n\t\t\t\t}.Build(),\n\t\t\t},\n\t\t}.Build(),\n\t}"

	s = strings.Replace(s, oldJSON, newJSON, 1)

	os.WriteFile("server/tests/integration/e2e_helpers.go", []byte(s), 0644)
}
