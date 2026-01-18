# Terraform Provider

**Status:** Experimental / Skeleton

The MCP Any Terraform Provider is currently in the design phase. It is intended to enable "Configuration as Code" for managing MCP resources.

## Resources (Planned)

- `mcp_server`: Manage MCP server instances.

## Example Usage (Proposed)

```hcl
resource "mcp_server" "example" {
  name    = "my-server"
  port    = 8080
  enabled = true
}
```
