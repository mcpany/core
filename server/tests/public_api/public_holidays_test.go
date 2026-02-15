//go:build e2e_public_api

package public_api

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/mcpany/core/server/pkg/util"
	apiv1 "github.com/mcpany/core/proto/api/v1"
	configv1 "github.com/mcpany/core/proto/config/v1"
	"github.com/mcpany/core/server/tests/integration"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestUpstreamService_PublicHolidaysWithTransformation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Public Holidays API with Transformation...")
	t.Parallel()

	// 1. Start MCPANY Server
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EPublicHolidaysTest")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// 2. Register Public Holidays Service with MCPANY
	const serviceID = "e2e_public_holidays"
	serviceURL := "https://date.nager.at"
	endpointPath := "/api/v3/PublicHolidays/{{year}}/{{countryCode}}"
	operationID := "getPublicHolidays"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s%s...", serviceID, serviceURL, endpointPath)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	outputTransformer := configv1.OutputTransformer_builder{
		Format: configv1.OutputTransformer_JSON.Enum(),
		ExtractionRules: map[string]string{
			"holidayName": "{[0].name}",
			"holidayDate": "{[0].date}",
		},
		Template: proto.String("The first public holiday is {{holidayName}} on {{holidayDate}}."),
	}.Build()

	callID := "getPublicHolidays"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String(endpointPath),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("year"),
				}.Build(),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("countryCode"),
				}.Build(),
			}.Build(),
		},
		OutputTransformer: outputTransformer,
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String(operationID),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(serviceURL),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(serviceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", serviceID)

	// 3. Call Tool via MCPANY
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	sanitizedServiceID, _ := util.SanitizeServiceName(serviceID)
	sanitizedToolName, _ := util.SanitizeToolName(operationID)
	toolName := sanitizedServiceID + "." + sanitizedToolName
	toolArgs := `{"year": 2024, "countryCode": "US"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(toolArgs)})
		if err == nil {
			break // Success
		}

		// If the error is a 503 or a timeout, we can retry. Otherwise, fail fast.
		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to date.nager.at failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		// For any other error, fail the test immediately.
		require.NoError(t, err, "unrecoverable error calling getPublicHolidays tool")
	}

	require.NoError(t, err, "all retries to date.nager.at failed with transient errors")

	require.NoError(t, err, "Error calling getPublicHolidays tool")
	require.NotNil(t, res, "Nil response from getPublicHolidays tool")

	// 4. Assert Response
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var result struct {
		Result string `json:"result"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &result)
	require.NoError(t, err, "Failed to unmarshal tool result")

	expectedOutput := "The first public holiday is New Year's Day on 2024-01-01."
	require.Equal(t, expectedOutput, result.Result)

	t.Log("INFO: E2E Test Scenario for Public Holidays API with Transformation Completed Successfully!")
}

