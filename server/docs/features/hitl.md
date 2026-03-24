# Human-in-the-Loop (HITL) Middleware

The HITL middleware provides a suspension protocol for user approval flows, preventing catastrophic agent actions by requiring explicit human confirmation for sensitive operations.

## Configuration

The middleware can be configured per-tool or globally:

```yaml
hitl:
  enabled: true
  require_mfa: false
  timeout: 300s
  sensitive_tools:
    - "database.drop_table"
    - "aws.terminate_instance"
```

## How it works

When an agent attempts to execute a tool that matches the configured sensitive tools, the HITL middleware intercepts the request. It initiates an approval flow and suspends the agent's execution until human approval is granted. If the timeout is reached or approval is denied, the execution is blocked.

If `require_mfa` is enabled, the approval process will demand multi-factor attestation for the executable configuration hooks.
