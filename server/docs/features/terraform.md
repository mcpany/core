# Terraform Provider

**Status:** Implemented

The MCP Any Terraform Provider enables "Configuration as Code" for managing MCP resources using HashiCorp Terraform.

## Proposed Resources

- `mcp_server`: Manage MCP server instances.

## Example Usage

```hcl
resource "mcp_server" "example" {
  name    = "my-server"
  port    = 8080
  enabled = true
}
```

