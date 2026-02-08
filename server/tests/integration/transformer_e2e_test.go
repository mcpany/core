// Copyright 2025 Author(s) of MCP Any
// SPDX-License-Identifier: Apache-2.0

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"text/template"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mcpanyConfigTemplate = `
upstream_services:
  - name: "mock-http-service"
    auto_discover_tool: true
    http_service:
      address: "{{ .ServerURL }}"
      calls:
        get_json:
          endpoint_path: "/json"
          method: "HTTP_METHOD_GET"
          output_transformer:
            format: "JSON"
            extraction_rules:
              name: "{.person.name}"
              age: "{.person.age}"
        get_xml:
          endpoint_path: "/xml"
          method: "HTTP_METHOD_GET"
          output_transformer:
            format: "XML"
            extraction_rules:
              name: "//name"
              value: "//value"
        get_text:
          endpoint_path: "/text"
          method: "HTTP_METHOD_GET"
          output_transformer:
            format: "TEXT"
            extraction_rules:
              name: "Name: ([\\w-]+)"
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



func TestTransformerE2E_Extraction(t *testing.T) {
	// Cannot use t.Parallel() with t.Setenv
	// t.Parallel()

	// Allow local IPs for testing
	t.Setenv("MCPANY_DANGEROUS_ALLOW_LOCAL_IPS", "true")

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

	client, cleanup := StartStdioServer(t, tmpFile.Name())
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// List tools to ensure they are discovered
	var listRes *mcp.ListToolsResult
	assert.Eventually(t, func() bool {
		listRes, err = client.ListTools(ctx)
		// Expect 3 discovered tools + 1 built-in roots tool = 4
		return err == nil && len(listRes.Tools) >= 3
	}, 30*time.Second, 100*time.Millisecond, "Expected at least 3 tools to be discovered")
	require.GreaterOrEqual(t, len(listRes.Tools), 3)

	// Helper to call tool and get result map
	callAndGetMap := func(toolName string) map[string]interface{} {
		// We use raw Call because we want to inspect content flexibly
		res, err := client.Call(ctx, "tools/call", map[string]interface{}{
			"name": toolName,
		})
		require.NoError(t, err)

		resMap, ok := res.(map[string]interface{})
		require.True(t, ok, "Result should be a map")

		content := resMap["content"]
		// Check if content is list (spec) or map (possible violation/extension)
		if contentList, ok := content.([]interface{}); ok {
			require.NotEmpty(t, contentList, "Content list empty")
			// Assume first item is text and contains JSON?
			// The transformer should return the extracted DATA.
			// If it returns {"name": ...}, mcpany might wrap it in TextContent?
			// Or maybe it returns it as embedded resource?
			// Let's log it if unsure.
			item := contentList[0].(map[string]interface{})
			if text, ok := item["text"].(string); ok {
				// It's a text content, maybe JSON string?
				// But transformer "JSON" output format usually implies STRUCTURED result?
				// mcpany doesn't support structured content unless EmbeddedResource?
				// If response transformer returns object, mcpany might serialize it to JSON string in TextContent.
				var extracted map[string]interface{}
				// Try parsing as JSON
				if err := json.Unmarshal([]byte(text), &extracted); err == nil {
					return extracted
				}
				// If not JSON, maybe text IS the value? (for text/plain)
				return map[string]interface{}{"text": text}
			}
			return item
		} else if contentMap, ok := content.(map[string]interface{}); ok {
			// Direct map return (non-spec but possible if mcpany transforms it so)
			return contentMap
		}
		require.Fail(t, "Unknown content format", "Content: %v", content)
		return nil
	}

	// Test JSON extraction
	jsonResult := callAndGetMap("mock-http-service.get_json")
	// If it was TextContent with JSON string, we parsed it.
	// We expect "name" and "age".
	// If parsing failed (e.g. invalid JSON), we assert on values.
	// But keys should exist.
	if _, ok := jsonResult["name"]; !ok {
		// Maybe it's inside "text" key if we returned map{"text":...}
		// mcpany response transformer output is usually the TOOL RESULT.
		// If tool result is {"name":"test-json"}, and mcpany wraps it...
		// It becomes content=[{type:"text", text:"{\"name\":...}"}]
		t.Logf("JSON Result: %v", jsonResult)
	}
	assert.Equal(t, "test-json", jsonResult["name"])
	assert.Equal(t, float64(123), jsonResult["age"])

	// Test XML extraction
	xmlResult := callAndGetMap("mock-http-service.get_xml")
	assert.Equal(t, "test-xml", xmlResult["name"])
	assert.Equal(t, "456", xmlResult["value"])

	// Test Text extraction
	textResult := callAndGetMap("mock-http-service.get_text")
	assert.Equal(t, "test-text", textResult["name"])
	assert.Equal(t, "789", textResult["value"])
}
