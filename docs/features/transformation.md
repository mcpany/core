# Structured Output Transformation

MCP Any provides powerful capabilities to transform the output of upstream services before sending it back to the client. This is useful for reshaping complex JSON responses, filtering sensitive data, or converting formats.

## Features

- **JQ Support**: Use the powerful `jq` query language to filter and reshape JSON data.
- **JSONPath**: Support for standard JSONPath expressions for simpler extractions.
- **Templates**: Combine multiple fields into a single text output using Go templates.

## Configuration

Transformations are configured as part of the tool definition or service configuration.

```yaml
services:
  - id: "weather-api"
    # ...
    tools:
      - name: "get_current_weather"
        # ...
        transform:
          jq: ".main | {temp: .temp, humidity: .humidity}" # Extract only specific fields
```

## Why use Transformations?

1.  **Token Efficiency**: Reduce the amount of data sent to the LLM, saving context window and cost.
2.  **Privacy**: Remove fields that the LLM shouldn't see (internal IDs, PII not caught by DLP).
3.  **Usability**: Present data in a format that is easier for the LLM to understand (e.g., flattening nested structures).
