# Prompt Injection Guardrails

The Guardrails middleware provides security by intercepting and blocking malicious prompts before they reach the backend tools.

## Configuration

To enable guardrails, add the `guardrails` configuration block to `global_settings`:

```yaml
global_settings:
  guardrails:
    blocked_phrases:
      - "ignore all previous instructions"
      - "system prompt"
```

## How it works

The middleware scans the body of POST requests. If a blocked phrase is detected (case-insensitive), the request is aborted with a `400 Bad Request` error and a policy violation message.
