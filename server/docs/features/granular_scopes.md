# Granular Scopes

Granular Scopes implement a capability-based token system, enabling "Least Privilege" security for agents and guarding against data exfiltration risks.

## Configuration

It applies token-based scoping directly to resources.

```yaml
scopes:
  default: "read"
  tokens:
    - "fs:read:/tmp"
    - "db:write:users"
```

## How it works

When an agent issues a request, the server inspects the capability tokens bound to the agent's identity. If an agent with only `fs:read:/tmp` attempts to write a file or access `/etc/passwd`, the request is blocked. This Zero-Trust subagent scoping strictly binds capabilities to intents.
