# Recursive Context Protocol

The Recursive Context Protocol standardizes headers for subagent inheritance. It solves configuration pain by allowing context state to be passed along a chain of subagents securely.

## Configuration

It can be configured globally to allow context inheritance for the swarm:

```yaml
recursive_context:
  enabled: true
  max_depth: 10
  ttl: 3600s
```

## How it works

When a subagent spawns, the protocol attaches an `X-MCP-Parent-Context-ID` header. The middleware automatically pulls the shared session state and makes it available to the new subagent, preventing "context loss" or the need to re-authenticate or re-initialize basic state variables.
