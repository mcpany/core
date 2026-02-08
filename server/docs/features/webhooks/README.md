# Webhooks

Webhooks in MCP Any allow you to intercept and modify tool executions. They utilize [CloudEvents](https://cloudevents.io/) for a standardized event format.

## Overview

You can configure webhooks to trigger at two stages of a tool call:

1.  **Pre-Call (`pre_call_hooks`)**: Triggered _before_ the tool executes.
    - **Validation**: specific arguments can be checked and the call denied if invalid.
    - **Policy Enforcement**: Prevent dangerous actions (like `rm -rf`).
    - **Input Transformation**: Modify arguments before they reach the tool.
2.  **Post-Call (`post_call_hooks`)**: Triggered _after_ the tool executes.
    - **Auditing**: Log the result of the tool.
    - **Output Transformation**: Convert the output format (e.g., HTML to Markdown).
    - **Sanitization**: Redact sensitive information from the result.

## Configuration

Webhooks are configured in `config.yaml` under each upstream service.

```yaml
upstream_services:
  - name: "my-service"
    # ... service config ...
    pre_call_hooks:
      - name: "validate-input"
        webhook:
          url: "http://my-webhook-service/validate"
    post_call_hooks:
      - name: "audit"
        webhook:
          url: "http://my-webhook-service/audit"
```

## Examples

We provide ready-to-run examples to demonstrate the power of webhooks.

### 1. Blocking Dangerous Commands (`block_rm`)

This example demonstrates a **Pre-Call Hook** that inspects arguments for a command-line tool and blocks execution if it contains the `rm` command.

- **[View Example Code](./examples/block_rm)**
- **Scenario**: A `busybox` container service that allows executing commands. We want to prevent users from running `rm`.
- **Mechanism**: The webhook receives the tool inputs, checks for "rm", and returns `allowed: false` if found.

#### Running the Example

1.  Navigate to the example directory:
    ```bash
    cd server/docs/features/webhooks/examples/block_rm
    ```
2.  Start the Webhook Server:

    ```bash
    # Runs on localhost:8081
    # Ensure dependencies are tidy
    go mod tidy
    go run main.go
    ```

    > **Note**: If you prefer to build the binary, we recommend outputting it to a temporary location to avoid committing it: `go build -o /tmp/block_rm_server main.go && /tmp/block_rm_server`

3.  In a separate terminal, start MCP Any with the example config:
    ```bash
    # Assuming you are in the root of the repo
    go run cmd/server/main.go run --config-path server/docs/features/webhooks/examples/block_rm/config.yaml
    ```
4.  Test functionality using `gemini` CLI (or any MCP client):

    ```bash
    # Should SUCCEED
    gemini -m gemini-2.5-flash -p "Run 'ls -la' in the busybox container"

    # Should FAIL (Blocked by Webhook)
    gemini -m gemini-2.5-flash -p "Run 'rm myfile' in the busybox container"
    ```

#### Verification (E2E Test)

You can also run the included End-to-End test to verify the behavior programmatically:

```bash
# In docs/features/webhooks/examples/block_rm
go test -v e2e_test.go
```

### 2. HTML to Markdown Conversion (`html_to_md`)

This example demonstrates a **Post-Call Hook** that transforms the output of a tool.

- **[View Example Code](./examples/html_to_md)**
- **Scenario**: We fetch a webpage which returns raw HTML. We want the LLM to receive clean Markdown.
- **Mechanism**: The webhook receives the tool result (HTML), converts it to Markdown, and returns the replacement object.

#### Running the Example

1.  Navigate to the example directory:
    ```bash
    cd server/docs/features/webhooks/examples/html_to_md
    ```
2.  Start the Webhook Server:

    ```bash
    # Runs on localhost:8082
    go run main.go
    ```

    > **Note**: If you build the binary, verify it outside the source tree: `go build -o /tmp/html_to_md_server main.go`

3.  Start MCP Any with the example config:
    ```bash
    go run cmd/server/main.go run --config-path server/docs/features/webhooks/examples/html_to_md/config.yaml
    ```

#### Verification (E2E Test)

Run the included E2E test to verify implementation:

```bash
# In docs/features/webhooks/examples/html_to_md
go test -v e2e_test.go
```
