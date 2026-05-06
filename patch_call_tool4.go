package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	path := "server/pkg/mcpserver/server.go"
	content, _ := os.ReadFile(path)

	oldCode := `	// 3. Fallback: If no structured result identified, treat as raw data
	if finalResult == nil {
		if len(jsonBytes) == 0 && marshalErr == nil {
			text, marshalErr = util.FastMarshalToString(result)
		} else if marshalErr == nil {
			text = util.BytesToString(jsonBytes)
		}

		if marshalErr != nil {
			text = util.ToString(result)
		}

		finalResult = &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}
	}`

	newCode := `	// 3. Fallback: If no structured result identified, treat as raw data
	if finalResult == nil {
		if len(jsonBytes) == 0 && marshalErr == nil {
			// ⚡ Bolt Optimization: Use FastMarshalToString instead of byte allocation if bytes aren't needed yet
			text, marshalErr = util.FastMarshalToString(result)
		} else if marshalErr == nil {
			text = util.BytesToString(jsonBytes)
		}

		if marshalErr != nil {
			text = util.ToString(result)
		}

		finalResult = &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: text},
			},
		}
	}`

	if !strings.Contains(string(content), oldCode) {
		fmt.Println("Could not find old code in server.go")
		os.Exit(1)
	}

	newContent := strings.Replace(string(content), oldCode, newCode, 1)
	os.WriteFile(path, []byte(newContent), 0644)
	fmt.Println("Patched server.go fallback")
}
