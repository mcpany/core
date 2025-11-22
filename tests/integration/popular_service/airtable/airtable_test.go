package airtable

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mcpany/core/pkg/test/e2e"
	"github.com/stretchr/testify/require"
)

func TestAirtable(t *testing.T) {
	// Create a mock Airtable server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"records": []}`))
	}))
	defer mockServer.Close()

	// Set the Airtable API key and base ID/table ID for the test
	t.Setenv("AIRTABLE_API_KEY", "test-api-key")
	t.Setenv("AIRTABLE_BASE_ID", "test-base-id")
	t.Setenv("AIRTABLE_TABLE_ID", "test-table-id")

	// Create a new e2e test
	test := e2e.New(t, e2e.BuildOpts{
		ConfigPaths: []string{
			"../../../../examples/popular_services/airtable/config.yaml",
		},
		ModifyUpstreamServices: map[string]e2e.UpstreamServiceModifier{
			"airtable": {
				HTTPAddress: mockServer.URL,
			},
		},
	})
	defer test.Close()

	// List the tools and check if the Airtable tool is present
	tools, err := test.Client.ListTools(context.Background())
	require.NoError(t, err)
	require.Contains(t, tools, "airtable/-/list_records")

	// Call the Airtable tool and check the result
	result, err := test.Client.CallTool(context.Background(), "airtable/-/list_records", map[string]interface{}{
		"baseId":        "test-base-id",
		"tableIdOrName": "test-table-id",
	})
	require.NoError(t, err)
	require.JSONEq(t, `{"records": []}`, result)
}
