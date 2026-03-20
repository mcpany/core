package main

import (
	"io/ioutil"
	"strings"
)

func main() {
	content, err := ioutil.ReadFile("server/pkg/terraform/resource_mcp_server.go")
	if err != nil {
		panic(err)
	}

	str := string(content)

	// Fix Create comments
	oldCreateComment := `// Create mimics the Create operation of a Terraform resource. _ is an unused parameter. Returns an error if the operation fails.
//
// Parameters:
//   - _ (*ResourceMCPServer): The _ parameter.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.`
	newCreateComment := `// Create mimics the Create operation of a Terraform resource. Returns an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - serverURL (string): The URL of the MCP server API.
//   - resource (*ResourceMCPServer): The resource to create.
//
// Returns:
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the HTTP request fails or the status is not created/ok.`
	str = strings.ReplaceAll(str, oldCreateComment, newCreateComment)

	// Fix Read comments
	oldReadComment := `// Read mimics the Read operation. name is the name of the resource. Returns the result. Returns an error if the operation fails.
//
// Parameters:
//   - name (string): The name parameter.
//
// Returns:
//   - *ResourceMCPServer: The resulting *ResourceMCPServer.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the operation fails or is invalid.`
	newReadComment := `// Read mimics the Read operation. Returns the result or an error if the operation fails.
//
// Parameters:
//   - ctx (context.Context): The context for the request.
//   - serverURL (string): The URL of the MCP server API.
//   - name (string): The name of the resource.
//
// Returns:
//   - *ResourceMCPServer: The resulting *ResourceMCPServer.
//   - error: An error if the operation fails.
//
// Errors:
//   - Returns an error if the HTTP request fails or the status is not ok.`
	str = strings.ReplaceAll(str, oldReadComment, newReadComment)

	ioutil.WriteFile("server/pkg/terraform/resource_mcp_server.go", []byte(str), 0644)
}
