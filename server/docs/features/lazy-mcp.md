# Lazy-MCP Middleware (On-Demand Discovery)

The Lazy-MCP middleware provides an on-demand tool discovery mechanism, preventing context pollution by dynamically fetching and registering tools based on similarity and relevance.

## Configuration

The middleware dynamically analyzes the agent's intent and loads only the necessary MCP tools into the active context.

```yaml
lazy_mcp:
  enabled: true
  threshold: 0.85
  cache_ttl: 600s
```

## How it works

When the agent attempts to discover tools, Lazy-MCP evaluates the request against a pre-indexed registry of available tools using a similarity-based approach. It responds with a subset of tools, reducing the tokens consumed by large schemas and improving overall agent focus.
