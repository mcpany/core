# Audit Logging

MCP Any provides a built-in audit logging capability to record details about every tool execution. This is essential for compliance, security monitoring, and debugging.

## Overview

When enabled, the audit logger intercepts every tool execution request and records structured data about the event to a specified file or database.

The audit log captures:
- **Timestamp**: When the execution started.
- **Tool Name**: The name of the tool being executed.
- **User ID**: The ID of the authenticated user (if available).
- **Profile ID**: The ID of the active profile (if available).
- **Duration**: How long the execution took.
- **Arguments**: The input arguments provided to the tool (optional).
- **Result**: The result returned by the tool (optional).
- **Error**: Any error that occurred during execution.
- **Hash Integrity**: SHA-256 hash chaining to detect tampering.

## Configuration

Audit logging is configured in the `GlobalSettings` section of your MCP Any configuration file.

### Example Configuration

```yaml
global_settings:
  audit:
    enabled: true
    output_path: "audit.log"
    log_arguments: true
    log_results: false
    storage_type: STORAGE_TYPE_FILE # or STORAGE_TYPE_SQLITE
```

### Configuration Options

| Option | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `enabled` | `bool` | `false` | Enables or disables audit logging. |
| `output_path` | `string` | `""` | Path to the log file or database. If empty, logs are written to stdout (File mode). |
| `log_arguments` | `bool` | `false` | If true, logs the input arguments. **Warning:** May log sensitive data. |
| `log_results` | `bool` | `false` | If true, logs the execution result. **Warning:** May log sensitive data. |
| `storage_type` | `enum` | `FILE` | The storage backend to use: `STORAGE_TYPE_FILE` or `STORAGE_TYPE_SQLITE`. |

## Log Format (File / JSON)

Audit logs are written as newline-delimited JSON (NDJSON). Each line represents a single tool execution event.

### Example Log Entry

```json
{
  "timestamp": "2023-10-27T10:00:00.123Z",
  "tool_name": "weather_get_forecast",
  "user_id": "alice",
  "profile_id": "prod",
  "duration": "150ms",
  "duration_ms": 150,
  "arguments": {
    "city": "London"
  },
  "result": {
    "temperature": 15,
    "unit": "C"
  },
  "prev_hash": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
  "hash": "88a7c9ce3aee64920d1d58892f9db28a7b76bf0a4785c31fc08176c28e080362"
}
```

## Tamper-Evidence

Audit logs are secured using a SHA-256 hash chain. Each log entry contains:
1.  **`prev_hash`**: The hash of the previous log entry (or empty string for the first entry).
2.  **`hash`**: The SHA-256 hash of the current entry's content combined with `prev_hash`.

This ensures that any modification to the log file (deletion, insertion, or modification of lines) will invalidate the hash chain of all subsequent entries, allowing auditors to detect tampering.

## Security Considerations

- **Sensitive Data**: By default, `log_arguments` and `log_results` are disabled. Enable them with caution, as they may expose API keys, PII, or other sensitive information handled by your tools.
- **File Permissions**: Ensure that the `output_path` is writable by the MCP Any server process and readable only by authorized personnel.
