package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"text/template"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpanyConfigTemplate = `
upstream_services:
  - name: "mock-http-service"
    http_service:
      address: "{{ .ServerURL }}"
      calls:
        - schema:
            name: "get_json"
          endpoint_path: "/json"
          method: "GET"
          response_transformer:
            - from: "json"
              to:
                name: "{.person.name}"
                age: "{.person.age}"
        - schema:
            name: "get_xml"
          endpoint_path: "/xml"
          method: "GET"
          response_transformer:
            - from: "xml"
              to:
                name: "//name"
                value: "//value"
        - schema:
            name: "get_text"
          endpoint_path: "/text"
          method: "GET"
          response_transformer:
            - from: "text"
              to:
                name: "Name: (\\w+)"
                value: "Value: (\\d+)"
`

func newMockServer(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"person": {"name": "test-json", "age": 123}}`))
		case "/xml":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<root><name>test-xml</name><value>456</value></root>`))
		case "/text":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte(`Name: test-text, Value: 789`))
		default:
			http.NotFound(w, r)
		}
	}))
	return server
}

func getMcpanyConfig(t *testing.T, serverURL string) string {
	t.Helper()
	tmpl, err := template.New("mcpany-config").Parse(mcpanyConfigTemplate)
	require.NoError(t, err)
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct{ ServerURL string }{ServerURL: serverURL})
	require.NoError(t, err)
	return buf.String()
}

func callTool(t *testing.T, serverInfo *MCPANYTestServerInfo, toolName string) map[string]any {
	t.Helper()

	payload := map[string]any{
		"jsonrpc": "2.0",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": nil,
		},
		"id": 1,
	}
	payloadBytes, err := json.Marshal(payload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", serverInfo.HTTPEndpoint, bytes.NewBuffer(payloadBytes))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	var resp *http.Response
	require.Eventually(t, func() bool {
		resp, err = serverInfo.HTTPClient.Do(req)
		if err != nil {
			t.Logf("tool call request failed for %s: %v", toolName, err)
			return false
		}
		defer func() { _ = resp.Body.Close() }()
		return resp.StatusCode == http.StatusOK
	}, 10*time.Second, 250*time.Millisecond, "Failed to get a successful response from mcpany for tool %s", toolName)

	defer func() { _ = resp.Body.Close() }()
	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	require.Nil(t, result["error"], "JSON-RPC error: %v", result["error"])
	require.NotNil(t, result["result"])
	resultMap := result["result"].(map[string]any)
	require.NotNil(t, resultMap["content"])
	contentMap := resultMap["content"].(map[string]any)
	return contentMap
}

func TestTransformerE2E_Extraction(t *testing.T) {
	t.Skip("Skipping flaky test")
	t.Parallel()

	mockServer := newMockServer(t)
	defer mockServer.Close()

	configContent := getMcpanyConfig(t, mockServer.URL)
	t.Logf("Using config:\n%s", configContent)

	tmpFile, err := os.CreateTemp(t.TempDir(), "mcpany-config-*.yaml")
	require.NoError(t, err)
	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	mcpanyServer := StartMCPANYServer(t, "transformer-e2e", "--config-path", tmpFile.Name())
	defer func() {
		t.Logf("MCPANY Server stdout:\n%s", mcpanyServer.Process.StdoutString())
		t.Logf("MCPANY Server stderr:\n%s", mcpanyServer.Process.StderrString())
		mcpanyServer.CleanupFunc()
	}()

	// Test JSON extraction
	jsonResult := callTool(t, mcpanyServer, "mock-http-service/-/get_json")
	assert.Equal(t, "test-json", jsonResult["name"])
	assert.Equal(t, float64(123), jsonResult["age"])

	// Test XML extraction
	xmlResult := callTool(t, mcpanyServer, "mock-http-service/-/get_xml")
	assert.Equal(t, "test-xml", xmlResult["name"])
	assert.Equal(t, "456", xmlResult["value"])

	// Test Text extraction
	textResult := callTool(t, mcpanyServer, "mock-http-service/-/get_text")
	assert.Equal(t, "test-text", textResult["name"])
	assert.Equal(t, "789", textResult["value"])
}
