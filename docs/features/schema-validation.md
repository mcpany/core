# Schema Validation

To ensure reliability and prevent misconfiguration, MCP Any validates all configuration files against a strict JSON Schema at startup.

## Features

- **Early Error Detection**: Catches configuration errors before the server starts.
- **Detailed Feedback**: Provides specific error messages pointing to the invalid configuration line.
- **IDE Support**: The schema can be used in IDEs (like VS Code) for autocompletion and validation while editing `config.yaml`.

For more information, see [Schema Validation](../../server/docs/features/schema-validation.md).
