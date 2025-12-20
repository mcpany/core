# JQ Output Transformation

MCP Any supports transforming upstream API responses using [JQ](https://jqlang.github.io/jq/) queries. This allows you to restructure, filter, and map complex JSON responses into a cleaner format for the LLM.

## Configuration

To use JQ transformation, set the `output_transformer` configuration in your `HttpCallDefinition` (or other call definitions).

Set `format` to `JQ` and provide a `jq_query`.

### Example

Suppose an upstream API returns the following JSON:

```json
{
  "users": [
    {"id": 1, "name": "Alice", "active": true},
    {"id": 2, "name": "Bob", "active": false},
    {"id": 3, "name": "Charlie", "active": true}
  ]
}
```

You want to extract only the names of active users. You can use the following configuration:

```yaml
calls:
  get_active_users:
    endpoint_path: "/users"
    method: "GET"
    output_transformer:
      format: "JQ"
      jq_query: ".users[] | select(.active) | .name"
```

The tool will return:

```json
["Alice", "Charlie"]
```

### Using with Templates

You can combine JQ transformation with a Go template. If you do this, the JQ query **must** return a JSON object (map), so that the template can access fields by name.

```yaml
calls:
  get_user_summary:
    endpoint_path: "/users"
    method: "GET"
    output_transformer:
      format: "JQ"
      # Construct an object with a 'names' field
      jq_query: "{names: [.users[] | select(.active) | .name]}"
      template: "Active users are: {{join \", \" .names}}"
```

Result:
```text
Active users are: Alice, Charlie
```

## Performance

The JQ queries are compiled and cached for performance.
