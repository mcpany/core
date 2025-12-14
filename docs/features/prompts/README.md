# Prompts

Prompts allow you to define reusable templates for user input, often used when integrating with LLMs.

## Configuration

Prompts are defined within the `prompts` block of a service.

### Fields

| Field         | Type     | Description                    |
| ------------- | -------- | ------------------------------ |
| `name`        | `string` | unique name of the prompt      |
| `description` | `string` | description of what it does    |
| `messages`    | `list`   | list of messages in the prompt |

### Configuration Snippet

```yaml
upstream_services:
  - name: "prompt-service"
    http_service:
      address: "https://example.com"
      prompts:
        - name: "welcomer"
          description: "Welcomes the user"
          messages:
            - role: "USER"
              text:
                text: "Hello, {{name}}!"
```

## Use Case

You want to provide a standard "Summarize this" prompt to your users. Instead of them typing it out, they can select the "summarize" prompt from the list exposed by your MCP server.

## Public API Example

Clients call `prompts/get` with the prompt name and arguments.
