# Conditional Prompt Example

This example demonstrates how to use conditional logic in prompts.

## Running the Example

1.  **Start the MCP Any server with the example configuration:**

    ```bash
    make run ARGS="--config-paths ./examples/conditional-prompt/config.yaml"
    ```

2.  **Call the `conditional-prompt` with `use_extra_message` set to `true`:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "prompts/get", "params": {"name": "conditional-prompt-service.conditional-prompt", "arguments": {"use_extra_message": true}}, "id": 1}' \
      http://localhost:50050
    ```

    You should see the following output:

    ```json
    {
      "jsonrpc": "2.0",
      "id": 1,
      "result": {
        "description": "A prompt that uses conditional logic",
        "messages": [
          {
            "role": "user",
            "content": {
              "text": "Hello! Here is an extra message."
            }
          }
        ]
      }
    }
    ```

3.  **Call the `conditional-prompt` with `use_extra_message` set to `false`:**

    ```bash
    curl -X POST -H "Content-Type: application/json" \
      -d '{"jsonrpc": "2.0", "method": "prompts/get", "params": {"name": "conditional-prompt-service.conditional-prompt", "arguments": {"use_extra_message": false}}, "id": 1}' \
      http://localhost:50050
    ```

    You should see the following output:

    ```json
    {
      "jsonrpc": "2.0",
      "id": 1,
      "result": {
        "description": "A prompt that uses conditional logic",
        "messages": [
          {
            "role": "user",
            "content": {
              "text": "Hello!"
            }
          }
        ]
      }
    }
    ```
