# Webhooks

**Status:** Implemented

## Goal

Intercept and modify tool executions. Webhooks utilize CloudEvents for a standardized event format and can be used for policy enforcement, data transformation, or auditing.

## Configuration

Webhooks are configured via `config.yaml` under each upstream service as `pre_call_hooks` and `post_call_hooks`. The UI dashboard at `/webhooks` provides an interface for managing global webhook subscriptions and allows you to test webhook delivery manually.

### YAML Example

```yaml
upstream_services:
  - name: "my-service"
    pre_call_hooks:
      - name: "validate-input"
        webhook:
          url: "http://my-webhook-service/validate"
```

## Supported Types

1.  **Pre-Call (`pre_call_hooks`)**: Triggered _before_ the tool executes to validate inputs or enforce policies.
2.  **Post-Call (`post_call_hooks`)**: Triggered _after_ the tool executes to audit results or transform output.

For advanced implementations, including the official webhook sidecar, see the Server Documentation.
