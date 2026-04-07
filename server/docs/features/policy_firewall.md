# Policy Firewall

The Policy Firewall Engine provides critical "Zero Trust" agent execution security by enforcing strict rules on tool calls.

## Configuration

The Policy Firewall is configured via `PolicyFirewallConfig` to define explicitly allowed and blocked tools.

```yaml
policy_firewall:
  enabled: true
  default_action: "deny"
  allowed_tools:
    - "fs.read.*"
    - "github.search"
  blocked_tools:
    - "aws.delete_bucket"
    - "db.drop_table"
```

## How it works

When an agent attempts to execute a tool, the Policy Firewall evaluates the request:

1. **Blocklist**: The firewall first checks if the tool matches any entries in the `blocked_tools` list. If a match is found, the request is immediately denied.
2. **Allowlist**: If the tool is not blocked, it checks the `allowed_tools` list. If a match is found, the execution is permitted.
3. **Default Action**: If the tool is neither explicitly blocked nor allowed, the system falls back to the `default_action` (either `allow` or `deny`, defaulting to `deny` for secure-by-default execution).

Wildcards (`.*`) are supported for prefix matching, allowing administrators to easily block entire modules (e.g. `aws.*`) or selectively allow safe modules (e.g. `fs.read.*`).
