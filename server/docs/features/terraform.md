# Terraform Provider (Proposal)

**Status:** Proposal / Not Implemented

The MCP Any Terraform Provider is a proposed feature to enable "Configuration as Code" for managing MCP resources using HashiCorp Terraform.

## Proposed Resources

- `mcp_server`: Manage MCP server instances.

## Example Usage (Conceptual)

```hcl
resource "mcp_server" "example" {
  name    = "my-server"
  port    = 8080
  enabled = true
}
```

## Note

This feature is currently in the design phase and is not yet available in the `mcpany` binary.
