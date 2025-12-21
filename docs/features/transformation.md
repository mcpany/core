# Structured Output Transformation

MCP Any provides powerful capabilities to transform the output of upstream services before sending it back to the client. This allows you to normalize data formats, extract specific fields, or reshape complex JSON/XML responses into something more consumable by LLMs.

## Supported Transformation Engines

*   **Go Templates**: Use standard Go text templates to format output.
*   **JSONPath**: Extract data from JSON responses using JSONPath syntax.
*   **JQ**: Use the powerful `jq` query language for advanced JSON processing.
*   **XPath**: Extract data from XML responses.
*   **Regex**: Extract data from plain text using regular expressions.

## Configuration

Transformations are defined in the tool configuration.

### Example: JQ Transformation

```yaml
tools:
  - name: get_user_summary
    http:
      url: https://api.example.com/users/{id}
    transform:
      type: jq
      query: "{ name: .name, email: .email, role: .roles[0] }"
```

### Example: Go Template

```yaml
tools:
  - name: weather_report
    http:
      url: https://api.weather.com/current
    transform:
      type: template
      template: "The weather in {{.location.name}} is {{.current.temp_c}}°C."
```

## Parsing

The `TextParser` component supports parsing raw text/JSON/XML inputs into structured maps, which can then be further transformed.
