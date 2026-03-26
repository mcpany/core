# Webhooks

<<<<<<< HEAD
**Status:** Implemented
=======
**Status:** Config-Driven (UI Planned)
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

## Goal

Intercept and modify tool executions. Webhooks utilize CloudEvents for a standardized event format and can be used for policy enforcement, data transformation, or auditing.

## Configuration

<<<<<<< HEAD
Webhooks can be configured both via the UI dashboard and `config.yaml` under each upstream service as `pre_call_hooks` and `post_call_hooks`.

### UI Dashboard

Navigate to `/webhooks` to access the Webhooks management console. Here you can:
- **Add Webhooks**: Create new webhooks by providing a Payload URL.
- **Status Management**: Instantly enable or disable specific webhooks via toggle switches.
- **Testing**: Trigger test events (`/api/v1/webhooks/:id/test`) to verify connectivity and validate payloads.
- **Delete Webhooks**: Remove unused or obsolete webhooks from your configuration.
=======
Webhooks are currently configured via `config.yaml` under each upstream service as `pre_call_hooks` and `post_call_hooks`. The UI dashboard at `/webhooks` provides a preview of future management capabilities but is not yet fully functional for creating or modifying these hooks.
>>>>>>> 4f039895e (⚡ Bolt: Render Optimization for System Status Banner (#6544))

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