func TestUpstreamService_PublicHolidaysWithTransformation_CA_2025(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), integration.TestWaitTimeShort)
	defer cancel()

	t.Log("INFO: Starting E2E Test Scenario for Public Holidays API with Transformation for CA 2025...")
	t.Parallel()

	// 1. Start MCPANY Server
	mcpAnyTestServerInfo := integration.StartMCPANYServer(t, "E2EPublicHolidaysTest_CA_2025")
	defer mcpAnyTestServerInfo.CleanupFunc()

	// 2. Register Public Holidays Service with MCPANY
	const serviceID = "e2e_public_holidays_ca_2025"
	serviceURL := "https://date.nager.at"
	endpointPath := "/api/v3/PublicHolidays/{{year}}/{{countryCode}}"
	operationID := "getPublicHolidays"
	t.Logf("INFO: Registering '%s' with MCPANY at endpoint %s%s...", serviceID, serviceURL, endpointPath)
	registrationGRPCClient := mcpAnyTestServerInfo.RegistrationClient

	outputTransformer := configv1.OutputTransformer_builder{
		Format: configv1.OutputTransformer_JSON.Enum(),
		ExtractionRules: map[string]string{
			"holidayName": "{[0].name}",
			"holidayDate": "{[0].date}",
		},
		Template: proto.String("The first public holiday is {{holidayName}} on {{holidayDate}}."),
	}.Build()

	callID := "getPublicHolidays"
	httpCall := configv1.HttpCallDefinition_builder{
		Id:           proto.String(callID),
		EndpointPath: proto.String(endpointPath),
		Method:       configv1.HttpCallDefinition_HttpMethod(configv1.HttpCallDefinition_HttpMethod_value["HTTP_METHOD_GET"]).Enum(),
		Parameters: []*configv1.HttpParameterMapping{
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("year"),
				}.Build(),
			}.Build(),
			configv1.HttpParameterMapping_builder{
				Schema: configv1.ParameterSchema_builder{
					Name: proto.String("countryCode"),
				}.Build(),
			}.Build(),
		},
		OutputTransformer: outputTransformer,
	}.Build()

	toolDef := configv1.ToolDefinition_builder{
		Name:   proto.String(operationID),
		CallId: proto.String(callID),
	}.Build()

	httpService := configv1.HttpUpstreamService_builder{
		Address: proto.String(serviceURL),
		Tools:   []*configv1.ToolDefinition{toolDef},
		Calls:   map[string]*configv1.HttpCallDefinition{callID: httpCall},
	}.Build()

	config := configv1.UpstreamServiceConfig_builder{
		Name:        proto.String(serviceID),
		HttpService: httpService,
	}.Build()

	req := apiv1.RegisterServiceRequest_builder{
		Config: config,
	}.Build()

	integration.RegisterServiceViaAPI(t, registrationGRPCClient, req)
	t.Logf("INFO: '%s' registered.", serviceID)

	// 3. Call Tool via MCPANY
	testMCPClient := mcp.NewClient(&mcp.Implementation{Name: "test-mcp-client", Version: "v1.0.0"}, nil)
	cs, err := testMCPClient.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: mcpAnyTestServerInfo.HTTPEndpoint}, nil)
	require.NoError(t, err)
	defer cs.Close()

	sanitizedServiceID, _ := util.SanitizeServiceName(serviceID)
	sanitizedToolName, _ := util.SanitizeToolName(operationID)
	toolName := sanitizedServiceID + "." + sanitizedToolName
	toolArgs := `{"year": 2025, "countryCode": "CA"}`

	const maxRetries = 3
	var res *mcp.CallToolResult

	for i := 0; i < maxRetries; i++ {
		res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: toolName, Arguments: json.RawMessage(toolArgs)})
		if err == nil {
			break // Success
		}

		// If the error is a 503 or a timeout, we can retry. Otherwise, fail fast.
		if strings.Contains(err.Error(), "503 Service Temporarily Unavailable") || strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "connection reset by peer") {
			t.Logf("Attempt %d/%d: Call to date.nager.at failed with a transient error: %v. Retrying...", i+1, maxRetries, err)
			time.Sleep(2 * time.Second) // Wait before retrying
			continue
		}

		// For any other error, fail the test immediately.
		require.NoError(t, err, "unrecoverable error calling getPublicHolidays tool")
	}

	require.NoError(t, err, "all retries to date.nager.at failed with transient errors")

	require.NoError(t, err, "Error calling getPublicHolidays tool")
	require.NotNil(t, res, "Nil response from getPublicHolidays tool")

	// 4. Assert Response
	require.Len(t, res.Content, 1, "Expected exactly one content item")
	textContent, ok := res.Content[0].(*mcp.TextContent)
	require.True(t, ok, "Expected text content")

	var result struct {
		Result string `json:"result"`
	}
	err = json.Unmarshal([]byte(textContent.Text), &result)
	require.NoError(t, err, "Failed to unmarshal tool result")

	expectedOutput := "The first public holiday is New Year's Day on 2025-01-01."
	require.Equal(t, expectedOutput, result.Result)

	t.Log("INFO: E2E Test Scenario for Public Holidays API with Transformation for CA 2025 Completed Successfully!")
}
