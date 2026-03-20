# Schema Validation

MCP Any includes built-in schema validation to ensure that your configuration files (`config.yaml` or `config.json`) are syntactically correct and adhere to the expected structure before the server starts.

## Overview

When you start the `mcpany` server, it parses your configuration file against the internal Protobuf definitions. This validation process checks for:

-   **Required Fields**: Ensuring all mandatory fields are present.
-   **Type Safety**: Verifying that values are of the correct type (e.g., integers, booleans, strings).
-   **Structure**: confirming that nested objects and lists are correctly formatted.
-   **Enums**: Ensuring that enum fields (like `log_level`) use valid values.

## How it Works

The validation happens automatically at startup. If the configuration file is invalid, the server will:

1.  Log a detailed error message indicating the specific field and nature of the error.
2.  Exit immediately with a non-zero status code.

This prevents the server from running with a broken or ambiguous configuration, which could lead to runtime errors or security vulnerabilities.

## Example Error

If you try to set a string value for a field that expects an integer, you might see an error like:

```text
Error loading config: invalid value for field 'max_connections': expected integer, got string "ten"
```

## Best Practices

-   **Validate in CI/CD**: Run a "dry run" or just start the server with the config in your CI pipeline to catch configuration errors before deploying to production.
-   **Use VS Code Extensions**: If you are using YAML or JSON, use an editor that supports schema validation (e.g., via JSON Schema) to get intellisense and error highlighting as you edit.
